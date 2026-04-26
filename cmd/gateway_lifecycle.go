package cmd

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nextlevelbuilder/goclaw/internal/edition"
	"github.com/nextlevelbuilder/goclaw/internal/scheduler"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/tasks"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
)

// lifecycleDeps bundles the extra parameters needed by runLifecycle that are not in gatewayDeps.
type lifecycleDeps struct {
	sched             *scheduler.Scheduler
	postTurn          tools.PostTurnProcessor
	subagentMgr       *tools.SubagentManager
	consumerTeamStore store.TeamStore
	sigCh             chan os.Signal
}

// runLifecycle wires config-reload subscribers, starts consumers, task recovery,
// the signal handler goroutine, and finally starts the gateway server.
// This is the last phase of runGateway() — called after all setup is complete.
func (d *gatewayDeps) runLifecycle(
	ctx context.Context,
	cancel context.CancelFunc,
	deps lifecycleDeps,
) {

	// Contact collector: auto-collect user info from channels with in-memory dedup cache.
	var contactCollector *store.ContactCollector

	go consumeInboundMessages(ctx, d.msgBus, d.agentRouter, d.cfg, deps.sched, d.channelMgr, deps.consumerTeamStore, nil, d.pgStores.Sessions, d.pgStores.Agents, contactCollector, deps.postTurn, deps.subagentMgr)

	// Task recovery ticker: re-dispatches stale/pending team tasks on startup and periodically.
	var taskTicker *tasks.TaskTicker
	if d.pgStores.Teams != nil {
		taskTicker = tasks.NewTaskTicker(d.pgStores.Teams, d.pgStores.Agents, d.msgBus, d.cfg.Gateway.TaskRecoveryIntervalSec)
		taskTicker.Start()
	}

	go func() {
		sig := <-deps.sigCh
		slog.Info("graceful shutdown initiated", "signal", sig)

		// Broadcast shutdown event
		d.server.BroadcastEvent(*protocol.NewEvent(protocol.EventShutdown, nil))

		// Stop channels, cron, heartbeat, and task ticker
		if taskTicker != nil {
			taskTicker.Stop()
		}

		// Close provider resources (e.g. Claude CLI temp files)
		d.providerRegistry.Close()

		// Stop permission cache sweep goroutines so they don't leak past shutdown.
		if d.permCache != nil {
			d.permCache.Close()
		}
		if deps.sched != nil {
			slog.Info("gateway: draining active runs", "timeout", "5s")
			deps.sched.Stop() // MarkDraining + StopAll
			time.Sleep(5 * time.Second)
		}

		cancel()
	}()

	slog.Info("goclaw gateway starting",
		"version", Version,
		"protocol", protocol.ProtocolVersion,
		"agents", d.agentRouter.List(),
		"tools", d.toolsReg.Count(),
		"channels", d.channelMgr.GetEnabledChannels(),
	)

	// Tailscale listener: build the mux first, then pass it to initTailscale
	// so the same routes are served on both the main listener and Tailscale.
	// Compiled via build tags: `go build -tags tsnet` to enable.
	mux := d.server.BuildMux()

	tsCleanup := initTailscale(ctx, d.cfg, mux)
	if tsCleanup != nil {
		defer tsCleanup()
	}

	// Phase 1: suggest localhost binding when Tailscale is active
	if d.cfg.Tailscale.Hostname != "" && d.cfg.Gateway.Host == "0.0.0.0" {
		slog.Info("Tailscale enabled. Consider setting GOCLAW_HOST=127.0.0.1 for localhost-only + Tailscale access")
	}

	// Security warnings
	if strings.Contains(d.cfg.Database.PostgresDSN, ":goclaw@") {
		slog.Warn("security.default_db_password: using default Postgres password — run ./prepare-env.sh to generate a strong one")
	}
	if len(d.cfg.Gateway.AllowedOrigins) > 0 {
		slog.Info("cors: allowed_origins configured", "origins", d.cfg.Gateway.AllowedOrigins)
	} else if !edition.Current().IsLimited() {
		slog.Warn("security.cors_open: no allowed_origins configured — all WebSocket origins accepted. Set gateway.allowed_origins or GOCLAW_ALLOWED_ORIGINS for production")
	}

	if err := d.server.Start(ctx); err != nil {
		slog.Error("gateway error", "error", err)
		os.Exit(1)
	}
}
