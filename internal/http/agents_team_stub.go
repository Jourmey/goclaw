package http

import "net/http"

// handleTeamExportPreview stub - team export removed
func (h *AgentsHandler) handleTeamExportPreview(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "team export removed"})
}

// handleTeamExport stub - team export removed
func (h *AgentsHandler) handleTeamExport(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "team export removed"})
}

// handleTeamImport stub - team import removed
func (h *AgentsHandler) handleTeamImport(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "team import removed"})
}
