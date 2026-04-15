package markdown

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdown_Basic(t *testing.T) {
	input := `<h1>Title</h1><p>Hello <strong>world</strong> with <em>emphasis</em>.</p>`
	result := HTMLToMarkdown(input)

	if !strings.Contains(result, "# Title") {
		t.Errorf("expected heading, got:\n%s", result)
	}
	if !strings.Contains(result, "**world**") {
		t.Errorf("expected bold markdown, got:\n%s", result)
	}
	if !strings.Contains(result, "*emphasis*") {
		t.Errorf("expected italic markdown, got:\n%s", result)
	}
}

func TestHTMLToMarkdown_LinksAndImages(t *testing.T) {
	input := `<p>Visit <a href="https://example.com">example</a></p>`
	result := HTMLToMarkdown(input)

	if !strings.Contains(result, "[example](https://example.com)") {
		t.Errorf("expected link markdown, got:\n%s", result)
	}
}

func TestTemplateResolve_Default(t *testing.T) {
	mgr, err := NewManager("")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	tpl := mgr.Resolve("readability", "")
	if tpl.Name != "article" {
		t.Errorf("expected 'article' template, got %q", tpl.Name)
	}
}

func TestTemplateRender_Default(t *testing.T) {
	mgr, _ := NewManager("")
	tpl := mgr.Resolve("readability", "")

	data := TemplateData{
		Title:       "Test Article",
		URL:         "https://example.com/test",
		Source:      "example.com",
		Author:      "John Doe",
		ExtractedAt: "2026-04-12T00:00:00Z",
		Extractor:   "readability",
		Content:     "This is the content.",
		WordCount:   100,
	}

	result := mgr.Render(tpl, data)

	if !strings.Contains(result, "---") {
		t.Errorf("expected frontmatter, got:\n%s", result)
	}
	if !strings.Contains(result, "title: Test Article") {
		t.Errorf("expected title in frontmatter, got:\n%s", result)
	}
	if !strings.Contains(result, "Test Article\n\n") {
		t.Errorf("expected heading in body, got:\n%s", result)
	}
	if !strings.Contains(result, "This is the content.") {
		t.Errorf("expected content in body, got:\n%s", result)
	}
}

func TestGenerateFilename(t *testing.T) {
	filename := GenerateFilename("My Great Article", "https://example.com/path")
	expected := "example.com/my-great-article"
	if !strings.Contains(filename, expected) {
		t.Errorf("expected %q in %q", expected, filename)
	}
}

func TestGenerateTags(t *testing.T) {
	tags := GenerateTags("How to Learn Go Programming", "en", 5)

	// Should have language tag + some keywords
	found := false
	for _, tag := range tags {
		if tag == "english" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'english' tag, got %v", tags)
	}
}

func TestYAMLValue_Quoting(t *testing.T) {
	// Simple value
	if yamlValue("hello") != "hello" {
		t.Errorf("expected unquoted, got %q", yamlValue("hello"))
	}
	// Value with special chars
	if yamlValue("hello: world") != `"hello: world"` {
		t.Errorf("expected quoted, got %q", yamlValue("hello: world"))
	}
}

func TestNowRFC3339(t *testing.T) {
	now := NowRFC3339()
	if now == "" {
		t.Error("expected non-empty timestamp")
	}
}
