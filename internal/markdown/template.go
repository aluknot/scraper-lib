// Package markdown provides template-driven Markdown output generation.
package markdown

import (
	"bytes"
	_ "embed" // required for //go:embed directive
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed templates/article.yaml
var defaultArticleTemplate []byte

// Template holds a parsed template definition for Markdown generation.
type Template struct {
	Name        string       `yaml:"name"`
	Frontmatter []string     `yaml:"frontmatter"`
	Body        TemplateBody `yaml:"body"`
}

// TemplateBody defines the structure of the Markdown body.
type TemplateBody struct {
	Heading  string `yaml:"heading"`
	Byline   string `yaml:"byline"`
	Content  string `yaml:"content"`
	Metadata string `yaml:"metadata"`
	Footer   string `yaml:"footer"`
}

// TemplateData holds the data available for template rendering.
type TemplateData struct {
	Title       string
	URL         string
	Source      string
	Author      string
	PublishedAt string
	ExtractedAt string
	Extractor   string
	Language    string
	WordCount   int
	Category    string
	Content     string
	Tags        []string
	Warnings    []string
	Extra       map[string]string // For extractor-specific fields
}

// Manager loads and resolves templates.
type Manager struct {
	dir       string
	templates map[string]*Template
	fallback  *Template
}

// NewManager creates a template manager loading templates from the given directory.
// If dir is empty, uses built-in defaults.
func NewManager(dir string) (*Manager, error) {
	m := &Manager{
		dir:       dir,
		templates: make(map[string]*Template),
	}

	// Load built-in default
	defaultTpl, err := parseTemplate("article", bytes.NewReader(defaultArticleTemplate))
	if err != nil {
		return nil, fmt.Errorf("parse default template: %w", err)
	}
	m.fallback = defaultTpl
	m.templates["article"] = defaultTpl

	// Load external templates
	if dir != "" {
		if err := m.loadDir(dir); err != nil {
			return nil, fmt.Errorf("load template dir: %w", err)
		}
	}

	return m, nil
}

// Resolve finds the best template for the given extractor name and category.
// Resolution order: extractor_category → extractor → article (fallback).
func (m *Manager) Resolve(extractor, category string) *Template {
	if category != "" {
		// Try extractor_category (e.g., "youtube_song")
		key := extractor + "_" + category
		if tpl, ok := m.templates[key]; ok {
			return tpl
		}
		// Try category alone
		if tpl, ok := m.templates[category]; ok {
			return tpl
		}
	}

	// Try extractor name
	if tpl, ok := m.templates[extractor]; ok {
		return tpl
	}

	// Try just the extractor prefix
	if category != "" {
		prefix := extractor + "_"
		for name, tpl := range m.templates {
			if strings.HasPrefix(name, prefix) {
				return tpl
			}
		}
	}

	// Fallback to article
	return m.fallback
}

// Render generates Markdown from a template and data.
func (m *Manager) Render(tpl *Template, data TemplateData) string {
	if tpl == nil {
		tpl = m.fallback
	}

	var buf strings.Builder

	// Frontmatter
	buf.WriteString("---\n")
	fm := m.buildFrontmatter(tpl, data)
	for _, field := range tpl.Frontmatter {
		value, ok := fm[field]
		if !ok || value == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("%s: %s\n", field, yamlValue(value)))
	}
	buf.WriteString("---\n\n")

	// Body
	if tpl.Body.Heading != "" {
		buf.WriteString(renderVar(tpl.Body.Heading, data))
		buf.WriteString("\n\n")
	}

	if tpl.Body.Byline != "" {
		byline := renderVar(tpl.Body.Byline, data)
		if strings.TrimSpace(byline) != "" {
			buf.WriteString(byline)
			buf.WriteString("\n\n")
		}
	}

	if tpl.Body.Content != "" {
		buf.WriteString(renderVar(tpl.Body.Content, data))
		buf.WriteString("\n\n")
	}

	if tpl.Body.Metadata != "" {
		buf.WriteString("---\n\n")
		buf.WriteString(renderVar(tpl.Body.Metadata, data))
		buf.WriteString("\n\n")
	}

	if tpl.Body.Footer != "" {
		buf.WriteString("---\n\n")
		buf.WriteString("*" + renderVar(tpl.Body.Footer, data) + "*")
	}

	return buf.String()
}

// loadDir loads all .yaml files from a directory.
func (m *Manager) loadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "scraper-lib: warning: cannot read template %s: %v\n", e.Name(), err)
			continue
		}
		tpl, err := parseTemplate(name, bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "scraper-lib: warning: cannot parse template %s: %v\n", e.Name(), err)
			continue
		}
		m.templates[name] = tpl
	}
	return nil
}

