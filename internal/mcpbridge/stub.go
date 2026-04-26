// Package mcpbridge - removed module stub
package mcpbridge

import (
	"context"

	"github.com/google/uuid"
	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// Stub package - MCP bridge module has been removed

// Pool stub - MCP connection pool removed
type Pool struct{}

// GrantChecker stub - MCP grant checker removed
type GrantChecker struct{}

// Check stub method
func (gc *GrantChecker) Check(ctx interface{}, userID, serverName, toolName string) (bool, error) {
	return false, nil
}

// ManagerOption stub
type ManagerOption func(*Manager)

// WithStore stub
func WithStore(s interface{}) ManagerOption {
	return func(m *Manager) {}
}

// WithPool stub
func WithPool(p *Pool) ManagerOption {
	return func(m *Manager) {}
}

// WithGrantChecker stub
func WithGrantChecker(gc *GrantChecker) ManagerOption {
	return func(m *Manager) {}
}

// Manager stub
type Manager struct{}

// NewManager stub
func NewManager(toolsReg interface{}, opts ...ManagerOption) *Manager {
	return &Manager{}
}

// LoadForAgent stub
func (m *Manager) LoadForAgent(ctx context.Context, agentID uuid.UUID, filter string) error {
	return nil
}

// UserCredServers stub
func (m *Manager) UserCredServers() []store.MCPAccessInfo {
	return nil
}

// IsSearchMode stub
func (m *Manager) IsSearchMode() bool {
	return false
}

// ActivateToolIfDeferred stub
func (m *Manager) ActivateToolIfDeferred(toolName string) bool {
	return false
}

// DeferredToolInfos stub
func (m *Manager) DeferredToolInfos() []interface{} {
	return nil
}

// ToolNames stub
func (m *Manager) ToolNames() []string {
	return nil
}

// NewMCPToolSearchTool stub
func NewMCPToolSearchTool(mgr *Manager) interface{} {
	return nil
}
