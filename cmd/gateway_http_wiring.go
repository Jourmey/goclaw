package cmd

import (
	"context"
	"log/slog"

	"github.com/nextlevelbuilder/goclaw/internal/bus"
	httpapi "github.com/nextlevelbuilder/goclaw/internal/http"
	mcpbridge "github.com/nextlevelbuilder/goclaw/internal/mcp"
	"github.com/nextlevelbuilder/goclaw/internal/media"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/store/pg"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

// httpHandlers bundles the results of wireHTTP() for passing to wireHTTPHandlersOnServer.
type httpHandlers struct {
	agents           *httpapi.AgentsHandler
	skills           *httpapi.SkillsHandler
	traces           *httpapi.TracesHandler
	mcp              *httpapi.MCPHandler
	channelInstances *httpapi.ChannelInstancesHandler
	providers        *httpapi.ProvidersHandler
	builtinTools     *httpapi.BuiltinToolsHandler
	pendingMessages  *httpapi.PendingMessagesHandler
	teamEvents       *httpapi.TeamEventsHandler
	secureCLI        *httpapi.SecureCLIHandler
	secureCLIGrant   *httpapi.SecureCLIGrantHandler
	mcpUserCreds     *httpapi.MCPUserCredentialsHandler
}

// wireHTTPHandlersOnServer registers all HTTP handler objects onto the gateway server.
// Called after wireHTTP() and wireExtras() have returned their results.
func (d *gatewayDeps) wireHTTPHandlersOnServer(
	h httpHandlers,
	wakeH *httpapi.WakeHandler,
	mcpPool *mcpbridge.Pool,
	postTurn tools.PostTurnProcessor,
	mediaStore *media.Store,
) {
	if h.providers != nil {
		h.providers.SetAPIBaseFallback(d.cfg.Providers.APIBaseForType)
	}
	if h.agents != nil {
		d.server.SetAgentsHandler(h.agents)
	}
	if h.skills != nil {
		d.server.SetSkillsHandler(h.skills)
	}
	if h.traces != nil {
		d.server.SetTracesHandler(h.traces)
	}
	// External wake/trigger API — wakeH was created by caller before invoking this method.
	d.server.SetWakeHandler(wakeH)
	if h.mcp != nil {
		if mcpPool != nil {
			h.mcp.SetPoolEvictor(mcpPool)
		}
		d.server.SetMCPHandler(h.mcp)
	}
	if h.mcpUserCreds != nil {
		d.server.SetMCPUserCredentialsHandler(h.mcpUserCreds)
	}
	if h.channelInstances != nil {
		d.server.SetChannelInstancesHandler(h.channelInstances)
	}
	if h.providers != nil {
		d.server.SetProvidersHandler(h.providers)
	}
	if h.teamEvents != nil {
		d.server.SetTeamEventsHandler(h.teamEvents)
	}
	if d.pgStores != nil && d.pgStores.Teams != nil {
		d.server.SetTeamAttachmentsHandler(httpapi.NewTeamAttachmentsHandler(d.pgStores.Teams, d.workspace))
		d.server.SetWorkspaceUploadHandler(httpapi.NewWorkspaceUploadHandler(d.pgStores.Teams, d.workspace, d.msgBus))
	}
	if h.builtinTools != nil {
		d.server.SetBuiltinToolsHandler(h.builtinTools)
	}
	if h.secureCLI != nil {
		d.server.SetSecureCLIHandler(h.secureCLI)
	}
	if h.secureCLIGrant != nil {
		d.server.SetSecureCLIGrantHandler(h.secureCLIGrant)
	}

	// System configs API
	if d.pgStores.SystemConfigs != nil {
		d.server.SetSystemConfigsHandler(httpapi.NewSystemConfigsHandler(d.pgStores.SystemConfigs, d.msgBus))

		// Refresh in-memory config when system_configs change via HTTP API
		d.msgBus.Subscribe(bus.TopicSystemConfigChanged, func(evt bus.Event) {
			// Use tenant context from the request that triggered the change
			ctx := context.Background()
			if reqCtx, ok := evt.Payload.(context.Context); ok {
				ctx = reqCtx
			} else {
				ctx = store.WithTenantID(ctx, store.MasterTenantID)
			}
			if sysConfigs, err := d.pgStores.SystemConfigs.List(ctx); err == nil && len(sysConfigs) > 0 {
				d.cfg.ApplySystemConfigs(sysConfigs)
				// Update PGMemoryStore chunk config so new documents use updated settings
				if mem := d.cfg.Agents.Defaults.Memory; mem != nil {
					if pgMem, ok := d.pgStores.Memory.(*pg.PGMemoryStore); ok {
						pgMem.UpdateChunkConfig(mem.MaxChunkLen, mem.ChunkOverlap)
					}
				}
				// Note: vault enrichment provider is resolved per-tenant at runtime,
				// no hot-reload needed here
				slog.Debug("system_configs refreshed to in-memory config", "keys", len(sysConfigs))
			}
		})
	}

	// Usage analytics API
	if d.pgStores.Snapshots != nil {
		d.server.SetUsageHandler(httpapi.NewUsageHandler(d.pgStores.Snapshots, d.pgStores.DB))
	}

	// Runtime package management (install/uninstall system/pip/npm/github packages)
	initGitHubInstaller()
	d.server.SetPackagesHandler(httpapi.NewPackagesHandler())

	// API documentation (OpenAPI spec + Swagger UI at /docs)
	d.server.SetDocsHandler(httpapi.NewDocsHandler())

	// Edition info (public, no auth — used by desktop UI comparison modal)
	d.server.SetEditionHandler(httpapi.NewEditionHandler())

	if d.pgStores != nil && d.pgStores.APIKeys != nil {
		d.server.SetAPIKeysHandler(httpapi.NewAPIKeysHandler(d.pgStores.APIKeys, d.msgBus))
		d.server.SetAPIKeyStore(d.pgStores.APIKeys)
		httpapi.InitAPIKeyCache(d.pgStores.APIKeys, d.msgBus)
	}

	// Allow browser-paired users to access HTTP APIs
	if d.pgStores.Pairing != nil {
		httpapi.InitPairingAuth(d.pgStores.Pairing)
	}

	// Memory management API
	if d.pgStores != nil && d.pgStores.Memory != nil {
		d.server.SetMemoryHandler(httpapi.NewMemoryHandler(d.pgStores.Memory))
	}

	// V3: Orchestration mode API (read-only)
	if d.pgStores != nil && d.pgStores.Agents != nil {
		d.server.SetOrchestrationHandler(httpapi.NewOrchestrationHandler(d.pgStores.Agents, d.pgStores.Teams, d.pgStores.AgentLinks))
	}

	// V3: Per-agent v3 feature flags API
	if d.pgStores != nil && d.pgStores.Agents != nil {
		d.server.SetV3FlagsHandler(httpapi.NewV3FlagsHandler(d.pgStores.Agents))
	}

	// Workspace file serving endpoint — serves files by absolute path, auth-token protected.
	d.server.SetFilesHandler(httpapi.NewFilesHandler(d.workspace, d.dataDir))

	// Storage file management — browse/delete files under the resolved workspace directory.
	d.server.SetStorageHandler(httpapi.NewStorageHandler(d.workspace))

	// Media upload endpoint — accepts multipart file uploads, returns temp path + MIME type.
	d.server.SetMediaUploadHandler(httpapi.NewMediaUploadHandler())

	// Media serve endpoint — serves persisted media files by ID for WS/web clients.
	if mediaStore != nil {
		d.server.SetMediaServeHandler(httpapi.NewMediaServeHandler(mediaStore))
	}

	// TTS removed in v3.x

	// Seed + apply builtin tool disables
	if d.pgStores.BuiltinTools != nil {
		seedBuiltinTools(context.Background(), d.pgStores.BuiltinTools)
		migrateBuiltinToolSettings(context.Background(), d.pgStores.BuiltinTools)
		backfillWebFetchSettings(context.Background(), d.pgStores.BuiltinTools)
		applyBuiltinToolDisables(context.Background(), d.pgStores.BuiltinTools, d.toolsReg)
	}
}
