package methods

import "github.com/nextlevelbuilder/goclaw/internal/hooks"

// HookMethods stub - hooks methods removed
type HookMethods struct{}

// NewHookMethods stub - hooks methods constructor removed
func NewHookMethods(hs hooks.HookStore, edition interface{}) *HookMethods {
	return &HookMethods{}
}

// SetTestRunner stub - set test runner removed
func (hm *HookMethods) SetTestRunner(runner interface{}) {
	// no-op
}

// Register stub - register hooks methods removed
func (hm *HookMethods) Register(router interface{}) {
	// no-op
}

// DispatcherTestRunner stub - dispatcher test runner removed
type DispatcherTestRunner struct{}

// NewDispatcherTestRunner stub - test runner constructor removed
func NewDispatcherTestRunner(handlers interface{}) *DispatcherTestRunner {
	return &DispatcherTestRunner{}
}

// QuotaMethods stub - quota methods removed
type QuotaMethods struct{}

// NewQuotaMethods stub - quota methods constructor removed
func NewQuotaMethods(quotaChecker interface{}, db interface{}) *QuotaMethods {
	return &QuotaMethods{}
}

// Register stub - register quota methods removed
func (qm *QuotaMethods) Register(router interface{}) {
	// no-op
}
