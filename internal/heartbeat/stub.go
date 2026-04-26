// Package heartbeat - removed module stub
package heartbeat

import (
	"context"

	"github.com/nextlevelbuilder/goclaw/internal/agent"
	"github.com/nextlevelbuilder/goclaw/internal/scheduler"
	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// Stub package - heartbeat module has been removed

// TickerConfig stub
type TickerConfig struct {
	Store         store.HeartbeatStore
	Agents        store.AgentStore
	Sessions      store.SessionStore
	ProviderStore store.ProviderStore
	ProviderReg   ProviderResolver
	MsgBus        EventPublisher
	Sched         ActiveSessionChecker
	RunAgent      func(ctx context.Context, req agent.RunRequest) <-chan scheduler.RunOutcome
}

// Ticker stub
type Ticker struct{}

// NewTicker stub - returns empty ticker
func NewTicker(cfg TickerConfig) *Ticker {
	return &Ticker{}
}

// Start stub - no-op
func (t *Ticker) Start(ctx context.Context) error {
	return nil
}

// Stop stub - no-op
func (t *Ticker) Stop() error {
	return nil
}

// ProviderResolver stub interface
type ProviderResolver interface {
	Resolve(ctx interface{}, providerID interface{}) (interface{}, error)
}

// EventPublisher stub interface
type EventPublisher interface {
	Publish(ctx interface{}, event interface{}) error
}

// ActiveSessionChecker stub interface
type ActiveSessionChecker interface {
	IsActive(sessionKey string) bool
}
