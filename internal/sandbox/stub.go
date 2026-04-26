// Package sandbox - removed module stub
package sandbox

// Stub package - sandbox/secure CLI module has been removed

// Config stub types
type Config struct {
	Mode             int
	Image            string
	WorkspaceAccess  int
	Scope            int
	MemoryMB         int
	CPUs             float64
	TimeoutSec       int
	NetworkEnabled   bool
	ReadOnlyRoot     bool
	SetupCommand     string
	Env              map[string]string
	User             string
	TmpfsSizeMB      int
	MaxOutputBytes   int
	IdleHours        int
	MaxAgeDays       int
	PruneIntervalMin int
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
