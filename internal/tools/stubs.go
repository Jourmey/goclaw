package tools

import (
	"context"

	"github.com/google/uuid"
)

// VaultInterceptor stub - vault removed
type VaultInterceptor struct{}

func (v *VaultInterceptor) AfterWrite(ctx context.Context, path, content string) error {
	return nil
}

func (v *VaultInterceptor) AfterWriteMedia(ctx context.Context, path, prompt, mimeType string) {
	// no-op
}

func (v *VaultInterceptor) BeforeRead(ctx context.Context, path string) {
	// no-op
}

// WorkspaceInterceptor stub - workspace interceptor removed
type WorkspaceInterceptor struct{}

func (w *WorkspaceInterceptor) HandleWrite(ctx context.Context, path, content string) (bool, error) {
	return false, nil
}

func (w *WorkspaceInterceptor) AfterWrite(ctx context.Context, path, action string) {
	// no-op
}

// TeamToolManager stub - team removed
type TeamToolManager struct{}

// SpawnTool stub - subagent removed
type SpawnTool struct{}

// NewSpawnTool stub - subagent spawn tool constructor removed
func NewSpawnTool(manager interface{}, agentKey string, maxConcurrent int) *SpawnTool {
	return &SpawnTool{}
}

// Name stub - Tool interface
func (s *SpawnTool) Name() string {
	return "spawn"
}

// Description stub - Tool interface
func (s *SpawnTool) Description() string {
	return "spawn tool stub"
}

// Parameters stub - Tool interface
func (s *SpawnTool) Parameters() map[string]any {
	return make(map[string]any)
}

// Execute stub - Tool interface
func (s *SpawnTool) Execute(ctx context.Context, args map[string]any) *Result {
	return &Result{ForLLM: "spawn tool removed"}
}

// Invoke stub - Tool interface (legacy)
func (s *SpawnTool) Invoke(ctx interface{}, input interface{}) (interface{}, error) {
	return nil, nil
}

// CreateVideoTool stub - video creation removed
type CreateVideoTool struct{}

// PostTurnProcessor stub interface
type PostTurnProcessor interface {
	ProcessPostTurn(ctx context.Context, result interface{}) error
	ProcessPendingTasks(ctx context.Context, teamID uuid.UUID, taskIDs []uuid.UUID) error
}

// MetaOriginSenderID is a stub constant for origin sender ID
const MetaOriginSenderID = "origin_sender_id"

// MetaOriginRole stub constant
const MetaOriginRole = "origin_role"

// MetaOriginChannel stub constant
const MetaOriginChannel = "origin_channel"

// MetaOriginPeerKind stub constant
const MetaOriginPeerKind = "origin_peer_kind"

// MetaOriginLocalKey stub constant
const MetaOriginLocalKey = "origin_local_key"

// MetaParentAgent stub constant
const MetaParentAgent = "parent_agent"

// MetaOriginSessionKey stub constant
const MetaOriginSessionKey = "origin_session_key"

// MetaSubagentLabel stub constant
const MetaSubagentLabel = "subagent_label"

// MetaOriginTraceID stub constant
const MetaOriginTraceID = "origin_trace_id"

// MetaOriginRootSpanID stub constant
const MetaOriginRootSpanID = "origin_root_span_id"

// MetaSubagentRuntime stub constant
const MetaSubagentRuntime = "subagent_runtime"

// MetaSubagentIterations stub constant
const MetaSubagentIterations = "subagent_iterations"

// MetaSubagentInputToks stub constant
const MetaSubagentInputToks = "subagent_input_toks"

// MetaSubagentOutputToks stub constant
const MetaSubagentOutputToks = "subagent_output_toks"

// MetaSubagentResult stub constant
const MetaSubagentResult = "subagent_result"

// MetaSubagentStatus stub constant
const MetaSubagentStatus = "subagent_status"

// MetaOriginChatID stub constant
const MetaOriginChatID = "origin_chat_id"

// MetaTeamID stub constant
const MetaTeamID = "team_id"

// MetaTeamTaskID stub constant
const MetaTeamTaskID = "team_task_id"

// MetaFromAgent stub constant
const MetaFromAgent = "from_agent"

// MetaToAgent stub constant
const MetaToAgent = "to_agent"

// WithTaskActionFlags stub function - team action flags removed
func WithTaskActionFlags(ctx context.Context, flags interface{}) context.Context {
	return ctx
}

// MetaTeamWorkspace stub constant
const MetaTeamWorkspace = "team_workspace"

// MetaLeaderAgentID stub constant
const MetaLeaderAgentID = "leader_agent_id"

// MetaToAgentDisplay stub constant
const MetaToAgentDisplay = "to_agent_display"

// MetaOriginUserID stub constant
const MetaOriginUserID = "origin_user_id"

// MetaCommand stub constant
const MetaCommand = "command"

// MetaIsForum stub constant
const MetaIsForum = "is_forum"

// MetaMessageThreadID stub constant
const MetaMessageThreadID = "message_thread_id"

// MetaDMThreadID stub constant
const MetaDMThreadID = "dm_thread_id"

// TaskMetaOriginTrace stub constant
const TaskMetaOriginTrace = "origin_trace"

// MetaUsername stub constant
const MetaUsername = "username"

// MetaChatTitle stub constant
const MetaChatTitle = "chat_title"

// TaskLocalKeyMetadata stub function - returns task metadata map
func TaskLocalKeyMetadata(task interface{}) map[string]string {
	return nil
}

// BuildTaskEventPayload stub function
func BuildTaskEventPayload(channel, chatID, peerKind, localKey, taskNumber string) interface{} {
	return nil
}

// AutoAttachWorkspaceFile stub - auto attach removed
func AutoAttachWorkspaceFile(ctx interface{}, teamStore interface{}, workspace, teamWorkspace string) {
	// no-op
}

// FullTeamPolicy stub - full team policy removed
type FullTeamPolicy struct{}

func (FullTeamPolicy) MemberGuidance() string {
	return ""
}

// LiteTeamPolicy stub - lite team policy removed
type LiteTeamPolicy struct{}

func (LiteTeamPolicy) MemberGuidance() string {
	return ""
}

// MemberRequestConfig stub - team member requests removed
type MemberRequestConfig struct {
	Enabled bool
}

// ParseMemberRequestConfig stub - returns default config
func ParseMemberRequestConfig(settings interface{}) *MemberRequestConfig {
	return &MemberRequestConfig{Enabled: false}
}

// TaskActionFlags stub - task action flags removed
type TaskActionFlags struct{}

// HeartbeatTool stub - heartbeat tool removed
type HeartbeatTool struct{}
