// Package sandbox - removed module stub
package sandbox

import "fmt"

// Stub package - sandbox/secure CLI module has been removed

// Config stub types
type Config struct {
	Mode               int
	Image              string
	WorkspaceAccess    int
	Scope              int
	MemoryMB           int
	CPUs               float64
	TimeoutSec         int
	NetworkEnabled     bool
	ReadOnlyRoot       bool
	SetupCommand       string
	Env                map[string]string
	User               string
	TmpfsSizeMB        int
	MaxOutputBytes     int
	IdleHours          int
	MaxAgeDays         int
	PruneIntervalMin   int
	ContainerWorkdir   string
}

// Mode constants
const (
	ModeAll int = iota
	ModeNonMain
	ModeOff
)

// Access constants
const (
	AccessNone int = iota
	AccessRO
	AccessRW
)

// Scope constants
const (
	ScopeAgent int = iota
	ScopeShared
	ScopeSession
)

// DefaultConfig returns default sandbox config (stub)
func DefaultConfig() Config {
	return Config{Mode: ModeOff}
}

// ExecResult holds the result of a sandbox exec
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Manager stub interface
type Manager interface {
	Exec(ctx interface{}, cmd []string, workdir string) (*ExecResult, error)
	Close() error
	ID() string
	Get(ctx interface{}, key, workspace string, cfg *Config) (Manager, error)
}

// StubManager is a stub implementation of Manager
type StubManager struct{}

func (m *StubManager) Exec(ctx interface{}, cmd []string, workdir string) (*ExecResult, error) {
	return &ExecResult{}, nil
}

func (m *StubManager) Close() error {
	return nil
}

func (m *StubManager) ID() string {
	return "stub"
}

func (m *StubManager) Get(ctx interface{}, key, workspace string, cfg *Config) (Manager, error) {
	return &StubManager{}, nil
}

// ErrSandboxDisabled is returned when sandbox is not available
var ErrSandboxDisabled = fmt.Errorf("sandbox disabled")

// FsBridge stub type - sandbox removed
type FsBridge struct{}

// NewFsBridge stub - returns empty bridge
func NewFsBridge(id, workdir string) *FsBridge {
	return &FsBridge{}
}

// ReadFile stub method
func (b *FsBridge) ReadFile(ctx interface{}, path string) (string, error) {
	return "", nil
}

// WriteFile stub method (4th param is fsync flag)
func (b *FsBridge) WriteFile(ctx interface{}, path, content string, fsync bool) error {
	return nil
}

// ListDir stub method
func (b *FsBridge) ListDir(ctx interface{}, path string) ([]string, error) {
	return nil, nil
}

// DefaultContainerWorkdir is the default working directory inside the container
const DefaultContainerWorkdir = "/workspace"
