package tools

// TtsTool is a stub for TTS support.
// TTS was removed in v3.x. This stub exists only to maintain compatibility with cmd/
// that references TtsTool in setupToolRegistry return types.
type TtsTool struct{}

// CreateAudioTool is a stub for audio generation tool.
// TTS was removed in v3.x. This stub exists only to maintain compatibility with cmd/
// that may reference CreateAudioTool in tool registrations.
type CreateAudioTool struct{}
