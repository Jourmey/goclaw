package agent

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/nextlevelbuilder/goclaw/internal/providers"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

func (l *Loop) emit(event AgentEvent) {
	if l.onEvent != nil {
		l.onEvent(event)
	}
}

// ID returns the agent's identifier (agent_key, e.g. "goctech-leader").
// Use for logs, UI, filesystem paths. NEVER for DB FK or DomainEvent.AgentID.
// See docs/agent-identity-conventions.md.
func (l *Loop) ID() string { return l.id }

// UUID returns the agent's canonical UUID (DB primary key).
// Use for SQL WHERE/JOIN, DomainEvent.AgentID, context propagation.
// See docs/agent-identity-conventions.md.
func (l *Loop) UUID() uuid.UUID { return l.agentUUID }

// OtherConfig returns the agent's other_config JSONB (extensibility bag).
// Used for per-agent TTS voice override (tts_voice_id, tts_model_id).
func (l *Loop) OtherConfig() json.RawMessage { return l.agentOtherConfig }

// Model returns the model identifier for this agent loop.
func (l *Loop) Model() string { return l.model }

// IsRunning returns whether the agent is currently processing.
func (l *Loop) IsRunning() bool { return l.activeRuns.Load() > 0 }

// ---------------------------------------------------------------------------
// Span options — functional options for overriding model/provider in spans.
// ---------------------------------------------------------------------------

// spanOption overrides span metadata (model, provider) when per-request
// overrides are active (e.g. heartbeat with a cheaper model).
type spanOption func(*spanOverrides)

type spanOverrides struct {
	model    string
	provider string
}

func withModel(m string) spanOption    { return func(o *spanOverrides) { o.model = m } }
func withProvider(p string) spanOption { return func(o *spanOverrides) { o.provider = p } }

// resolveSpan returns (model, provider) applying any overrides on top of agent defaults.
func (l *Loop) resolveSpan(opts []spanOption) (string, string) {
	o := spanOverrides{model: l.model, provider: l.provider.Name()}
	for _, fn := range opts {
		fn(&o)
	}
	return o.model, o.provider
}

// ---------------------------------------------------------------------------
// Two-phase LLM span: start (running) + end (completed/error)
// ---------------------------------------------------------------------------

// emitLLMSpanStart stub - tracing removed
func (l *Loop) emitLLMSpanStart(ctx context.Context, start time.Time, iteration int, messages []providers.Message, opts ...spanOption) uuid.UUID {
	return uuid.Nil
}

// emitLLMSpanEnd stub - tracing removed
func (l *Loop) emitLLMSpanEnd(ctx context.Context, spanID uuid.UUID, start time.Time, resp *providers.ChatResponse, callErr error, opts ...spanOption) {
}

// ---------------------------------------------------------------------------
// Two-phase tool span: start (running) + end (completed/error)
// ---------------------------------------------------------------------------

// emitToolSpanStart stub - tracing removed
func (l *Loop) emitToolSpanStart(ctx context.Context, start time.Time, toolName, toolCallID, input string) uuid.UUID {
	return uuid.Nil
}

// emitToolSpanEnd stub - tracing removed
func (l *Loop) emitToolSpanEnd(ctx context.Context, spanID uuid.UUID, start time.Time, result *tools.Result) {
}

// ---------------------------------------------------------------------------
// Two-phase agent span: start (running) + end (completed/error)
// ---------------------------------------------------------------------------

// emitAgentSpanStart stub - tracing removed
func (l *Loop) emitAgentSpanStart(ctx context.Context, agentSpanID uuid.UUID, start time.Time, inputPreview string, opts ...spanOption) {
}

// emitAgentSpanEnd stub - tracing removed
func (l *Loop) emitAgentSpanEnd(ctx context.Context, agentSpanID uuid.UUID, start time.Time, result *RunResult, runErr error) {
}

// previewLimitForVerbose returns the preview character limit based on verbose mode.
func previewLimitForVerbose(verbose bool) int {
	if verbose {
		return 200_000
	}
	return 40_000
}

func truncateStr(s string, maxLen int) string {
	s = strings.ToValidUTF8(s, "")
	if len(s) <= maxLen {
		return s
	}
	// Keep the tail — recent context is more useful for debugging.
	start := len(s) - maxLen
	// Don't cut in the middle of a multi-byte rune
	for start < len(s) && !utf8.RuneStart(s[start]) {
		start++
	}
	return "..." + s[start:]
}

// estimateMessageTokens returns a rough token estimate for a single message,
// including content text and tool call arguments.
func estimateMessageTokens(m providers.Message) int {
	tokens := utf8.RuneCountInString(m.Content) / 3
	for _, tc := range m.ToolCalls {
		tokens += len(tc.ID)/3 + len(tc.Name)/3
		for k, v := range tc.Arguments {
			tokens += len(k) / 3
			switch val := v.(type) {
			case string:
				tokens += len(val) / 3
			default:
				tokens += 10 // small fixed estimate for non-string args (numbers, booleans, etc.)
			}
		}
	}
	return tokens
}

// EstimateTokens returns a rough token estimate for a slice of messages.
// Includes content text and tool call arguments (JSON overhead).
// Used internally for summarization thresholds and externally for adaptive throttle.
func EstimateTokens(messages []providers.Message) int {
	total := 0
	for _, m := range messages {
		total += estimateMessageTokens(m)
	}
	return total
}

// EstimateHistoryTokens estimates tokens for history messages only,
// excluding system messages (which are overhead: system prompt, tool defs, context files).
// Used for compaction threshold checks where we need history-only token count.
func EstimateHistoryTokens(messages []providers.Message) int {
	total := 0
	for _, m := range messages {
		if m.Role == "system" {
			continue
		}
		total += estimateMessageTokens(m)
	}
	return total
}

// EstimateTokensWithCalibration uses actual prompt tokens from the last LLM
// response as a calibration base, then estimates only new messages on top.
// Falls back to EstimateTokens() when no calibration data is available.
func EstimateTokensWithCalibration(messages []providers.Message, lastPromptTokens, lastMsgCount int) int {
	if lastPromptTokens <= 0 || lastMsgCount <= 0 {
		return EstimateTokens(messages)
	}

	currentCount := len(messages)
	newMsgs := currentCount - lastMsgCount
	if newMsgs <= 0 {
		// No new messages since last calibration (or history was truncated).
		// Use calibration value as-is; it's the best estimate we have.
		return lastPromptTokens
	}

	// Estimate only the new messages with the heuristic and add to base.
	delta := 0
	for _, m := range messages[lastMsgCount:] {
		delta += estimateMessageTokens(m)
	}
	return lastPromptTokens + delta
}
