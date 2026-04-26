package pg

import (
	"github.com/google/uuid"
)

// Stub package - removed exports for import functionality

// CronJobExport stub - cron jobs import removed
type CronJobExport struct {
	ID uuid.UUID
}

// UserProfileExport stub - user profiles import removed
type UserProfileExport struct {
	UserID string
}

// UserOverrideExport stub - user overrides import removed
type UserOverrideExport struct {
	ID uuid.UUID
}

// EpisodicSummaryExport stub - episodic summaries import removed
type EpisodicSummaryExport struct {
	ID uuid.UUID
}

// EvolutionMetricExport stub - evolution metrics import removed
type EvolutionMetricExport struct {
	ID uuid.UUID
}

// EvolutionSuggestionExport stub - evolution suggestions import removed
type EvolutionSuggestionExport struct {
	ID uuid.UUID
}

// VaultDocumentExport stub - vault documents import removed
type VaultDocumentExport struct {
	ID uuid.UUID
}

// VaultLinkExport stub - vault links import removed
type VaultLinkExport struct {
	ID uuid.UUID
}

// TeamExport stub - team export import removed
type TeamExport struct {
	ID uuid.UUID
}

// TeamMemberExport stub - team member export import removed
type TeamMemberExport struct {
	ID uuid.UUID
}

// TeamTaskExport stub - team task export import removed
type TeamTaskExport struct {
	ID uuid.UUID
}

// TeamTaskCommentExport stub - team task comment export import removed
type TeamTaskCommentExport struct {
	ID uuid.UUID
}

// TeamTaskEventExport stub - team task event export import removed
type TeamTaskEventExport struct {
	ID uuid.UUID
}

// AgentLinkExport stub - agent link export import removed
type AgentLinkExport struct {
	ID uuid.UUID
}

// ExportAgentContextFiles stub
func ExportAgentContextFiles(ctx interface{}, db interface{}, agentID interface{}) ([]interface{}, error) {
	return nil, nil
}

// ExportUserContextFiles stub
func ExportUserContextFiles(ctx interface{}, db interface{}, userID interface{}) ([]interface{}, error) {
	return nil, nil
}

// ExportMemoryDocuments stub
func ExportMemoryDocuments(ctx interface{}, db interface{}, userID interface{}) ([]interface{}, error) {
	return nil, nil
}

// ExportKGEntities stub
func ExportKGEntities(ctx interface{}, db interface{}, agentID interface{}) ([]interface{}, error) {
	return nil, nil
}

// ExportKGRelations stub
func ExportKGRelations(ctx interface{}, db interface{}, agentID interface{}) ([]interface{}, error) {
	return nil, nil
}

// ExportCronJobs stub
func ExportCronJobs(ctx interface{}, db interface{}, agentID interface{}) ([]CronJobExport, error) {
	return nil, nil
}

// ExportUserProfiles stub
func ExportUserProfiles(ctx interface{}, db interface{}, agentID interface{}) ([]UserProfileExport, error) {
	return nil, nil
}

// ExportUserOverrides stub
func ExportUserOverrides(ctx interface{}, db interface{}, userID interface{}) ([]UserOverrideExport, error) {
	return nil, nil
}

// ExportTeamByLead stub
func ExportTeamByLead(ctx interface{}, db interface{}, leadID interface{}) (*TeamExport, error) {
	return nil, nil
}

// ExportTeamTasks stub
func ExportTeamTasks(ctx interface{}, db interface{}, teamID interface{}) ([]TeamTaskExport, error) {
	return nil, nil
}

// TeamTasksExport stub
type TeamTasksExport struct {
	ID uuid.UUID
}

// ExportTeamComments stub
func ExportTeamComments(ctx interface{}, db interface{}, teamID interface{}) ([]TeamTaskCommentExport, error) {
	return nil, nil
}

// ExportTeamEvents stub
func ExportTeamEvents(ctx interface{}, db interface{}, teamID interface{}) ([]TeamTaskEventExport, error) {
	return nil, nil
}

// ExportAgentLinks stub
func ExportAgentLinks(ctx interface{}, db interface{}, sourceID interface{}) ([]AgentLinkExport, error) {
	return nil, nil
}

// ExportPreviewCounts stub
type ExportPreviewCounts struct{}

// ExportPreview stub
type ExportPreview struct{}

// GetExportPreviewCounts stub
func GetExportPreviewCounts(ctx interface{}, db interface{}, agentID interface{}) (*ExportPreviewCounts, error) {
	return nil, nil
}
