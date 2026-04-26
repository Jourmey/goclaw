// Package media - removed module stub
package media

// Stub package - channels/media module has been removed

// MediaInfo stub - media info removed
type MediaInfo struct {
	Type        string
	FilePath    string
	ContentType string
	FileName    string
}

// DetectMIMEType stub - MIME type detection removed
func DetectMIMEType(filename string) string {
	return "application/octet-stream"
}

// MediaKindFromMime stub - media kind removed
func MediaKindFromMime(mimeType string) string {
	return "file"
}

// BuildMediaTags stub - media tags removed
func BuildMediaTags(infos []MediaInfo) string {
	return ""
}
