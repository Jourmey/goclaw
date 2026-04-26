// Package mcp - removed module stub
package mcp

import (
	"context"

	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

// Stub package - MCP module has been removed

// Pool stub - MCP connection pool removed
type Pool struct{}

// PoolEntry stub for MCP pool entries
type PoolEntry struct{}

func (e *PoolEntry) MCPTools() []interface{} {
	return nil
}

func (e *PoolEntry) ClientPtr() interface{} {
	return nil
}

func (e *PoolEntry) Connected() bool {
	return false
}

func (p *Pool) AcquireUser(ctx interface{}, tenantID interface{}, serverName, userID, transport, command string, args []string, env map[string]string, url string, headers map[string]string, timeout int) (*PoolEntry, error) {
	return &PoolEntry{}, nil
}

func (p *Pool) ReleaseUser(key UserPoolKey) {
	// no-op
}

// GrantChecker stub - MCP grant checker removed
type GrantChecker interface {
	Check(ctx interface{}, userID, serverName, toolName string) (bool, error)
}

// UserPoolKey stub
type UserPoolKey struct {
	TenantID   interface{}
	ServerName string
	UserID     string
}

// ParseJSONBytesToStringSlice stub
func ParseJSONBytesToStringSlice(data []byte) []string {
	return nil
}

// ParseJSONBytesToStringMap stub
func ParseJSONBytesToStringMap(data []byte) map[string]string {
	return nil
}

// BridgeTool stub
type BridgeTool struct{}

func (bt *BridgeTool) Name() string {
	return ""
}

func (bt *BridgeTool) Description() string {
	return ""
}

func (bt *BridgeTool) Execute(ctx context.Context, args map[string]any) *tools.Result {
	return nil
}

func (bt *BridgeTool) Parameters() map[string]any {
	return nil
}

// NewBridgeTool stub
func NewBridgeTool(serverName string, mcpTool, clientPtr interface{}, toolPrefix string, timeout int, connectedFunc interface{}, serverID interface{}, grantChecker GrantChecker) *BridgeTool {
	return &BridgeTool{}
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
func (m *Manager) LoadForAgent(ctx interface{}, agentID interface{}, filter string) error {
	return nil
}

// UserCredServers stub
func (m *Manager) UserCredServers() []interface{} {
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

// DefaultPoolConfig stub
func DefaultPoolConfig() interface{} {
	return nil
}

// NewPool stub
func NewPool(config interface{}) *Pool {
	return &Pool{}
}

// NewStoreGrantChecker stub
func NewStoreGrantChecker(store interface{}, msgBus interface{}) GrantChecker {
	return nil
}

// BridgeServer stub
type BridgeServer struct{}

// NewBridgeServer stub
func NewBridgeServer(toolReg interface{}, version string, msgBus interface{}) *BridgeServer {
	return &BridgeServer{}
}

// ServeHTTP stub
func (bs *BridgeServer) ServeHTTP(w interface{}, r interface{}) {
	// no-op
}
