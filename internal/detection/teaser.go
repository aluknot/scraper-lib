// Package detection provides content quality detection utilities.
package detection

import (
	"strings"
)

var knownTeaserDomains = []string{
	"lobste.rs",
	"news.ycombinator.com",
	"medium.com",
	"substack.com",
	"dev.to",
	"reddit.com",
	"github.com",
}

var teaserPatterns = []string{
	"comments",
	"read more",
	"click here",
	"click to read",
	"continue reading",
	"full article",
	"subscribe to",
	"sign up",
	"log in to read",
	"paywall",
}

// IsLikelyTeaser checks if content appears to be a truncated summary
// rather than a full article.
func IsLikelyTeaser(content string) bool {
	if len(content) > 300 {
		return false
	}

	lower := strings.ToLower(content)

	for _, pattern := range teaserPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	if strings.HasSuffix(content, "...") && len(content) < 150 {
		return true
	}

	if len(content) < 50 {
		return true
	}

	return false
}

// IsKnownTeaserDomain checks if the URL belongs to a domain known to
// serve teaser/summary content in RSS feeds.
func IsKnownTeaserDomain(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	for _, domain := range knownTeaserDomains {
		if strings.Contains(lower, domain) {
			return true
		}
	}
	return false
}

// NeedsScraping determines whether full HTML scraping is needed
// based on content length and known teaser signals.
func NeedsScraping(content string, rawURL string) bool {
	if len(content) < 200 {
		return true
	}
	if IsLikelyTeaser(content) {
		return true
	}
	if IsKnownTeaserDomain(rawURL) {
		return true
	}
	return false
}