// buildFrontmatter builds a map of frontmatter fields from the template and data.
func (m *Manager) buildFrontmatter(tpl *Template, data TemplateData) map[string]string {
	fm := make(map[string]string)
	for _, field := range tpl.Frontmatter {
		switch field {
		case "title":
			fm["title"] = data.Title
		case "url":
			fm["url"] = data.URL
		case "source":
			fm["source"] = data.Source
		case "author":
			fm["author"] = data.Author
		case "published_at":
			fm["published_at"] = data.PublishedAt
		case "extracted_at":
			fm["extracted_at"] = data.ExtractedAt
		case "extractor":
			fm["extractor"] = data.Extractor
		case "language":
			fm["language"] = data.Language
		case "word_count":
			if data.WordCount > 0 {
				fm["word_count"] = fmt.Sprintf("%d", data.WordCount)
			}
		case "category":
			fm["category"] = data.Category
		case "tags":
			if len(data.Tags) > 0 {
				fm["tags"] = strings.Join(data.Tags, ", ")
			}
		case "warnings":
			if len(data.Warnings) > 0 {
				fm["warnings"] = strings.Join(data.Warnings, ", ")
			}
		default:
			if v, ok := data.Extra[field]; ok {
				fm[field] = v
			}
		}
	}
	return fm
}

// parseTemplate parses a YAML template from a byte reader.
func parseTemplate(name string, r *bytes.Reader) (*Template, error) {
	var tpl Template
	if err := yaml.NewDecoder(r).Decode(&tpl); err != nil {
		return nil, err
	}
	if tpl.Name == "" {
		tpl.Name = name
	}
	return &tpl, nil
}

// renderVar replaces {{var}} placeholders with data values.
func renderVar(template string, data TemplateData) string {
	result := template
	result = strings.ReplaceAll(result, "{{title}}", data.Title)
	result = strings.ReplaceAll(result, "{{url}}", data.URL)
	result = strings.ReplaceAll(result, "{{source}}", data.Source)
	result = strings.ReplaceAll(result, "{{author}}", data.Author)
	result = strings.ReplaceAll(result, "{{published_at}}", data.PublishedAt)
	result = strings.ReplaceAll(result, "{{extracted_at}}", data.ExtractedAt)
	result = strings.ReplaceAll(result, "{{extractor}}", data.Extractor)
	result = strings.ReplaceAll(result, "{{language}}", data.Language)
	result = strings.ReplaceAll(result, "{{content}}", data.Content)
	result = strings.ReplaceAll(result, "{{category}}", data.Category)
	result = strings.ReplaceAll(result, "{{word_count}}", fmt.Sprintf("%d", data.WordCount))
	if len(data.Tags) > 0 {
		result = strings.ReplaceAll(result, "{{tags}}", strings.Join(data.Tags, ", "))
	}
	for k, v := range data.Extra {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

// yamlValue formats a value for YAML frontmatter, quoting if needed.
func yamlValue(value string) string {
	if strings.ContainsAny(value, `:"{},[]&*#?|-<>=!%@`) {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

// GenerateFilename creates a sanitized filename from title and source.
func GenerateFilename(title, source string) string {
	var buf strings.Builder
	for _, r := range title {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' {
			buf.WriteRune(r)
		}
	}
	name := buf.String()
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	if len(name) > 80 {
		name = name[:80]
	}
	name = strings.TrimRight(name, "-")

	// Clean source domain
	domain := source
	if idx := strings.Index(domain, "://"); idx >= 0 {
		domain = domain[idx+3:]
	}
	if idx := strings.Index(domain, "/"); idx >= 0 {
		domain = domain[:idx]
	}
	if domain == "" {
		domain = "untitled"
	}

	return fmt.Sprintf("%s/%s.md", domain, name)
}

// GenerateTags extracts simple tags from title and language.
func GenerateTags(title, language string, maxTags int) []string {
	var tags []string

	langTags := map[string]string{
		"en": "english",
		"es": "spanish",
		"pt": "portuguese",
		"fr": "french",
		"de": "german",
	}
	if tag, ok := langTags[language]; ok {
		tags = append(tags, tag)
	}

	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"this": true, "that": true, "it": true, "from": true, "as": true,
	}

	words := strings.Fields(strings.ToLower(title))
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?\"'()-")
		if len(w) > 3 && !stopWords[w] && len(tags) < maxTags {
			tags = append(tags, w)
		}
	}

	return tags
}

// NowRFC3339 returns the current UTC time in RFC3339 format.
func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
