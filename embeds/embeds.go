// Package embeds provides embed extraction and restoration.
package embeds

import (
	"fmt"
	"strings"

	"github.com/aluknot/scraper-lib/internal/dom"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

// EmbedExtractor handles extraction and restoration of embedded content
// (YouTube, Twitter, Vimeo, Spotify, etc.) from HTML documents.
//
// The strategy is conservative: only known platforms are preserved.
// Embeds are extracted BEFORE extractors run (as placeholders) and
// restored AFTER sanitization to prevent bluemonday from stripping iframes.
type EmbedExtractor struct{}

// NewEmbedExtractor creates a new EmbedExtractor.
func NewEmbedExtractor() *EmbedExtractor {
	return &EmbedExtractor{}
}

// ExtractAndReplace extracts known embeds from rawHTML and replaces them with
// unique placeholders. Returns the processed HTML and a map of placeholders
// to original outerHTML.
//
// Uses UUID in placeholders to prevent collisions with page content.
// Walks the DOM tree saving NextSibling before recursion to avoid bugs
// when the tree is modified during the walk.
func (e *EmbedExtractor) ExtractAndReplace(rawHTML string) (processedHTML string, embedMap map[string]string) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML, nil
	}

	embedMap = map[string]string{}
	runID := uuid.New().String()
	counter := 0

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if isSupportedEmbed(n) {
			// Save the complete outerHTML (preserves width, height, allowfullscreen, etc.)
			var buf strings.Builder
			html.Render(&buf, n)
			outerHTML := buf.String()

			placeholder := fmt.Sprintf("[[EMBED_%s_%d]]", runID, counter)
			embedMap[placeholder] = outerHTML

			textNode := &html.Node{
				Type: html.TextNode,
				Data: placeholder,
			}
			n.Parent.InsertBefore(textNode, n)
			n.Parent.RemoveChild(n)
			counter++
			// Don't recurse into the replaced node.
		} else {
			// Save NextSibling before recursing to avoid skipping nodes
			// when the tree is modified during the walk.
			for c := n.FirstChild; c != nil; {
				next := c.NextSibling
				walk(c)
				c = next
			}
		}
	}
	walk(doc)

	var buf strings.Builder
	html.Render(&buf, doc)
	return buf.String(), embedMap
}

// Restore replaces placeholders with their original embed HTML,
// wrapped in <figure class="embedded-content"> for semantic context.
// MUST be called AFTER sanitizeHTML to prevent bluemonday from
// stripping the restored iframes.
func (e *EmbedExtractor) Restore(extractedHTML string, embedMap map[string]string) string {
	result := extractedHTML
	for placeholder, originalHTML := range embedMap {
		wrapper := fmt.Sprintf(`<figure class="embedded-content">%s</figure>`, originalHTML)
		result = strings.Replace(result, placeholder, wrapper, 1)
	}
	return result
}

// isSupportedEmbed determines if an HTML node is a known embed that should
// be preserved. Uses an explicit allowlist of platforms to avoid preserving
// ads, cookie consent widgets, or subscription widgets.
func isSupportedEmbed(n *html.Node) bool {
	// Case 1: Iframes from known platforms
	if n.Type == html.ElementNode && n.Data == "iframe" {
		src := dom.Attr(n, "src")
		for _, platform := range knownPlatforms {
			if strings.Contains(src, platform) {
				return true
			}
		}
		return false
	}

	// Case 2: Blockquote-based embeds (Twitter/X, Instagram, TikTok)
	if n.Type == html.ElementNode && n.Data == "blockquote" {
		class := dom.Attr(n, "class")
		return strings.Contains(class, "twitter-tweet") ||
			strings.Contains(class, "instagram-media") ||
			strings.Contains(class, "tiktok-preserve")
	}

	return false
}

// knownPlatforms lists URL patterns that indicate an iframe should be preserved.
var knownPlatforms = []string{
	"youtube.com",
	"youtu.be",
	"vimeo.com",
	"player.vimeo.com",
	"spotify.com",
	"open.spotify.com",
	"soundcloud.com",
	"w.soundcloud.com",
	"codepen.io",
	"instagram.com/embed",
	"tiktok.com/embed",
	"www.tiktok.com/embed",
	"twitter.com/embed",
	"x.com/embed",
	"platform.twitter.com",
}
