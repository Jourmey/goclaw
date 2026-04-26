package store

import "database/sql"

// Stores is the top-level container for all storage backends.
type Stores struct {
	DB                     *sql.DB // underlying connection
	Sessions               SessionStore
	Memory                 MemoryStore
	Skills                 SkillStore
	Agents                 AgentStore
	Providers              ProviderStore
	ConfigSecrets          ConfigSecretsStore
	BuiltinTools           BuiltinToolStore
	APIKeys                APIKeyStore
	ConfigPermissions      ConfigPermissionStore
	Tenants                TenantStore
	BuiltinToolTenantCfgs  BuiltinToolTenantConfigStore
	SkillTenantCfgs        SkillTenantConfigStore
	SystemConfigs          SystemConfigStore
}
