package pg

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// NewPGStores creates all stores backed by Postgres.
func NewPGStores(cfg store.StoreConfig) (*store.Stores, error) {
	db, err := OpenDB(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	initSqlx(db)

	memCfg := DefaultPGMemoryConfig()

	skillsDir := cfg.SkillsStorageDir
	if skillsDir == "" {
		skillsDir = config.ResolvedDataDirFromEnv() + "/skills-store"
	}

	return &store.Stores{
		DB:                    db,
		Sessions:              NewPGSessionStore(db),
		Memory:                NewPGMemoryStore(db, memCfg),
		Skills:                NewPGSkillStore(db, skillsDir),
		Agents:                NewPGAgentStore(db),
		Providers:             NewPGProviderStore(db, cfg.EncryptionKey),
		ConfigSecrets:         NewPGConfigSecretsStore(db, cfg.EncryptionKey),
		BuiltinTools:          NewPGBuiltinToolStore(db),
		APIKeys:               NewPGAPIKeyStore(db),
		ConfigPermissions:     NewPGConfigPermissionStore(db),
		Tenants:               NewPGTenantStore(db),
		BuiltinToolTenantCfgs: NewPGBuiltinToolTenantConfigStore(db),
		SkillTenantCfgs:       NewPGSkillTenantConfigStore(db),
		SystemConfigs:         NewPGSystemConfigStore(db),
	}, nil
}
