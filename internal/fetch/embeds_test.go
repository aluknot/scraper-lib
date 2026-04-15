package fetch

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractAndReplace_SingleYouTube(t *testing.T) {
	input := `<html><body>
<p>Intro text with enough words to be interesting.</p>
<iframe src="https://www.youtube.com/embed/abc123" width="560" height="315" allowfullscreen></iframe>
<p>More text after the video embed.</p>
</body></html>`

	e := NewEmbedExtractor()
	processed, embedMap := e.ExtractAndReplace(input)

	if len(embedMap) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embedMap))
	}

	if !strings.Contains(processed, "[[EMBED_") {
		t.Errorf("expected placeholder in processed HTML, got:\n%s", processed)
	}

	for _, outerHTML := range embedMap {
		if !strings.Contains(outerHTML, `width="560"`) {
			t.Errorf("expected width attribute preserved, got:\n%s", outerHTML)
		}
		if !strings.Contains(outerHTML, `height="315"`) {
			t.Errorf("expected height attribute preserved, got:\n%s", outerHTML)
		}
	}
}

func TestExtractAndReplace_ConsecutiveEmbeds(t *testing.T) {
	input := `<html><body>
<p>Intro text with enough content to be interesting for testing.</p>
<iframe src="https://www.youtube.com/embed/aaa111" allowfullscreen></iframe>
<iframe src="https://www.youtube.com/embed/bbb222" allowfullscreen></iframe>
<p>More content after both embeds.</p>
</body></html>`

	e := NewEmbedExtractor()
	processed, embedMap := e.ExtractAndReplace(input)

	if len(embedMap) != 2 {
		t.Fatalf("expected 2 embeds, got %d", len(embedMap))
	}

	count := strings.Count(processed, "[[EMBED_")
	if count != 2 {
		t.Errorf("expected 2 placeholders, found %d", count)
	}

	// Both should share the same runID
	var runID string
	for placeholder := range embedMap {
		start := strings.Index(placeholder, "EMBED_") + 6
		end := strings.LastIndex(placeholder, "_")
		id := placeholder[start:end]
		if runID == "" {
			runID = id
		} else if runID != id {
			t.Errorf("runIDs differ: %s vs %s", runID, id)
		}
	}
}

func TestExtractAndReplace_PlaceholderTextInPage(t *testing.T) {
	input := `<html><body>
<article>
<p>This page discusses the format [[EMBED_test_0]] which is used in some systems.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.</p>
</article>
</body></html>`

	e := NewEmbedExtractor()
	processed, embedMap := e.ExtractAndReplace(input)

	if !strings.Contains(processed, "[[EMBED_test_0]]") {
		t.Errorf("expected literal placeholder text preserved, got:\n%s", processed)
	}

	if len(embedMap) != 0 {
		t.Errorf("expected 0 embeds from text, got %d", len(embedMap))
	}
}

func TestSanitizeBeforeRestore_IframesPreserved(t *testing.T) {
	input := `<html><body>
<article>
<p>Article text with enough words to pass any threshold. Lorem ipsum dolor sit amet,
consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore.</p>
<iframe src="https://www.youtube.com/embed/xyz789" allowfullscreen></iframe>
</article>
</body></html>`

	e := NewEmbedExtractor()
	s := NewSanitizer()

	processed, embedMap := e.ExtractAndReplace(input)
	sanitized := s.Clean(processed)
	restored := e.Restore(sanitized, embedMap)

	if !strings.Contains(restored, "youtube") {
		t.Errorf("expected youtube iframe after restore, got:\n%s", restored)
	}
	if !strings.Contains(restored, "embedded-content") {
		t.Errorf("expected embedded-content wrapper, got:\n%s", restored)
	}
}

func TestIsSupportedEmbed_TwitterVariants(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{"twitter-tweet simple", `<blockquote class="twitter-tweet"><p>tweet</p></blockquote>`, true},
		{"twitter-tweet aligned", `<blockquote class="twitter-tweet tw-align-center"><p>tweet</p></blockquote>`, true},
		{"instagram-media", `<blockquote class="instagram-media" data-instagram>post</blockquote>`, true},
		{"tiktok-preserve", `<blockquote class="tiktok-preserve">video</blockquote>`, true},
		{"random blockquote", `<blockquote class="some-class">quote</blockquote>`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			bq := findBlockquote(doc)
			if bq == nil {
				t.Fatal("no blockquote found in test HTML")
			}
			result := isSupportedEmbed(bq)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractAndRestore_Vimeo(t *testing.T) {
	input := `<html><body>
<article>
<p>Article discussing a Vimeo video. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Sed do eiusmod tempor incididunt ut labore.</p>
<iframe src="https://player.vimeo.com/video/12345" width="640" height="360" allowfullscreen></iframe>
</article>
</body></html>`

	e := NewEmbedExtractor()
	processed, embedMap := e.ExtractAndReplace(input)

	if len(embedMap) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(embedMap))
	}

	found := false
	for _, v := range embedMap {
		if strings.Contains(v, "vimeo.com") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected vimeo URL in embedMap, got:\n%v", embedMap)
	}

	_ = processed
}

func TestRestore_WrapsInFigure(t *testing.T) {
	e := NewEmbedExtractor()
	embedMap := map[string]string{
		"[[EMBED_test_0]]": `<iframe src="https://youtube.com/embed/test"></iframe>`,
	}

	result := e.Restore("Hello [[EMBED_test_0]] World", embedMap)

	expected := `<figure class="embedded-content"><iframe src="https://youtube.com/embed/test"></iframe></figure>`
	if !strings.Contains(result, expected) {
		t.Errorf("expected figure wrapper, got:\n%s", result)
	}
}

// findBlockquote walks the HTML tree and returns the first blockquote element.
func findBlockquote(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "blockquote" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findBlockquote(c); result != nil {
			return result
		}
	}
	return nil
}
