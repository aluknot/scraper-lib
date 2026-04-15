package fetch

import (
	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer wraps bluemonday policies for HTML sanitization.
type Sanitizer struct {
	policy *bluemonday.Policy
}

// NewSanitizer creates a sanitizer with the UGC policy (User Generated Content),
// which allows common HTML elements while stripping dangerous ones (scripts,
// event handlers, etc.).
func NewSanitizer() *Sanitizer {
	p := bluemonday.UGCPolicy()

	// Semantic content structure
	p.AllowElements("p", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "blockquote", "pre", "code",
		"strong", "em", "b", "i", "u", "a", "img", "br", "hr",
		"table", "thead", "tbody", "tr", "th", "td", "figure", "figcaption",
		"iframe", "sup", "sub", "del", "ins", "kbd", "samp", "var")

	// Links with href
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("target", "rel").OnElements("a")

	// Images with src, alt, title
	p.AllowAttrs("src", "alt", "title", "loading").OnElements("img")

	// Code blocks with class
	p.AllowAttrs("class").OnElements("pre", "code", "figure")

	// Table cell spanning
	p.AllowAttrs("colspan", "rowspan").OnElements("th", "td")

	// Iframes for embeds (YouTube, Vimeo, etc.) — restored after sanitization
	p.AllowAttrs("src", "frameborder", "allowfullscreen", "allow", "loading",
		"width", "height").OnElements("iframe")

	// Blockquotes with class (for restored embed wrappers)
	// Allow class on figure elements generally
	p.AllowAttrs("class").OnElements("figure")

	return &Sanitizer{policy: p}
}

// Clean sanitizes HTML content, stripping scripts, event handlers, and
// potentially dangerous elements while preserving article content.
func (s *Sanitizer) Clean(input string) string {
	return s.policy.Sanitize(input)
}
