package pg

import (
	"time"

	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// documentInfoRow is a scan target for SELECT agent_id, path, hash, user_id, updated_at FROM memory_documents.
type documentInfoRow struct {
	AgentID   string     `db:"agent_id"`
	Path      string     `db:"path"`
	Hash      string     `db:"hash"`
	UserID    *string    `db:"user_id"`
	UpdatedAt *time.Time `db:"updated_at"`
}

func (r documentInfoRow) toDocumentInfo() store.DocumentInfo {
	info := store.DocumentInfo{
		Path:    r.Path,
		Hash:    r.Hash,
		AgentID: r.AgentID,
	}
	if r.UserID != nil {
		info.UserID = *r.UserID
	}
	if r.UpdatedAt != nil {
		info.UpdatedAt = r.UpdatedAt.Unix()
	}
	return info
}

// documentDetailRow is a scan target for GetDocument queries.
type documentDetailRow struct {
	Path          string     `db:"path"`
	Content       string     `db:"content"`
	Hash          string     `db:"hash"`
	UserID        *string    `db:"user_id"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	ChunkCount    int        `db:"chunk_count"`
	EmbeddedCount int        `db:"embedded_count"`
}

func (r documentDetailRow) toDocumentDetail() store.DocumentDetail {
	detail := store.DocumentDetail{
		Path:          r.Path,
		Content:       r.Content,
		Hash:          r.Hash,
		ChunkCount:    r.ChunkCount,
		EmbeddedCount: r.EmbeddedCount,
	}
	if r.UserID != nil {
		detail.UserID = *r.UserID
	}
	if r.CreatedAt != nil {
		detail.CreatedAt = r.CreatedAt.Unix()
	}
	if r.UpdatedAt != nil {
		detail.UpdatedAt = r.UpdatedAt.Unix()
	}
	return detail
}

// chunkInfoRow is a scan target for ListChunks queries.
type chunkInfoRow struct {
	ID           string `db:"id"`
	StartLine    int    `db:"start_line"`
	EndLine      int    `db:"end_line"`
	TextPreview  string `db:"text_preview"`
	HasEmbedding bool   `db:"has_embedding"`
}

func (r chunkInfoRow) toChunkInfo() store.ChunkInfo {
	return store.ChunkInfo{
		ID:           r.ID,
		StartLine:    r.StartLine,
		EndLine:      r.EndLine,
		TextPreview:  r.TextPreview,
		HasEmbedding: r.HasEmbedding,
	}
}

// scoredChunkRow is a scan target for FTS/vector search queries.
type scoredChunkRow struct {
	Path      string   `db:"path"`
	StartLine int      `db:"start_line"`
	EndLine   int      `db:"end_line"`
	Text      string   `db:"text"`
	UserID    *string  `db:"user_id"`
	Score     float64  `db:"score"`
}

func (r scoredChunkRow) toScoredChunk() scoredChunk {
	return scoredChunk{
		Path:      r.Path,
		StartLine: r.StartLine,
		EndLine:   r.EndLine,
		Text:      r.Text,
		Score:     r.Score,
		UserID:    r.UserID,
	}
}
