package http

import "net/http"

// TracesHandler stub - tracing removed
type TracesHandler struct{}

// NewTracesHandler creates a stub traces handler
func NewTracesHandler(tracingStore interface{}) *TracesHandler {
	return &TracesHandler{}
}

// RegisterRoutes registers stub routes
func (h *TracesHandler) RegisterRoutes(mux *http.ServeMux) {}

// MCPHandler stub - MCP removed
type MCPHandler struct{}

// NewMCPHandler creates a stub MCP handler
func NewMCPHandler(mcpStore interface{}, msgBus interface{}, mcpToolLister MCPToolLister) *MCPHandler {
	return &MCPHandler{}
}

// SetPoolEvictor stub
func (h *MCPHandler) SetPoolEvictor(pool interface{}) {}

// SetDB stub
func (h *MCPHandler) SetDB(db interface{}) {}

// RegisterRoutes stub
func (h *MCPHandler) RegisterRoutes(mux *http.ServeMux) {}

// MCPUserCredentialsHandler stub - MCP removed
type MCPUserCredentialsHandler struct{}

// NewMCPUserCredentialsHandler creates a stub MCP user credentials handler
func NewMCPUserCredentialsHandler(mcpStore interface{}, tenantStore interface{}) *MCPUserCredentialsHandler {
	return &MCPUserCredentialsHandler{}
}

// RegisterRoutes stub
func (h *MCPUserCredentialsHandler) RegisterRoutes(mux *http.ServeMux) {}

// ChannelInstancesHandler stub - channel instances removed
type ChannelInstancesHandler struct{}

// NewChannelInstancesHandler creates a stub channel instances handler
func NewChannelInstancesHandler(channelStore interface{}, agentStore interface{}, configStore interface{}, contactStore interface{}, tenantStore interface{}, msgBus interface{}) *ChannelInstancesHandler {
	return &ChannelInstancesHandler{}
}

// SetMemberResolver stub
func (h *ChannelInstancesHandler) SetMemberResolver(resolver interface{}) {
	// no-op
}

// RegisterRoutes stub
func (h *ChannelInstancesHandler) RegisterRoutes(mux *http.ServeMux) {}

// PendingMessagesHandler stub - pending messages removed
type PendingMessagesHandler struct{}

// NewPendingMessagesHandler creates a stub pending messages handler
func NewPendingMessagesHandler(pendingMessageStore interface{}, agentStore interface{}, providerReg interface{}) *PendingMessagesHandler {
	return &PendingMessagesHandler{}
}

// RegisterRoutes stub
func (h *PendingMessagesHandler) RegisterRoutes(mux *http.ServeMux) {}

// OAuthHandler stub - OAuth removed
type OAuthHandler struct{}

// NewOAuthHandler creates a stub OAuth handler
func NewOAuthHandler(providerStore interface{}, configSecretsStore interface{}, providerRegistry interface{}, msgBus interface{}) *OAuthHandler {
	return &OAuthHandler{}
}

// RegisterRoutes stub
func (h *OAuthHandler) RegisterRoutes(mux *http.ServeMux) {}

// MediaUploadHandler stub - media upload removed
type MediaUploadHandler struct{}

// NewMediaUploadHandler creates a stub media upload handler
func NewMediaUploadHandler() *MediaUploadHandler {
	return &MediaUploadHandler{}
}

// RegisterRoutes stub
func (h *MediaUploadHandler) RegisterRoutes(mux *http.ServeMux) {}

// VaultHandler stub - vault removed
type VaultHandler struct{}

// NewVaultHandler creates a stub vault handler
func NewVaultHandler() *VaultHandler {
	return &VaultHandler{}
}

// RegisterRoutes stub
func (h *VaultHandler) RegisterRoutes(mux *http.ServeMux) {}

// VaultGraphHandler stub - vault graph removed
type VaultGraphHandler struct{}

// NewVaultGraphHandler creates a stub vault graph handler
func NewVaultGraphHandler() *VaultGraphHandler {
	return &VaultGraphHandler{}
}

// RegisterRoutes stub
func (h *VaultGraphHandler) RegisterRoutes(mux *http.ServeMux) {}

// MCPToolLister is an interface for listing MCP tools
type MCPToolLister interface {
	ListTools(ctx interface{}) (interface{}, error)
}
