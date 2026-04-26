package store

import "database/sql"

// Stores is the top-level container for all storage backends.
type Stores struct {
	DB                    *sql.DB // underlying connection
	Sessions              SessionStore
	Memory                MemoryStore
	Skills                SkillStore
	Agents                AgentStore
	Providers             ProviderStore
	ConfigSecrets         ConfigSecretsStore
	BuiltinTools          BuiltinToolStore
	APIKeys               APIKeyStore
	ConfigPermissions     ConfigPermissionStore
	Tenants               TenantStore
	BuiltinToolTenantCfgs BuiltinToolTenantConfigStore
	SkillTenantCfgs       SkillTenantConfigStore
	SystemConfigs         SystemConfigStore

	// Removed subsystems - stubs only
	MCP             MCPServerStore // stub
	Episodic        interface{}    // stub
	KnowledgeGraph  interface{}    // stub
	Teams           TeamStore
	Tracing         TracingStore
	Heartbeat       HeartbeatStore
	Heartbeats      HeartbeatStore // heartbeat store
	AgentLinks      AgentLinkStore
	Cron            CronStore
	Snapshots       SnapshotStore
	PendingMessages PendingMessageStore
	Pairing         PairingStore
	Vault           interface{}             // stub
	Hooks           interface{}             // stub - hooks removed
	ChannelInstances interface{}            // stub
	SecureCLI        SecureCLIStore         // secure CLI binary configs
	SecureCLIGrants  SecureCLIAgentGrantStore // per-agent grants
	SubagentTasks    SubagentTaskStore      // subagent task audit trail
}
