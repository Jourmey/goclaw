package channels

import (
	"context"
	"encoding/json"
)

// EnrichFileWriterMetadata stub - file writer metadata removed
func EnrichFileWriterMetadata(ctx context.Context, resolver interface{}, userID, name string) (json.RawMessage, bool) {
	return nil, false
}

// IsEmptyWriterMetadata stub - writer metadata removed
func IsEmptyWriterMetadata(metadata interface{}) bool {
	return true
}
