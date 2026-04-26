// Package hooks - removed module stub
package hooks

import (
	"context"

	"github.com/google/uuid"
)

// Stub package - hooks module has been removed

// Dispatcher stub interface
type Dispatcher interface {
	Fire(ctx context.Context, event Event) FireResult
}

// Event stub type
type Event struct {
	Name       string
	Data       map[string]interface{}
	EventID    string
	SessionID  string
	TenantID   string
	AgentID    string
	RawInput   string
	HookEvent  string
}

// FireResult stub type
type FireResult struct {
	Blocked          bool
	Message          string
	Decision         Decision
	UpdatedRawInput  string
}

// Decision constants
type Decision int

const (
	DecisionAllow Decision = iota
	DecisionBlock
)

// Event name constants
const (
	EventSessionStart       = "session.start"
	EventUserPromptSubmit   = "user.prompt.submit"
	EventAssistantTurnStart = "assistant.turn.start"
)

// HookStore stub - hooks store removed
type HookStore interface {
	ListHooks(ctx context.Context) ([]HookConfig, error)
	GetHook(ctx context.Context, id interface{}) (*HookConfig, error)
	CreateHook(ctx context.Context, cfg *HookConfig) error
	UpdateHook(ctx context.Context, cfg *HookConfig) error
	DeleteHook(ctx context.Context, id interface{}) error
	List(ctx context.Context, filter *ListFilter) ([]HookConfig, error)
}

// HookConfig stub - hook configuration removed
type HookConfig struct {
	ID       interface{}
	Name     string
	Scope    Scope
	TenantID uuid.UUID
}

// ListFilter stub - list filter removed
type ListFilter struct {
	Enabled bool
	Event   *HookEvent
	Scope   *Scope
	AgentID *uuid.UUID
}

// HookEvent stub - hook event removed
type HookEvent string

// Scope stub - scope removed
type Scope string

const (
	ScopeGlobal Scope = "global"
)

var (
	SentinelTenantID = uuid.UUID{} // sentinel value for missing tenant ID
)

// DecisionDeny stub constant
const DecisionDeny Decision = 0

// HandlerType stub - handler type removed
type HandlerType string

// Handler stub - hook handler removed
type Handler struct {
	Type string
}
