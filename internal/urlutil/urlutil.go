// Package urlutil provides URL parsing utilities.
package urlutil

import "strings"

// Domain extracts the domain (host) from a URL string.
// For example, "https://www.example.com/path?a=1" → "www.example.com".
func Domain(rawURL string) string {
	if idx := strings.Index(rawURL, "://"); idx >= 0 {
		rawURL = rawURL[idx+3:]
	}
	if idx := strings.Index(rawURL, "/"); idx >= 0 {
		rawURL = rawURL[:idx]
	}
	// Remove query string
	if idx := strings.Index(rawURL, "?"); idx >= 0 {
		rawURL = rawURL[:idx]
	}
	// Remove fragment
	if idx := strings.Index(rawURL, "#"); idx >= 0 {
		rawURL = rawURL[:idx]
	}
	return rawURL
}

// Path extracts the path component from a URL string.
// For example, "https://www.example.com/path?a=1" → "/path".
func Path(rawURL string) string {
	if idx := strings.Index(rawURL, "://"); idx >= 0 {
		rawURL = rawURL[idx+3:]
	}
	if idx := strings.Index(rawURL, "/"); idx >= 0 {
		path := rawURL[idx:]
		// Remove query string
		if qIdx := strings.Index(path, "?"); qIdx >= 0 {
			path = path[:qIdx]
		}
		return path
	}
	return "/"
}
