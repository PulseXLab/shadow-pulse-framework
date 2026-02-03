package utils

import "strings"

// SanitizeURL creates a valid filename from a URL string for image files.
func SanitizeURL(rawURL string) string {
	// Remove protocol scheme
	sanitized := strings.Replace(rawURL, "https://", "", 1)
	sanitized = strings.Replace(sanitized, "http://", "", 1)
	// Replace invalid filename characters
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	return sanitized + ".png"
}

// SanitizeFilename creates a safe filename by replacing invalid characters.
func SanitizeFilename(filename string) string {
	// Remove protocol schemes
	sanitized := strings.Replace(filename, "https://", "", 1)
	sanitized = strings.Replace(sanitized, "http://", "", 1)

	// Replace common invalid characters with an underscore
	replacer := strings.NewReplacer(
		"/", "_",
		":", "_",
		"?", "_",
		"&", "_",
		"=", "_",
		"\\", "_",
		"*", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"\"", "_",
	)
	return replacer.Replace(sanitized)
}
