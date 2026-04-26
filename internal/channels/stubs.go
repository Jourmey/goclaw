package channels

// GroupMember stub - channels removed
type GroupMember struct {
	UserID      string
	DisplayName string
}

// WriterLabel stub - removed
func WriterLabel(metadata interface{}, userID string) string {
	return "writer"
}

// MemberResolver stub - member resolution removed
type MemberResolver interface {
	ResolveMember(ctx interface{}, id string) (*GroupMember, error)
}

// Manager stub - channels subsystem removed
type Manager struct{}

// NewManager stub - channels manager constructor removed
func NewManager(msgBus interface{}) *Manager {
	return &Manager{}
}

// GetEnabledChannels stub
func (m *Manager) GetEnabledChannels() []string {
	return []string{}
}

// GetChannel stub
func (m *Manager) GetChannel(name string) interface{} {
	return nil
}

// GetStatus stub (can be called with or without parameters)
func (m *Manager) GetStatus() map[string]bool {
	return make(map[string]bool)
}

// SendToChannel stub - implements ChannelSender
func (m *Manager) SendToChannel(ctx interface{}, channel, chatID string, payload interface{}) error {
	return nil
}

// ChannelTenantID stub - returns (tenantID, exists)
func (m *Manager) ChannelTenantID(channel string) (interface{}, bool) {
	return nil, false
}

// ListGroupMembers stub
func (m *Manager) ListGroupMembers(ctx interface{}, channel, groupID string) ([]GroupMember, error) {
	return nil, nil
}

// ResolveMember stub - implements MemberResolver
func (m *Manager) ResolveMember(ctx interface{}, id string) (*GroupMember, error) {
	return nil, nil
}

// StartAll stub
func (m *Manager) StartAll(ctx interface{}) error {
	return nil
}

// ChannelTypeForName stub
func (m *Manager) ChannelTypeForName(name string) string {
	return ""
}

// ResolveBlockReply stub
func (m *Manager) ResolveBlockReply(ctx interface{}, channel, threadID string) (interface{}, error) {
	return nil, nil
}

// RegisterRun stub
func (m *Manager) RegisterRun(ctx interface{}, channel, chatID, traceID, runID string) error {
	return nil
}

// InstanceLoader stub - instance loader removed
type InstanceLoader struct{}

// NewInstanceLoader stub - instance loader constructor removed
func NewInstanceLoader(channelStore interface{}, agentStore interface{}, mgr *Manager, msgBus interface{}, pairingStore interface{}) *InstanceLoader {
	return &InstanceLoader{}
}

// SetProviderRegistry stub
func (il *InstanceLoader) SetProviderRegistry(reg interface{}) {
	// no-op
}

// SetPendingCompactionConfig stub
func (il *InstanceLoader) SetPendingCompactionConfig(cfg interface{}) {
	// no-op
}

// RegisterFactory stub
func (il *InstanceLoader) RegisterFactory(typename string, factory interface{}) {
	// no-op
}

// LoadAll stub
func (il *InstanceLoader) LoadAll(ctx interface{}) error {
	return nil
}

// Channel type constants
const (
	TypeTelegram      = "telegram"
	TypeDiscord       = "discord"
	TypeFeishu        = "feishu"
	TypeZaloOA        = "zalo_oa"
	TypeZaloPersonal  = "zalo_personal"
	TypeWhatsApp      = "whatsapp"
	TypeSlack         = "slack"
	TypeFacebook      = "facebook"
	TypePancake       = "pancake"
)

// CopyFinalRoutingMeta stub function - channels routing removed
func CopyFinalRoutingMeta(metadata interface{}) interface{} {
	return nil
}

// QuotaChecker stub - quota checking removed
type QuotaChecker struct{}

// NewQuotaChecker stub - quota checker constructor removed
func NewQuotaChecker(db interface{}, cfg interface{}) *QuotaChecker {
	return &QuotaChecker{}
}

// Stop stub
func (qc *QuotaChecker) Stop() {
	// no-op
}

// CheckQuota stub
func (qc *QuotaChecker) CheckQuota(clientID string, limit int) bool {
	return false
}

// QuotaUsageResult stub - quota usage removed
type QuotaUsageResult struct{}

// QuotaUsageEntry stub - quota usage entry removed
type QuotaUsageEntry struct{}

// QuotaResult stub - quota result removed
type QuotaResult struct{}

// QueryTodaySummary stub - query today summary removed
func QueryTodaySummary(ctx interface{}, channel string) interface{} {
	return nil
}
