package consolidation

// Stub package - consolidation removed

// ConsolidationDeps stub
type ConsolidationDeps struct {
	EpisodicStore   interface{}
	MemoryStore     interface{}
	KGStore         interface{}
	SessionStore    interface{}
	EventBus        interface{}
	SystemConfigs   interface{}
	Registry        interface{}
	Extractor       interface{}
	AlertDeps       interface{}
	AgentStore      interface{}
}

// Register stub
func Register(deps ConsolidationDeps) func() {
	return func() {}
}
