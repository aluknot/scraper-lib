package markdown

import (
	"regexp"
	"strings"
)

// HTMLToMarkdown converts a subset of HTML to Markdown.
// Handles: headings, bold, italic, links, lists, code, blockquotes.
func HTMLToMarkdown(htmlContent string) string {
	result := htmlContent

	// Headings
	result = replaceAllGroups(result, `<h1[^>]*>(.*?)</h1>`, "# $1\n\n")
	result = replaceAllGroups(result, `<h2[^>]*>(.*?)</h2>`, "## $1\n\n")
	result = replaceAllGroups(result, `<h3[^>]*>(.*?)</h3>`, "### $1\n\n")
	result = replaceAllGroups(result, `<h4[^>]*>(.*?)</h4>`, "#### $1\n\n")
	result = replaceAllGroups(result, `<h5[^>]*>(.*?)</h5>`, "##### $1\n\n")
	result = replaceAllGroups(result, `<h6[^>]*>(.*?)</h6>`, "###### $1\n\n")

	// Bold and italic
	result = replaceAllGroups(result, `<strong[^>]*>(.*?)</strong>`, "**$1**")
	result = replaceAllGroups(result, `<b[^>]*>(.*?)</b>`, "**$1**")
	result = replaceAllGroups(result, `<em[^>]*>(.*?)</em>`, "*$1*")
	result = replaceAllGroups(result, `<i[^>]*>(.*?)</i>`, "*$1*")

	// Links
	result = replaceAllGroups(result, `<a[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`, "[$2]($1)")

	// Images
	result = replaceAllGroups(result, `<img[^>]*src=["']([^"']*)["'][^>]*alt=["']([^"']*)["'][^>]*/?>`, "![$2]($1)")
	result = replaceAllGroups(result, `<img[^>]*src=["']([^"']*)["'][^>]*/?>`, "![]($1)")

	// Code
	result = replaceAllGroups(result, `<pre[^>]*><code[^>]*>(.*?)</code></pre>`, "```\n$1\n```\n\n")
	result = replaceAllGroups(result, `<code[^>]*>(.*?)</code>`, "`$1`")

	// Blockquotes
	result = replaceAllGroups(result, `<blockquote[^>]*>(.*?)</blockquote>`, "> $1\n\n")

	// Lists
	result = replaceAllGroups(result, `<li[^>]*>(.*?)</li>`, "- $1\n")

	// Line breaks
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")

	// Paragraphs
	result = strings.ReplaceAll(result, "</p>", "\n\n")
	result = replaceAllGroups(result, `<p[^>]*>`, "")

	// Remove all remaining HTML tags
	result = removeAllTags(result)

	// Normalize whitespace
	result = normalizeWhitespace(result)

	return strings.TrimSpace(result)
}

func replaceAllGroups(s, pattern, replacement string) string {
	re := regexp.MustCompile(`(?is)` + pattern)
	return re.ReplaceAllString(s, replacement)
}

func removeAllTags(s string) string {
	re := regexp.MustCompile(`</?[^>]+>`)
	return re.ReplaceAllString(s, "")
}

var multiNewline = regexp.MustCompile(`\n{3,}`)

func normalizeWhitespace(s string) string {
	s = multiNewline.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
