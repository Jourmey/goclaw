package store

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TeamStore stub - teams removed
type TeamStore interface {
	GetTeam(ctx interface{}, teamID uuid.UUID) (*TeamData, error)
	GetTeamForAgent(ctx interface{}, agentID uuid.UUID) (*TeamData, error)
	GetTask(ctx interface{}, taskID uuid.UUID) (*TeamTaskData, error)
	ForceRecoverAllTasks(ctx interface{}) ([]RecoveredTaskInfo, error)
	RecoverAllStaleTasks(ctx interface{}) ([]RecoveredTaskInfo, error)
	MarkAllStaleTasks(ctx interface{}, time interface{}) ([]RecoveredTaskInfo, error)
	MarkInReviewStaleTasks(ctx interface{}, time interface{}) ([]RecoveredTaskInfo, error)
	FixOrphanedBlockedTasks(ctx interface{}) ([]RecoveredTaskInfo, error)
	ListMembers(ctx interface{}, teamID uuid.UUID) ([]TeamMemberData, error)
	ListAllFollowupDueTasks(ctx interface{}) ([]TeamTaskData, error)
	IncrementFollowupCount(ctx interface{}, taskID uuid.UUID, time interface{}) error
	ListTasks(ctx interface{}, teamID uuid.UUID, sort, statusFilter, userFilter, labelFilter, searchQuery string, offset, limit int) ([]TeamTaskData, error)
	GetAttachment(ctx interface{}, attachmentID uuid.UUID) (*TeamAttachmentData, error)
	ListTeamEvents(ctx interface{}, teamID uuid.UUID, limit, offset int) ([]TeamTaskEventData, error)
	HasTeamAccess(ctx interface{}, teamID uuid.UUID, userID string) (bool, error)
	HasActiveMemberTasks(ctx interface{}, teamID uuid.UUID, memberID interface{}) (bool, error)
	SetFollowupForActiveTasks(ctx interface{}, teamID uuid.UUID, channel, chatID string, dueAt interface{}, count int, msg string) (int64, error)
	ListRecentTaskComments(ctx interface{}, taskID uuid.UUID, limit int) ([]interface{}, error)
	ListTaskAttachments(ctx interface{}, taskID uuid.UUID) ([]interface{}, error)
}

// TracingStore stub - tracing removed
type TracingStore interface {
	SaveSpan(ctx interface{}, span interface{}) error
	ListCodexPoolSpans(ctx interface{}, agentID uuid.UUID, startTime uuid.UUID, providers []string, limit int) ([]CodexPoolSpan, error)
	ListCodexPoolSpansByProviders(ctx interface{}, tenantID uuid.UUID, providers map[string]interface{}, limit int) ([]interface{}, error)
}

// AgentLink stub - agent links removed
type AgentLink struct {
	ID                uuid.UUID
	SourceAgentID     uuid.UUID
	SourceAgentKey    string
	TargetAgentID     uuid.UUID
	TargetAgentKey    string
	TargetDisplayName string
	Description       string
	Direction         string
	Status            string
}

// AgentLinkData stub - agent links removed
type AgentLinkData struct {
	ID            uuid.UUID
	TeamID        *uuid.UUID
	SourceAgentID uuid.UUID
	TargetAgentID uuid.UUID
	Direction     string
	Description   string
	MaxConcurrent int
	Settings      json.RawMessage
	Status        string
	CreatedBy     string
}

// AgentLinkStore stub - agent links removed
type AgentLinkStore interface {
	GetLink(ctx interface{}, linkID uuid.UUID) (*AgentLink, error)
	DelegateTargets(ctx interface{}, agentID uuid.UUID) ([]*AgentLink, error)
	ListLinksTo(ctx interface{}, targetID uuid.UUID) ([]AgentLinkData, error)
	ListLinksFrom(ctx interface{}, sourceID uuid.UUID) ([]AgentLinkData, error)
	CreateLink(ctx interface{}, link *AgentLinkData) error
	UpdateLink(ctx interface{}, linkID uuid.UUID, updates map[string]interface{}) error
	DeleteLink(ctx interface{}, linkID uuid.UUID) error
}

// LinkDirection stub - link direction removed
type LinkDirection string

const (
	LinkDirectionOutbound      = "outbound"
	LinkDirectionInbound       = "inbound"
	LinkDirectionBidirectional = "bidirectional"
	LinkStatusActive           = "active"
	LinkStatusDisabled         = "disabled"
)

