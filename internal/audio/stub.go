package audio

import "context"

// Manager is a stub for TTS/Audio support.
// TTS was removed in v3.x. This stub exists only to maintain compatibility with channels
// that reference audio.Manager in their factory/constructor signatures.
// All audio.Manager instances are nil at runtime.
type Manager struct{}

func NewManager() *Manager {
	return nil
}

// TTS Result
type TTSResult struct {
	AudioPath string
	AudioMime string
	Text      string
}

// Stub methods to satisfy interface usage in channels
func (m *Manager) AutoApplyToText(ctx context.Context, content, channel string, isVoiceInbound bool, agentID string) (*TTSResult, error) {
	return nil, nil
}

func (m *Manager) Transcribe(ctx context.Context, in STTInput, opts STTOptions) (*STTResult, error) {
	return nil, nil
}

// STTInput, STTOptions, and STTResult are stub types
type STTInput struct {
	FilePath string
	MimeType string
}

type STTOptions struct{}

type STTResult struct {
	Text     string
	Duration int
	Language string
}

// Module-level stub functions
func StripTTSDirectives(s string) string { return s }
func ValidateAgentTTSParams(params map[string]interface{}) error { return nil }
