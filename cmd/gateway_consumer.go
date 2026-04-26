package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/nextlevelbuilder/goclaw/internal/agent"
	"github.com/nextlevelbuilder/goclaw/internal/bus"
	"github.com/nextlevelbuilder/goclaw/internal/channels"
	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/scheduler"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
)

// consumeInboundMessages reads inbound messages from channels (Telegram, Discord, etc.)
// and routes them through the scheduler/agent loop, then publishes the response back.
// Also handles subagent announcements: routes them through the parent agent's session
// (matching TS subagent-announce.ts pattern) so the agent can reformulate for the user.
func consumeInboundMessages(ctx context.Context, msgBus *bus.MessageBus, agents *agent.Router, cfg *config.Config, sched *scheduler.Scheduler, channelMgr *channels.Manager, teamStore store.TeamStore, quotaChecker *channels.QuotaChecker, sessStore store.SessionStore, agentStore store.AgentStore, contactCollector *store.ContactCollector, postTurn tools.PostTurnProcessor, subagentMgr *tools.SubagentManager) {
	slog.Info("inbound message consumer started")

	// Inbound message deduplication (matching TS src/infra/dedupe.ts + inbound-dedupe.ts).
	// TTL=20min, max=5000 entries — prevents webhook retries / double-taps from duplicating agent runs.
	dedupe := bus.NewDedupeCache(20*time.Minute, 5000)

	// Per-session announce serialization: prevents concurrent announce runs from
	// reading stale session history. Without this, Announce #2 can start while
	// Announce #1 is still running, read history that doesn't include Announce #1's
	// messages (written only after agent loop completes), and generate responses
	// with wrong context (e.g. "waiting for Tiểu La" when Tiểu La already finished).
	var announceMu sync.Map // sessionKey → *sync.Mutex
	getAnnounceMu := func(key string) *sync.Mutex {
		v, _ := announceMu.LoadOrStore(key, &sync.Mutex{})
		return v.(*sync.Mutex)
	}

	// Construct shared dependencies once — passed by pointer to all handlers.
	deps := &ConsumerDeps{
		Cfg:              cfg,
		Agents:           agents,
		Sched:            sched,
		ChannelMgr:       channelMgr,
		MsgBus:           msgBus,
		TeamStore:        teamStore,
		AgentStore:       agentStore,
		SessStore:        sessStore,
		PostTurn:         postTurn,
		QuotaChecker:     quotaChecker,
		ContactCollector: contactCollector,
		SubagentMgr:      subagentMgr,
		GetAnnounceMu:    getAnnounceMu,
	}

	// Track running teammate tasks so they can be cancelled when the task is
	// cancelled/failed externally (e.g. lead cancels via team_tasks tool).
	msgBus.Subscribe("consumer.team-task-cancel", func(event bus.Event) {
		if event.Name != protocol.EventTeamTaskCancelled && event.Name != protocol.EventTeamTaskFailed {
			return
		}
		payload, ok := event.Payload.(protocol.TeamTaskEventPayload)
		if !ok {
			return
		}
		if sessKey, ok := deps.TaskRunSessions.Load(payload.TaskID); ok {
			if cancelled := sched.CancelSession(sessKey.(string)); cancelled {
				slog.Info("team task cancelled: stopped running agent",
					"task_id", payload.TaskID, "session", sessKey)
			}
			deps.TaskRunSessions.Delete(payload.TaskID)
		}
	})

	// Inbound debounce: merge rapid messages from the same sender before processing.
	// Matching TS createInboundDebouncer from src/auto-reply/inbound-debounce.ts.
	debounceMs := cfg.Gateway.InboundDebounceMs
	if debounceMs == 0 {
		debounceMs = 1000 // default: 1000ms
	}
	debouncer := bus.NewInboundDebouncer(
		time.Duration(debounceMs)*time.Millisecond,
		func(msg bus.InboundMessage) {
			processNormalMessage(ctx, msg, deps)
		},
	)
	defer debouncer.Stop()

	slog.Info("inbound debounce configured", "debounce_ms", debounceMs)

	// Track background goroutines (subagent announces, teammate messages)
	// so shutdown can wait for in-flight work to complete.
	defer func() {
		deps.BgWg.Wait()
		slog.Info("inbound consumer: all background goroutines drained")
	}()

	for {
		msg, ok := msgBus.ConsumeInbound(ctx)
		if !ok {
			slog.Info("inbound message consumer stopped")
			return
		}

		// --- Dedup: skip duplicate inbound messages (matching TS shouldSkipDuplicateInbound) ---
		if msgID := msg.Metadata["message_id"]; msgID != "" {
			dedupeKey := fmt.Sprintf("%s|%s|%s|%s", msg.Channel, msg.SenderID, msg.ChatID, msgID)
			if dedupe.IsDuplicate(dedupeKey) {
				slog.Debug("dedup: skipping duplicate message", "key", dedupeKey)
				continue
			}
		}

		if handleResetCommand(msg, deps) {
			continue
		}
		if handleStopCommand(msg, deps) {
			continue
		}

		// Blocker escalation messages bypass debounce — deliver immediately to leader.
		if msg.SenderID == "system:escalation" {
			go processNormalMessage(ctx, msg, deps)
			continue
		}

		// --- Normal messages: route through debouncer ---
		debouncer.Push(msg)
	}
}

// truncateForReminder truncates content to maxLen chars, taking the last line as context.
func truncateForReminder(content string, maxLen int) string {
	// Use last non-empty line as it's typically the most relevant.
	lines := strings.Split(strings.TrimSpace(content), "\n")
	msg := lines[len(lines)-1]
	// Ensure we only persist valid UTF-8 into PostgreSQL.
	msg = strings.ToValidUTF8(msg, "")
	if maxLen <= 0 {
		return msg
	}
	if utf8.RuneCountInString(msg) > maxLen {
		r := []rune(msg)
		msg = string(r[:maxLen]) + "..."
	}
	return msg
}

// appendMediaToOutbound converts agent MediaResults to outbound MediaAttachments
// on the given OutboundMessage. Handles voice annotation when applicable.
func appendMediaToOutbound(msg *bus.OutboundMessage, media []agent.MediaResult) {
	for _, mr := range media {
		msg.Media = append(msg.Media, bus.MediaAttachment{
			URL:         mr.Path,
			ContentType: mr.ContentType,
		})
		if mr.AsVoice {
			if msg.Metadata == nil {
				msg.Metadata = make(map[string]string)
			}
			msg.Metadata["audio_as_voice"] = "true"
		}
	}
}