// RecoveredTaskInfo stub - tasks removed
type RecoveredTaskInfo struct {
	ID         uuid.UUID
	TaskID     uuid.UUID
	TeamID     uuid.UUID
	DueAt      int64
	TaskData   interface{}
	TenantID   uuid.UUID
	Channel    string
	ChatID     string
	TaskNumber int
	Subject    string
}

// TeamData stub - teams removed
type TeamData struct {
	ID           uuid.UUID
	Name         string
	Description  string
	LeadAgentID  uuid.UUID
	Settings     json.RawMessage
	CreatedBy    string
}

// TeamMemberData stub - team members removed
type TeamMemberData struct {
	ID           uuid.UUID
	TeamID       uuid.UUID
	UserID       string
	AgentID      uuid.UUID
	AgentKey     string
	DisplayName  string
	Role         string
	Frontmatter  string
}

// TeamTaskData stub - team tasks removed
type TeamTaskData struct {
	ID              uuid.UUID
	TeamID          uuid.UUID
	Title           string
	Metadata        map[string]interface{}
	TenantID        uuid.UUID
	FollowupChannel string
	FollowupChatID  string
	FollowupMessage string
	FollowupCount   int
	FollowupMax     int
	TaskNumber      int
	Status          string
	Subject         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ProgressPercent int
	ProgressStep    string
}

// TeamTaskEventData stub - team task events removed
type TeamTaskEventData struct {
	ID        uuid.UUID
	TeamID    uuid.UUID
	TaskID    uuid.UUID
	Type      string
	Timestamp time.Time
}

// TeamAttachmentData stub - team attachments removed
type TeamAttachmentData struct {
	ID       uuid.UUID
	TeamID   uuid.UUID
	ChatID   string
	Path     string
	MimeType string
}

// TeamTaskStatusPending stub constant
const TeamTaskStatusPending = "pending"

// TeamTaskStatusInProgress stub constant
const TeamTaskStatusInProgress = "in_progress"

// TeamTaskStatusStale stub constant
const TeamTaskStatusStale = "stale"

// TeamTaskStatusCompleted stub constant
const TeamTaskStatusCompleted = "completed"

// TeamTaskStatusCancelled stub constant
const TeamTaskStatusCancelled = "cancelled"

// TeamTaskStatusFailed stub constant
const TeamTaskStatusFailed = "failed"

// TeamTaskFilterAll stub constant
const TeamTaskFilterAll = "all"

// TraceData stub - tracing removed
type TraceData struct {
	ID             uuid.UUID
	RunID          string
	TraceID        uuid.UUID
	SpanID         uuid.UUID
	SessionKey     string
	UserID         string
	Channel        string
	Name           string
	InputPreview   string
	Status         string
	StartTime      int64
	CreatedAt      int64
	AgentID        uuid.UUID
	ParentTraceID  *uuid.UUID
	TeamID         *uuid.UUID
	Tags           map[string]string
}

// TraceStatusRunning stub constant
const TraceStatusRunning = "running"

// TraceStatusError stub constant
const TraceStatusError = "error"

// TraceStatusCancelled stub constant
const TraceStatusCancelled = "cancelled"

// TraceStatusCompleted stub constant
const TraceStatusCompleted = "completed"

// SpanData stub - tracing spans removed
type SpanData struct {
	ID            uuid.UUID
	TraceID       uuid.UUID
	ParentID      *uuid.UUID
	ParentSpanID  *uuid.UUID
	Name          string
	SpanType      string
	Status        string
	Level         string
	TeamID        *uuid.UUID
	TenantID      *uuid.UUID
	AgentID       *uuid.UUID
	ModelInput    string
	InputPreview  string
	StartTime     int64
	Model         string
	Provider      string
	CreatedAt     int64
	ToolName      string
	ToolCallID    string
}

// SpanTypeLLMCall stub constant
const SpanTypeLLMCall = "llm_call"

// SpanStatusRunning stub constant
const SpanStatusRunning = "running"

// SpanLevelDefault stub constant
const SpanLevelDefault = "default"

// SpanStatusCompleted stub constant
const SpanStatusCompleted = "completed"

// SpanStatusError stub constant
const SpanStatusError = "error"

// SpanTypeToolCall stub constant
const SpanTypeToolCall = "tool_call"

// SpanTypeAgent stub constant
const SpanTypeAgent = "agent"

// TeamRoleLead stub constant - teams removed
const TeamRoleLead = "lead"

// TeamRoleMember stub constant - teams removed
const TeamRoleMember = "member"

// TeamRoleReviewer stub constant - teams removed
const TeamRoleReviewer = "reviewer"

// KnowledgeGraphStore stub - knowledge graph removed
type KnowledgeGraphStore interface {
	GetNode(ctx interface{}, nodeID interface{}) (interface{}, error)
}

