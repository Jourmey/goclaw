package cmd

import (
	"context"
	"github.com/nextlevelbuilder/goclaw/internal/hooks"
	"log/slog"
	"path/filepath"

	"github.com/google/uuid"

	"strings"

	"github.com/nextlevelbuilder/goclaw/internal/agent"
	"github.com/nextlevelbuilder/goclaw/internal/bus"
	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/eventbus"
	httpapi "github.com/nextlevelbuilder/goclaw/internal/http"
	mcpbridge "github.com/nextlevelbuilder/goclaw/internal/mcp"
	"github.com/nextlevelbuilder/goclaw/internal/media"
	memorypkg "github.com/nextlevelbuilder/goclaw/internal/memory"
	"github.com/nextlevelbuilder/goclaw/internal/providers"
	"github.com/nextlevelbuilder/goclaw/internal/sandbox"
	"github.com/nextlevelbuilder/goclaw/internal/skills"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/store/pg"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
	"github.com/nextlevelbuilder/goclaw/internal/tracing"
	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
)

// wireExtras wires components that require PG stores:
// agent resolver (lazy-creates Loops from DB), virtual FS interceptors, memory tools,
// and cache invalidation event subscribers.
// PG store creation and tracing are handled in gateway.go before this is called.
// Returns the ContextFileInterceptor so callers can pass it to AgentsMethods
// for immediate cache invalidation on agents.files.set.
func wireExtras(
	stores *store.Stores,
	agentRouter *agent.Router,
	providerReg *providers.Registry,
	modelReg providers.ModelRegistry,
	msgBus *bus.MessageBus,
	sessStore store.SessionStore,
	toolsReg *tools.Registry,
	toolPE *tools.PolicyEngine,
	skillsLoader *skills.Loader,
	hasMemory bool,
	traceCollector *tracing.Collector,
	workspace string,
	injectionAction string,
	appCfg *config.Config,
	sandboxMgr sandbox.Manager,
	redisClient any, // nil when built without -tags redis or when Redis is unconfigured
	domainBus eventbus.DomainEventBus,
) (*tools.ContextFileInterceptor, *mcpbridge.Pool, *media.Store, tools.PostTurnProcessor) {
	// 1. Build cache instances (in-memory or Redis depending on build tags)
	agentCtxCache, userCtxCache := makeCaches(redisClient)

	// 1a. Context file interceptor (created before resolver so callbacks can reference it)
	var contextFileInterceptor *tools.ContextFileInterceptor
	if stores.Agents != nil {
		contextFileInterceptor = tools.NewContextFileInterceptor(stores.Agents, workspace, agentCtxCache, userCtxCache)
	}

	// 1c. Persistent media storage for cross-turn image/document access
	mediaStore, err := media.NewStore(filepath.Join(workspace, ".media"))
	if err != nil {
		slog.Warn("media store creation failed, images will not persist across turns", "error", err)
	}

	// Wire media cleanup on session delete.
	if mediaStore != nil {
		if pgSess, ok := sessStore.(*pg.PGSessionStore); ok {
			pgSess.OnDelete = func(sessionKey string) {
				_ = mediaStore.DeleteSession(sessionKey)
			}
		}
		// Register media analysis tools (need mediaStore for file access).
		toolsReg.Register(tools.NewReadDocumentTool(providerReg, mediaStore))
		toolsReg.Register(tools.NewReadAudioTool(providerReg, mediaStore))
		toolsReg.Register(tools.NewReadVideoTool(providerReg, mediaStore))
		slog.Info("media tools registered", "tools", "read_document,read_audio,read_video,create_video")
	}

	// 1e. Wire secure CLI store into exec tool for credentialed exec
	if stores.SecureCLI != nil {
		if execTool, ok := toolsReg.Get("exec"); ok {
			if et, ok := execTool.(*tools.ExecTool); ok {
				et.SetSecureCLIStore(stores.SecureCLI)
			}
		}
	}

	// 2. Per-user profile + context file seeding callbacks
	var ensureUserProfile agent.EnsureUserProfileFunc
	var seedUserFiles agent.SeedUserFilesFunc
	if stores.Agents != nil {
		ensureUserProfile = buildEnsureUserProfile(stores.Agents)
		seedUserFiles = buildSeedUserFiles(stores.Agents)
	}

	// 3. Context file loader callback: loads per-user context files dynamically
	var contextFileLoader agent.ContextFileLoaderFunc
	if contextFileInterceptor != nil {
		contextFileLoader = buildContextFileLoader(contextFileInterceptor)
	}

	// 4. Compute global sandbox defaults for resolver
	sandboxEnabled := sandboxMgr != nil
	sandboxContainerDir := ""
	sandboxWorkspaceAccess := ""

	// 5. Shared MCP connection pool (eliminates duplicate connections across agents)
	var mcpPool *mcpbridge.Pool
	if stores.MCP != nil {
		mcpPool = mcpbridge.NewPool(mcpbridge.DefaultPoolConfig())
	}

	// 6. Set up agent resolver: lazy-creates Loops from DB
	var skillAccessStore store.SkillAccessStore
	if sas, ok := stores.Skills.(store.SkillAccessStore); ok {
		skillAccessStore = sas
	}

	// V3 auto-inject: create AutoInjector if episodic store is available.
	var autoInjector memorypkg.AutoInjector

	// vaultIntc is set later by wireVault but captured by closure in OnTextUploaded.
	var vaultIntc *tools.VaultInterceptor

	// Agent Hooks (Issue #875) — lifecycle dispatcher + handlers.
	var hookDispatcher hooks.Dispatcher

	resolver := agent.NewManagedResolver(agent.ResolverDeps{
		AgentStore:             stores.Agents,
		ProviderStore:          stores.Providers,
		ProviderReg:            providerReg,
		ModelRegistry:          modelReg,
		Bus:                    msgBus,
		Sessions:               sessStore,
		Tools:                  toolsReg,
		ToolPolicy:             toolPE,
		Skills:                 skillsLoader,
		SkillAccessStore:       skillAccessStore,
		HasMemory:              hasMemory,
		TraceCollector:         traceCollector,
		EnsureUserProfile:      ensureUserProfile,
		SeedUserFiles:          seedUserFiles,
		ContextFileLoader:      contextFileLoader,
		BootstrapCleanup:       buildBootstrapCleanup(stores.Agents),
		CacheInvalidate:        buildCacheInvalidate(contextFileInterceptor),
		DefaultTimezone:        appCfg.Cron.DefaultTimezone,
		InjectionAction:        injectionAction,
		MaxMessageChars:        appCfg.Gateway.MaxMessageChars,
		CompactionCfg:          appCfg.Agents.Defaults.Compaction,
		ContextPruningCfg:      appCfg.Agents.Defaults.ContextPruning,
		SandboxEnabled:         sandboxEnabled,
		SandboxContainerDir:    sandboxContainerDir,
		SandboxWorkspaceAccess: sandboxWorkspaceAccess,
		AgentLinkStore:         stores.AgentLinks,
		TeamStore:              stores.Teams,
		DataDir:                workspace,
		SecureCLIStore:         stores.SecureCLI,
		BuiltinToolStore:       stores.BuiltinTools,
		ConfigPermStore:        stores.ConfigPermissions,
		MediaStore:             mediaStore,
		ModelPricing:           appCfg.Telemetry.ModelPricing,
		TracingStore:           stores.Tracing,
		MemoryStore:            stores.Memory,
		TenantStore:            stores.Tenants,
		BuiltinToolTenantCfgs:  stores.BuiltinToolTenantCfgs,
		SkillTenantCfgs:        stores.SkillTenantCfgs,
		SystemConfigs:          stores.SystemConfigs,
		Workspace:              workspace,
		// TTS removed in v3.x
		AutoInjector:   autoInjector,
		DomainBus:      domainBus,
		HookDispatcher: hookDispatcher,
		OnTextUploaded: func(ctx context.Context, path, content string) {
			if vaultIntc != nil {
				vaultIntc.AfterWrite(ctx, path, content)
			}
		},
		OnEvent: func(event agent.AgentEvent) {
			// Sign /v1/files/ and /v1/media/ URLs in content before delivery.
			// Sessions store clean paths; signing happens only at delivery time.
			secret := httpapi.FileSigningKey()
			switch m := event.Payload.(type) {
			case map[string]string:
				if c, has := m["content"]; has && strings.Contains(c, "/v1/") {
					m["content"] = httpapi.SignFileURLs(c, secret)
				}
			case map[string]any:
				// Sign /v1/ URLs in content text (run.completed payload is map[string]any).
				if c, ok := m["content"].(string); ok && strings.Contains(c, "/v1/") {
					m["content"] = httpapi.SignFileURLs(c, secret)
				}
				// Convert media local paths → signed /v1/files/{full_path}?ft=hash
				if rawMedia, ok := m["media"].([]agent.MediaResult); ok {
					// Clone slice — the original is shared with RunResult.Media;
					// mutating in-place corrupts paths for downstream consumers
					// (announce queue, outbound channels) that expect local paths.
					signed := make([]agent.MediaResult, len(rawMedia))
					for i, mr := range rawMedia {
						signed[i] = mr
						// Use full path so backend resolves directly via os.Stat,
						// no findInWorkspace fallback needed.
						urlPath := strings.TrimPrefix(filepath.Clean(mr.Path), "/")
						url := "/v1/files/" + urlPath
						ft := httpapi.SignFileToken(url, secret, httpapi.FileTokenTTL)
						signed[i].Path = url + "?ft=" + ft
					}
					m["media"] = signed
				}
			}
			msgBus.Broadcast(bus.Event{
				Name:     protocol.EventAgent,
				Payload:  event,
				TenantID: event.TenantID,
			})
		},
	})
	agentRouter.SetResolver(resolver)

	// Wire virtual FS interceptors: route context + memory file reads/writes to DB.
	// Share ONE ContextFileInterceptor instance between read_file and write_file
	// so they share the same cache.
	// Write-capable tools share a memory interceptor with optional KG extraction hook.
	var writeMemIntc *tools.MemoryInterceptor
	if stores.Memory != nil {
		writeMemIntc = tools.NewMemoryInterceptor(stores.Memory, workspace)
	}
	if readTool, ok := toolsReg.Get("read_file"); ok {
		if ia, ok := readTool.(tools.InterceptorAware); ok {
			if contextFileInterceptor != nil {
				ia.SetContextFileInterceptor(contextFileInterceptor)
			}
			if stores.Memory != nil {
				ia.SetMemoryInterceptor(tools.NewMemoryInterceptor(stores.Memory, workspace))
			}
		}
	}
	if writeTool, ok := toolsReg.Get("write_file"); ok {
		if ia, ok := writeTool.(tools.InterceptorAware); ok {
			if contextFileInterceptor != nil {
				ia.SetContextFileInterceptor(contextFileInterceptor)
			}
			if writeMemIntc != nil {
				ia.SetMemoryInterceptor(writeMemIntc)
			}
		}
	}
	if editTool, ok := toolsReg.Get("edit"); ok {
		if ia, ok := editTool.(tools.InterceptorAware); ok {
			if contextFileInterceptor != nil {
				ia.SetContextFileInterceptor(contextFileInterceptor)
			}
			if writeMemIntc != nil {
				ia.SetMemoryInterceptor(writeMemIntc)
			}
		}
	}
	if listTool, ok := toolsReg.Get("list_files"); ok {
		if ia, ok := listTool.(tools.InterceptorAware); ok {
			if stores.Memory != nil {
				ia.SetMemoryInterceptor(tools.NewMemoryInterceptor(stores.Memory, workspace))
			}
		}
	}

	// Wire config perm store for file writer permission checks
	if stores.ConfigPermissions != nil {
		for _, toolName := range []string{"read_file", "write_file", "edit", "cron"} {
			if t, ok := toolsReg.Get(toolName); ok {
				if cpa, ok := t.(tools.ConfigPermAware); ok {
					cpa.SetConfigPermStore(stores.ConfigPermissions)
				}
			}
		}
		if contextFileInterceptor != nil {
			contextFileInterceptor.SetConfigPermStore(stores.ConfigPermissions)
		}
	}

	// Wire memory store on memory tools (search + get)
	if stores.Memory != nil {
		if searchTool, ok := toolsReg.Get("memory_search"); ok {
			if ms, ok := searchTool.(tools.MemoryStoreAware); ok {
				ms.SetMemoryStore(stores.Memory)
			}
		}
		if getTool, ok := toolsReg.Get("memory_get"); ok {
			if ms, ok := getTool.(tools.MemoryStoreAware); ok {
				ms.SetMemoryStore(stores.Memory)
			}
		}
		slog.Info("memory layering enabled")
	}

	// Context file cache: invalidate on agent/context data changes
	if contextFileInterceptor != nil {
		msgBus.Subscribe(bus.TopicCacheBootstrap, func(event bus.Event) {
			if event.Name != protocol.EventCacheInvalidate {
				return
			}
			payload, ok := event.Payload.(bus.CacheInvalidatePayload)
			if !ok {
				return
			}
			if payload.Kind == bus.CacheKindBootstrap || payload.Kind == bus.CacheKindAgent {
				if payload.Key != "" {
					agentID, err := uuid.Parse(payload.Key)
					if err == nil {
						contextFileInterceptor.InvalidateAgent(agentID)
					}
				} else {
					contextFileInterceptor.InvalidateAll()
				}
			}
		})
	}

	// Agent router: invalidate Loop cache on agent config changes
	msgBus.Subscribe(bus.TopicCacheAgent, func(event bus.Event) {
		if event.Name != protocol.EventCacheInvalidate {
			return
		}
		payload, ok := event.Payload.(bus.CacheInvalidatePayload)
		if !ok || payload.Kind != bus.CacheKindAgent {
			return
		}
		if payload.Key != "" {
			agentRouter.InvalidateAgent(payload.Key)
		}
	})

	// Skills cache: bump version on every event (global listCache is
	// version-keyed, so bump invalidates every tenant's ListSkills cache —
	// cheap since rebuild is a single DB read). Then the agent router
	// receives a scoped wipe: tenant-scoped events only wipe that tenant's
	// cached Loops; master/global events wipe the entire router cache.
	if stores.Skills != nil {
		msgBus.Subscribe(bus.TopicCacheSkills, func(event bus.Event) {
			if event.Name != protocol.EventCacheInvalidate {
				return
			}
			payload, ok := event.Payload.(bus.CacheInvalidatePayload)
			if !ok || payload.Kind != bus.CacheKindSkills {
				return
			}
			stores.Skills.BumpVersion()
			if payload.TenantID != uuid.Nil {
				agentRouter.InvalidateTenant(payload.TenantID)
				return
			}
			agentRouter.InvalidateAll()
		})
	}

	// Skill grants cache: invalidate all agent caches when grants change
	msgBus.Subscribe(bus.TopicCacheSkillGrants, func(event bus.Event) {
		if event.Name != protocol.EventCacheInvalidate {
			return
		}
		payload, ok := event.Payload.(bus.CacheInvalidatePayload)
		if !ok || payload.Kind != bus.CacheKindSkillGrants {
			return
		}
		agentRouter.InvalidateAll()
	})

	// MCP cache: invalidate all agent caches when MCP servers/grants change
	msgBus.Subscribe(bus.TopicCacheMCP, func(event bus.Event) {
		if event.Name != protocol.EventCacheInvalidate {
			return
		}
		payload, ok := event.Payload.(bus.CacheInvalidatePayload)
		if !ok || payload.Kind != bus.CacheKindMCP {
			return
		}
		agentRouter.InvalidateAll()
	})

	// Builtin tools cache: re-apply disables on settings/enabled changes.
	// Tenant-scoped events only invalidate that tenant's cached agents — the
	// global registry disables list is master-only and unaffected.
	if stores.BuiltinTools != nil {
		msgBus.Subscribe(bus.TopicCacheBuiltinTools, func(event bus.Event) {
			if event.Name != protocol.EventCacheInvalidate {
				return
			}
			payload, ok := event.Payload.(bus.CacheInvalidatePayload)
			if !ok || payload.Kind != bus.CacheKindBuiltinTools {
				return
			}
			if payload.TenantID != uuid.Nil {
				agentRouter.InvalidateTenant(payload.TenantID)
				return
			}
			applyBuiltinToolDisables(context.Background(), stores.BuiltinTools, toolsReg)
			agentRouter.InvalidateAll()
		})
	}

	// Register team tools (team_tasks + workspace interceptor) if team store is available.
	var postTurn tools.PostTurnProcessor

	// User workspace cache: invalidate per-user workspace path on profile changes
	msgBus.Subscribe(bus.TopicCacheUserWorkspace, func(event bus.Event) {
		if event.Name != protocol.EventCacheInvalidate {
			return
		}
		payload, ok := event.Payload.(bus.CacheInvalidatePayload)
		if !ok || payload.Kind != bus.CacheKindUserWorkspace {
			return
		}
		if payload.Key != "" {
			agentRouter.InvalidateUserWorkspace(payload.Key)
		}
	})

	// Provider cache: re-register ACP providers on create/update/delete
	msgBus.Subscribe(bus.TopicCacheProvider, func(event bus.Event) {
		if event.Name != protocol.EventCacheInvalidate {
			return
		}
		payload, ok := event.Payload.(bus.CacheInvalidatePayload)
		if !ok || payload.Kind != bus.CacheKindProvider {
			return
		}
		if payload.Key == "" {
			return
		}
		// Re-register from DB if provider still exists and is ACP type
		provCtx := store.WithTenantID(context.Background(), event.TenantID)
		p, err := stores.Providers.GetProviderByName(provCtx, payload.Key)
		if err != nil {
			// Provider was deleted or not found — already unregistered by handler
			return
		}
		if p.ProviderType != store.ProviderACP {
			return
		}
		// Unregister old instance (closes ProcessPool) then re-register
		providerReg.Unregister(p.Name)
		if p.Enabled {
			registerACPFromDB(providerReg, *p)
		}
	})

	slog.Info("resolver + interceptors + cache subscribers wired")
	return contextFileInterceptor, mcpPool, mediaStore, postTurn
}

