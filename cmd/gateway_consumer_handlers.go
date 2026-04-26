package cmd

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"strings"

	"github.com/nextlevelbuilder/goclaw/internal/bus"
	"github.com/nextlevelbuilder/goclaw/internal/providers"
	"github.com/nextlevelbuilder/goclaw/internal/sessions"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

// handleResetCommand processes /reset command: clears session history.
// Returns true if the message was handled (caller should continue).
func handleResetCommand(
	msg bus.InboundMessage,
	deps *ConsumerDeps,
) bool {
	if msg.Metadata[tools.MetaCommand] != "reset" {
		return false
	}

	agentID := msg.AgentID
	if agentID == "" {
		agentID = resolveAgentRoute(deps.Cfg, msg.Channel, msg.ChatID, msg.PeerKind)
	}
	peerKind := msg.PeerKind
	if peerKind == "" {
		peerKind = string(sessions.PeerDirect)
	}
	sessionKey := sessions.BuildScopedSessionKey(agentID, msg.Channel, sessions.PeerKind(peerKind), msg.ChatID)
	if msg.Metadata[tools.MetaIsForum] == "true" && peerKind == string(sessions.PeerGroup) {
		var topicID int
		fmt.Sscanf(msg.Metadata[tools.MetaMessageThreadID], "%d", &topicID)
		if topicID > 0 {
			sessionKey = sessions.BuildGroupTopicSessionKey(agentID, msg.Channel, msg.ChatID, topicID)
		}
	}
	ctx := store.WithTenantID(context.Background(), msg.TenantID)
	deps.SessStore.Reset(ctx, sessionKey)
	deps.SessStore.Save(ctx, sessionKey)
	providers.ResetCLISession("", sessionKey)
	slog.Info("inbound: /reset command", "session", sessionKey)

	return true
}

// handleStopCommand processes /stop and /stopall commands: cancel active runs for a session.
// Returns true if the message was handled (caller should continue).
func handleStopCommand(
	msg bus.InboundMessage,
	deps *ConsumerDeps,
) bool {
	cmd := msg.Metadata[tools.MetaCommand]
	if cmd != "stop" && cmd != "stopall" {
		return false
	}

	agentID := msg.AgentID
	if agentID == "" {
		agentID = resolveAgentRoute(deps.Cfg, msg.Channel, msg.ChatID, msg.PeerKind)
	}
	peerKind := msg.PeerKind
	if peerKind == "" {
		peerKind = string(sessions.PeerDirect)
	}
	sessionKey := sessions.BuildScopedSessionKey(agentID, msg.Channel, sessions.PeerKind(peerKind), msg.ChatID)
	if msg.Metadata[tools.MetaIsForum] == "true" && peerKind == string(sessions.PeerGroup) {
		var topicID int
		fmt.Sscanf(msg.Metadata[tools.MetaMessageThreadID], "%d", &topicID)
		if topicID > 0 {
			sessionKey = sessions.BuildGroupTopicSessionKey(agentID, msg.Channel, msg.ChatID, topicID)
		}
	}
	if msg.Metadata[tools.MetaDMThreadID] != "" && peerKind == string(sessions.PeerDirect) {
		var threadID int
		fmt.Sscanf(msg.Metadata[tools.MetaDMThreadID], "%d", &threadID)
		if threadID > 0 {
			sessionKey = sessions.BuildDMThreadSessionKey(agentID, msg.Channel, msg.ChatID, threadID)
		}
	}

	// sessStore is referenced in the original code but not used in this branch beyond
	// session key construction; kept as parameter for API consistency.
	_ = deps.SessStore

	var cancelled bool
	if cmd == "stopall" {
		cancelled = deps.Sched.CancelSession(sessionKey)
		slog.Info("inbound: /stopall command", "session", sessionKey, "cancelled", cancelled)
	} else {
		cancelled = deps.Sched.CancelOneSession(sessionKey)
		slog.Info("inbound: /stop command", "session", sessionKey, "cancelled", cancelled)
	}

	// Publish feedback so the channel can show the result.
	var feedback string
	if cancelled {
		if cmd == "stopall" {
			feedback = "All tasks stopped."
		} else {
			feedback = "Task stopped."
		}
	} else {
		if cmd == "stopall" {
			feedback = "No active tasks to stop."
		} else {
			feedback = "No active task to stop."
		}
	}
	deps.MsgBus.PublishOutbound(bus.OutboundMessage{
		Channel:  msg.Channel,
		ChatID:   msg.ChatID,
		Content:  feedback,
		Metadata: msg.Metadata,
	})

	return true
}

// buildTaskBoardSnapshot returns a formatted summary of batch task statuses
// for inclusion in the announce message to the leader. Scoped by (teamID, chatID)
// and filtered by origin_trace_id to show only tasks from the current batch.
func buildTaskBoardSnapshot(ctx context.Context, teamStore store.TeamStore, teamID uuid.UUID, chatID, originTraceID string) string {
	if teamStore == nil || originTraceID == "" {
		return ""
	}
	// Shared workspace: show all tasks across chats.
	snapshotChatID := chatID
	if team, err := teamStore.GetTeam(ctx, teamID); err == nil && tools.IsSharedWorkspace(team.Settings) {
		snapshotChatID = ""
	}
	allTasks, err := teamStore.ListTasks(ctx, teamID, "", store.TeamTaskFilterAll, "", "", snapshotChatID, 0, 0)
	if err != nil || len(allTasks) == 0 {
		return ""
	}

	// Filter to current batch by origin_trace_id stored in task metadata.
	var active, completed int
	var activeLines []string
	for _, t := range allTasks {
		tid, _ := t.Metadata[tools.TaskMetaOriginTrace].(string)
		if tid != originTraceID {
			continue
		}
		switch t.Status {
		case store.TeamTaskStatusCompleted, store.TeamTaskStatusCancelled, store.TeamTaskStatusFailed:
			completed++
		default:
			active++
			activeLines = append(activeLines, fmt.Sprintf("  #%d %s — %s", t.TaskNumber, t.Subject, t.Status))
		}
	}
	total := active + completed
	if total == 0 {
		return ""
	}
	if active == 0 {
		return fmt.Sprintf("=== Task board (this batch) ===\nAll %d tasks completed.", total)
	}
	return fmt.Sprintf("=== Task board (this batch) ===\nTask progress: %d/%d completed, %d active:\n%s",
		completed, total, active, strings.Join(activeLines, "\n"))
}
