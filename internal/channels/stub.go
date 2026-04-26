// Package channels - removed module stub
package channels

import "strings"

// Stub package - channels module has been removed

// SanitizeDisplayName stub - simple implementation
func SanitizeDisplayName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Unknown"
	}
	return name
}
