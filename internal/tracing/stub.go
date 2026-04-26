// Package tracing - removed module stub
package tracing

import (
	"context"

	"github.com/google/uuid"
)

// Stub package - tracing module has been removed

// Collector stub - tracing collector removed
type Collector struct{}

func (c *Collector) Start() error {
	return nil
}

func (c *Collector) Stop() error {
	return nil
}

func (c *Collector) PreviewMaxLen() int {
	return 0
}

func (c *Collector) Verbose() bool {
	return false
}

func (c *Collector) EmitSpan(ctx context.Context, span interface{}) error {
	return nil
}

func (c *Collector) CreateTrace(ctx context.Context, trace interface{}) error {
	return nil
}

func (c *Collector) FinishTrace(ctx context.Context, traceID uuid.UUID, status, errorMessage, outputPreview string) {
	// no-op
}

func (c *Collector) SetTraceStatus(ctx context.Context, traceID interface{}, status string) error {
	return nil
}

func (c *Collector) EmitSpanUpdate(ctx context.Context, spanID interface{}, updates map[string]interface{}) error {
	return nil
}

// WithTraceID stub
func WithTraceID(ctx context.Context, traceID uuid.UUID) context.Context {
	return ctx
}

// WithCollector stub
func WithCollector(ctx context.Context, collector interface{}) context.Context {
	return ctx
}

// WithParentSpanID stub
func WithParentSpanID(ctx context.Context, spanID uuid.UUID) context.Context {
	return ctx
}

// WithAnnounceParentSpanID stub
func WithAnnounceParentSpanID(ctx context.Context, spanID uuid.UUID) context.Context {
	return ctx
}

// WithTraceTeamID stub
func WithTraceTeamID(ctx context.Context, teamID uuid.UUID) context.Context {
	return ctx
}

// DelegateParentTraceIDFromContext stub
func DelegateParentTraceIDFromContext(ctx context.Context) uuid.UUID {
	return uuid.Nil
}

// CollectorFromContext stub
func CollectorFromContext(ctx context.Context) *Collector {
	return nil
}

// TraceIDFromContext stub
func TraceIDFromContext(ctx context.Context) uuid.UUID {
	return uuid.Nil
}

// ParentSpanIDFromContext stub
func ParentSpanIDFromContext(ctx context.Context) uuid.UUID {
	return uuid.Nil
}

// TraceTeamIDPtrFromContext stub
func TraceTeamIDPtrFromContext(ctx context.Context) *uuid.UUID {
	return nil
}

// TruncateJSON stub
func TruncateJSON(data interface{}, maxLen int) string {
	return ""
}

// LookupPricing stub
func LookupPricing(pricingMap interface{}, provider, model string) interface{} {
	return nil
}

// CalculateCost stub
func CalculateCost(pricing, usage interface{}) float64 {
	return 0
}

// TruncateMid stub
func TruncateMid(s string, maxLen int) string {
	return s
}