// VaultStore stub - vault removed
type VaultStore interface {
	GetDocument(ctx interface{}, docID interface{}) (interface{}, error)
}

// HeartbeatStore stub - heartbeat removed
type HeartbeatStore interface {
	RecordHeartbeat(ctx interface{}, beat interface{}) error
	Get(ctx interface{}, agentID uuid.UUID) (*AgentHeartbeat, error)
	Upsert(ctx interface{}, hb *AgentHeartbeat) error
	ListLogs(ctx interface{}, agentID uuid.UUID, limit, offset int) ([]interface{}, int, error)
	ListDeliveryTargets(ctx interface{}, agentID uuid.UUID) ([]interface{}, error)
}

// StaggerOffset stub function - heartbeat stagger offset removed
func StaggerOffset(agentID uuid.UUID, intervalSec int) time.Duration {
	return time.Duration(0)
}

// HeartbeatEvent stub - heartbeat removed
type HeartbeatEvent struct {
	AgentID uuid.UUID
	Status  string
}

// AgentHeartbeat stub - heartbeat removed
type AgentHeartbeat struct {
	AgentID          uuid.UUID
	LastSeen         time.Time
	IntervalSec      int
	IsolatedSession  bool
	AckMaxChars      int
	MaxRetries       int
	Enabled          bool
	Prompt           *string
	ProviderID       *uuid.UUID
	Model            *string
	LightContext     bool
	ActiveHoursStart *string
	ActiveHoursEnd   *string
	Timezone         *string
	Channel          *string
	ChatID           *string
	NextRunAt        *time.Time
}

// PendingMessageStore stub - pending messages removed
type PendingMessageStore interface {
	GetPendingMessages(ctx interface{}) ([]interface{}, error)
	ListGroups(ctx interface{}) ([]interface{}, error)
	ResolveGroupTitles(ctx interface{}, ids []interface{}) (map[string]string, error)
	ListByKey(ctx interface{}, key interface{}) ([]interface{}, error)
}

// SnapshotStore stub - snapshots removed
type SnapshotStore interface {
	GetSnapshot(ctx interface{}, id interface{}) (interface{}, error)
	GetTimeSeries(ctx interface{}, q SnapshotQuery) ([]SnapshotTimeSeries, error)
	GetBreakdown(ctx interface{}, q SnapshotQuery) ([]SnapshotTimeSeries, error)
}

// CodexPoolSpan stub - Codex pool removed
type CodexPoolSpan struct {
	ID        uuid.UUID
	SpanID    uuid.UUID
	TraceID   uuid.UUID
	StartedAt time.Time
	Status    string
	Provider  string
	Model     string
	DurationMS int
	Metadata  json.RawMessage
}

// PairingStore stub - pairing removed
type PairingStore interface {
	GetPairing(ctx interface{}, id interface{}) (interface{}, error)
	IsPaired(ctx interface{}, arg1, arg2 interface{}) (bool, error)
	RequestPairing(ctx interface{}, clientID, pairingType, userID, role string, meta interface{}) (string, error)
}

// SnapshotQuery stub - usage snapshots removed
type SnapshotQuery struct {
	From     time.Time
	To       time.Time
	GroupBy  string
	AgentID  *uuid.UUID
	Provider string
	Model    string
	Channel  string
}

// SnapshotTimeSeries stub - usage time series removed
type SnapshotTimeSeries struct {
	Timestamp     time.Time
	Value         int
	BucketTime    time.Time
	RequestCount  int
	ErrorCount    int
	UniqueUsers   int
	InputTokens   int64
	OutputTokens  int64
	TotalCost     float64
	LLMCallCount  int
	ToolCallCount int
	AvgDurationMS float64
}

// CronEvent stub - cron events removed
type CronEvent struct {
	AgentID     uuid.UUID
	UserID      string
	EventType   string
	Timestamp   time.Time
}

// CronSchedule stub - cron schedule removed
type CronSchedule struct {
	Kind   string
	Expr   string
	EveryMS *int64
}

// CronState stub - cron state removed
type CronState struct {
	LastRunAtMS *int64
}

// CronJob stub - cron jobs removed
type CronJob struct {
	ID       uuid.UUID
	AgentID  uuid.UUID
	Spec     string
	NextRun  *time.Time
	Schedule CronSchedule
	State    CronState
	Name     string
	Enabled  bool
}

// CronStore stub - cron store removed
type CronStore interface {
	GetCronJob(ctx interface{}, jobID uuid.UUID) (*CronJob, error)
	ListCronJobs(ctx interface{}) ([]CronJob, error)
}
