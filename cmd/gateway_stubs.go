package cmd

import (
	"context"

	"github.com/nextlevelbuilder/goclaw/internal/channels"
	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/gateway"
	"github.com/nextlevelbuilder/goclaw/internal/gateway/methods"
	"github.com/nextlevelbuilder/goclaw/internal/scheduler"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

// registerConfigChannels stub - channels subsystem removed
func registerConfigChannels(cfg *config.Config, mgr *channels.Manager, msgBus interface{}, stores *store.Stores, instanceLoader *channels.InstanceLoader) {
	// no-op
}

// wireChannelRPCMethods stub - channels subsystem removed
func wireChannelRPCMethods(server *gateway.Server, stores *store.Stores, mgr *channels.Manager, agentRouter interface{}, msgBus interface{}, workspace string) {
	// no-op
}

// wireChannelEventSubscribers stub - channels subsystem removed
func wireChannelEventSubscribers(msgBus interface{}, server *gateway.Server, stores *store.Stores, mgr *channels.Manager, instanceLoader *channels.InstanceLoader, pairingMethods *methods.PairingMethods, cfg *config.Config) {
	// no-op
}

// startCronAndHeartbeat stub - cron/heartbeat subsystem removed
func startCronAndHeartbeat(stores *store.Stores, server *gateway.Server, sched *scheduler.Scheduler, msgBus interface{}, providerReg interface{}, mgr *channels.Manager, cfg *config.Config, heartbeatTool *tools.HeartbeatTool, heartbeatMethods interface{}) interface{} {
	return nil
}
