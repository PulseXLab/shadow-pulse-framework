package utils

import "strings"

// SanitizeURL creates a valid filename from a URL string.
func SanitizeURL(rawURL string) string {
	// Remove protocol scheme
	sanitized := strings.Replace(rawURL, "https://", "", 1)
	sanitized = strings.Replace(sanitized, "http://", "", 1)
	// Replace invalid filename characters
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	return sanitized + ".png"
}
