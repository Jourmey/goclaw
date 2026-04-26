package vault

// Stub package - vault subsystem removed

// Store stub - vault store removed
type Store struct{}

// LoadDocuments stub
func (s *Store) LoadDocuments() error {
	return nil
}

// LockDocuments stub
func (s *Store) LockDocuments() {
	// no-op
}

// UnlockDocuments stub
func (s *Store) UnlockDocuments() {
	// no-op
}

// EnrichProgress stub - vault enrichment progress removed
type EnrichProgress struct{}

// EnrichWorker stub - vault enrichment worker removed
type EnrichWorker struct{}

// Stop stub
func (ew *EnrichWorker) Stop() {
	// no-op
}

// Enqueue stub
func (ew *EnrichWorker) Enqueue(id interface{}) {
	// no-op
}

// EnrichWorkerDeps stub
type EnrichWorkerDeps struct {
	VaultStore    interface{}
	SystemConfigs interface{}
	Registry      interface{}
	EventBus      interface{}
	MsgBus        interface{}
	TeamStore     interface{}
	AlertDeps     interface{}
}

// RegisterEnrichWorker stub
func RegisterEnrichWorker(deps EnrichWorkerDeps) (func(), *EnrichProgress, *EnrichWorker) {
	return func() {}, &EnrichProgress{}, &EnrichWorker{}
}
