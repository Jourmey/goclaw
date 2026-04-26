package http

import "net/http"

// KnowledgeGraphHandler stub - knowledge graph removed
type KnowledgeGraphHandler struct {
	store interface{}
}

// NewKnowledgeGraphHandler stub
func NewKnowledgeGraphHandler(store interface{}) *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{store: store}
}

// RegisterRoutes stub
func (h *KnowledgeGraphHandler) RegisterRoutes(mux *http.ServeMux) {}
