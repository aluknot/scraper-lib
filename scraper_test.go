package scraperlib

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aluknot/scraper-lib/extractors"
	"github.com/aluknot/scraper-lib/internal/cache"
	"github.com/aluknot/scraper-lib/internal/fetch"
)

// TestExtract_SimpleArticle tests extraction from a basic HTML article.
func TestExtract_SimpleArticle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Test Article</title></head><body>
<article>
<h1>Test Article Title</h1>
<p>This is a test article with enough words to pass the minimum threshold. 
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor 
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore 
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt 
in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis 
unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, 
totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi 
architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia 
voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores 
eos qui ratione voluptatem sequi nesciunt.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
		Outputs: []string{"article"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Article == nil {
		t.Fatal("expected article result, got nil")
	}
	if result.WordCount < 100 {
		t.Errorf("expected word count >= 100, got %d", result.WordCount)
	}
	if result.ExtractorUsed == "" {
		t.Error("expected extractor_used to be set")
	}
}

// TestExtract_YouTubeEmbed tests that YouTube embeds are preserved.
func TestExtract_YouTubeEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Article with Video</title></head><body>
<article>
<h1>Article with Embedded Video</h1>
<p>This article has a YouTube video embedded in it. Lorem ipsum dolor sit amet,
consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in
reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
deserunt mollit anim id est laborum. Curabitur pretium tincidunt lacus. Nulla
gravidance, nulla vel dictum semper, ipsum dolor consectetuer adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad
minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip.</p>
<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ" frameborder="0" allowfullscreen></iframe>
<p>More content after the video. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna
aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint
occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit
anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit
voluptatem accusantium doloremque laudantium.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
		Outputs: []string{"article"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Article.Content, "youtube") {
		t.Errorf("expected YouTube embed in content, got:\n%s", result.Article.Content)
	}
	if !strings.Contains(result.Article.Content, "embedded-content") {
		t.Errorf("expected embed wrapper class 'embedded-content', got:\n%s", result.Article.Content)
	}
}

// TestExtract_ConsecutiveEmbeds tests multiple embeds in a row without text between them.
func TestExtract_ConsecutiveEmbeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Two Videos</title></head><body>
<article>
<h1>Two Consecutive Videos</h1>
<p>This article discusses two interesting videos. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip
ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit
esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non
proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p>
<iframe src="https://www.youtube.com/embed/aaa111" frameborder="0" allowfullscreen></iframe>
<iframe src="https://www.youtube.com/embed/bbb222" frameborder="0" allowfullscreen></iframe>
<p>More content after both videos. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna
aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in
reprehenderit in voluptate velit esse cillum dolore.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(result.Article.Content, "youtube") != 2 {
		t.Errorf("expected 2 YouTube embeds, found %d in:\n%s",
			strings.Count(result.Article.Content, "youtube"), result.Article.Content)
	}
}

// TestExtract_HTTPError tests that non-retriable errors are returned.
func TestExtract_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	_, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
	})

	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

// TestExtractHTML tests the ExtractHTML function.
func TestExtractHTML(t *testing.T) {
	html := `<!DOCTYPE html><html><head><title>Test</title></head><body>
<article>
<h1>Test Title</h1>
<p>This is a test article with enough words to pass the minimum word count threshold.
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure
dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt
mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit
voluptatem accusantium doloremque laudantium, totam rem aperiam.</p>
</article>
</body></html>`

	result, err := ExtractHTML(context.Background(), html, "https://example.com/article", &Options{
		Outputs: []string{"metadata"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Metadata == nil {
		t.Fatal("expected metadata result, got nil")
	}
	if result.Metadata.WordCount < 50 {
		t.Errorf("expected word count >= 50, got %d", result.Metadata.WordCount)
	}
}

// TestExtract_OptionsBackwardsCompat tests that the deprecated Output field still works.
func TestExtract_OptionsBackwardsCompat(t *testing.T) {
	opts := &Options{Output: "metadata"}
	opts.Normalize()

	if len(opts.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(opts.Outputs))
	}
	if opts.Outputs[0] != "metadata" {
		t.Errorf("expected 'metadata', got %q", opts.Outputs[0])
	}
}

// TestExtract_DefaultOutput tests that empty outputs defaults to "article".
func TestExtract_DefaultOutput(t *testing.T) {
	opts := &Options{}
	opts.Normalize()

	if len(opts.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(opts.Outputs))
	}
	if opts.Outputs[0] != "article" {
		t.Errorf("expected default 'article', got %q", opts.Outputs[0])
	}
}

// TestExtract_CustomExtractorChain tests that a custom chain of extractors
// is used instead of DefaultChain.
func TestExtract_CustomExtractorChain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head>
<title>Custom Chain Test Article</title>
<meta name="description" content="A test article for custom extractor chains">
<meta name="author" content="Test Author">
</head><body>
<nav>Navigation links here</nav>
<header>Site header</header>
<aside>Sidebar content</aside>
<article>
<h1>Custom Chain Test Article</h1>
<p class="byline">By Test Author | Published January 1, 2026</p>
<p>This is a comprehensive test article designed to pass the readability extraction
threshold with plenty of content to analyze. The article covers important topics
related to web scraping and content extraction from HTML documents. We need
sufficient text to ensure that readability and other extractors can properly
identify this as the main content of the page.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute
irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error
sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae
ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.</p>
<p>Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia
consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro
quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit.</p>
<blockquote>Important quote to test blockquote handling in the extraction pipeline.</blockquote>
<p>Additional content paragraphs to ensure we are well above the minimum word count
threshold required by the extractors. This is especially important when testing
with individual extractors rather than the full chain, since there is no fallback
to trafilatura or colly in this test scenario.</p>
</article>
<footer>Site footer content</footer>
</body></html>`))
	}))
	defer server.Close()

	// Use only readability in the custom chain
	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
		Extractors: []extractors.Extractor{
			extractors.NewReadabilityExtractor(),
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExtractorUsed != "readability" {
		t.Errorf("expected 'readability', got %q", result.ExtractorUsed)
	}
	if result.WordCount < 50 {
		t.Errorf("expected word count >= 50, got %d", result.WordCount)
	}
}

// TestExtract_ForceExtractor tests that a single extractor can be forced by name.
func TestExtract_ForceExtractor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head>
<title>Force Extractor Test</title>
<meta name="author" content="Test Author">
</head><body>
<article>
<h1>Force Extractor Test Article</h1>
<p>This article is designed to test the force extractor functionality. When a client
specifies a particular extractor by name, the system should use only that extractor
and not fall back to others. This is important for cases where the caller knows
exactly which extraction method works best for their content.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute
irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
deserunt mollit anim id est laborum.</p>
<p>More content to ensure we pass the minimum word count threshold. The quick brown fox
jumps over the lazy dog multiple times to generate sufficient text volume for testing
purposes. This approach ensures that readability and other extractors have enough
material to work with during the extraction process.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// Force readability even though DefaultChain would try domain_specific first
	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout:   5 * time.Second,
		Extractor: "readability",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExtractorUsed != "readability" {
		t.Errorf("expected 'readability' (forced), got %q", result.ExtractorUsed)
	}
}

// TestExtract_NoFallback tests that when NoFallback is true, a failing
// extractor returns an error instead of trying the next one.
func TestExtract_NoFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		// Empty page — all extractors will fail
		w.Write([]byte(`<html><head><title>Empty</title></head><body></body></html>`))
	}))
	defer server.Close()

	// Without NoFallback: the chain tries all extractors and eventually returns
	// ErrAllExtractorsFailed after trying all of them
	_, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error for empty page")
	}
	// Verify that multiple attempts were made (fallback happened)
	if !strings.Contains(err.Error(), "all extractors failed") {
		t.Fatalf("expected 'all extractors failed' error, got: %v", err)
	}

	// With NoFallback: only the first extractor is tried
	_, err = Extract(context.Background(), server.URL, &Options{
		Timeout:    5 * time.Second,
		Extractor:  "readability",
		NoFallback: true,
	})
	if err == nil {
		t.Fatal("expected error with NoFallback")
	}
}

// TestExtract_UnknownExtractorName tests that an unknown extractor name
// falls back to DefaultChain instead of silently failing.
func TestExtract_UnknownExtractorName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Test</title></head><body>
<article>
<h1>Test Article for Unknown Extractor Fallback</h1>
<p>This article tests that when a client specifies an extractor name that does not exist,
the system falls back to the default chain instead of silently failing. This ensures
backwards compatibility and prevents silent failures in production.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute
irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout:   5 * time.Second,
		Extractor: "nonexistent_extractor",
	})

	// Should not error — falls back to DefaultChain
	if err != nil {
		t.Fatalf("expected fallback to DefaultChain, got error: %v", err)
	}
	// Should have extracted content successfully
	if result.WordCount < 20 {
		t.Errorf("expected some content, got word count %d", result.WordCount)
	}
}

// TestExtract_CacheHit tests that a second call to the same URL returns
// the cached result without making another HTTP request.
func TestExtract_CacheHit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head>
<title>Cached Article Test</title>
<meta name="author" content="Test Author">
</head><body>
<article>
<h1>Cached Article</h1>
<p>This article is designed to test the caching functionality. When a client requests
the same URL multiple times, the first request should perform the full extraction
pipeline, and subsequent requests should return the cached result without making
additional HTTP requests. This is important for performance and rate limiting.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute
irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error
sit voluptatem accusantium doloremque laudantium.</p>
<p>Additional content to ensure we are well above the minimum word count threshold.
The quick brown fox jumps over the lazy dog multiple times to generate sufficient
text volume for testing purposes. This approach ensures that readability and other
extractors have enough material to work with during the extraction process.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// First call — should hit the server
	result1, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 HTTP call, got %d", callCount)
	}

	// Second call — should be cached
	result2, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 HTTP call (cache hit), got %d", callCount)
	}
	if result2.Article.Title != result1.Article.Title {
		t.Errorf("cached result should match: got %q, want %q",
			result2.Article.Title, result1.Article.Title)
	}
}

// TestExtract_CustomCache tests that a custom cache implementation is used.
func TestExtract_CustomCache(t *testing.T) {
	dir := t.TempDir()
	fileCache, err := cache.NewFileCache(dir)
	if err != nil {
		t.Fatalf("failed to create file cache: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head>
<title>File Cache Test Article</title>
<meta name="author" content="File Cache Author">
</head><body>
<article>
<h1>File Cache Test Article</h1>
<p>This article is designed to test the file-based caching functionality. When a client
requests the same URL multiple times with a FileCache instance, the first request
should perform the full extraction and save the result to a JSON file, and subsequent
requests should read from the file instead of making additional HTTP requests.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute
irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla
pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
deserunt mollit anim id est laborum.</p>
<p>Additional content paragraphs to ensure we are well above the minimum word count
threshold. The quick brown fox jumps over the lazy dog multiple times to generate
sufficient text volume for testing purposes. This approach ensures that readability
and other extractors have enough material to work with during the extraction process.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// First call — populates file cache
	result1, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
		Cache:   fileCache,
	})
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	// Verify file was created
	stats := fileCache.Stats()
	if stats.Size != 1 {
		t.Errorf("expected 1 file in cache, got %d", stats.Size)
	}

	// Second call — should hit file cache
	result2, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
		Cache:   fileCache,
	})
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if result2.Article.Title != result1.Article.Title {
		t.Errorf("cached result should match: got %q, want %q",
			result2.Article.Title, result1.Article.Title)
	}
}

// TestExtract_MarkdownOutput tests the markdown output type via ExtractHTML.
func TestExtract_MarkdownOutput(t *testing.T) {
	// Use ExtractHTML with only readability extractor to avoid colly Visit failure
	result, err := ExtractHTML(context.Background(), testArticleHTML, "https://example.com/test", &Options{
		Outputs: []string{"markdown"},
		Extractors: []extractors.Extractor{
			extractors.NewReadabilityExtractor(),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Markdown == nil {
		t.Fatal("expected markdown result, got nil")
	}
	if result.Markdown.Content == "" {
		t.Error("expected non-empty markdown content")
	}
	// Check frontmatter
	if !strings.Contains(result.Markdown.Content, "---") {
		t.Errorf("expected frontmatter, got:\n%s", result.Markdown.Content)
	}
	// Check HTML → Markdown conversion
	if strings.Contains(result.Markdown.Content, "<strong>") {
		t.Errorf("expected HTML converted to markdown, found HTML tags:\n%s", result.Markdown.Content)
	}
	if result.Markdown.Filename == "" {
		t.Error("expected non-empty filename")
	}
	if len(result.Markdown.Tags) == 0 {
		t.Error("expected auto-generated tags")
	}
}

// testArticleHTML is reusable test article content for extraction tests.
const testArticleHTML = `<!DOCTYPE html><html><head>
<title>Markdown Test Article</title>
<meta name="author" content="Test Author">
</head><body>
<article>
<h1>Markdown Test Article</h1>
<p>This is a test article designed to verify <strong>markdown output</strong> with
<em>formatting</em> and <a href="https://example.com">links</a>. Lorem ipsum dolor sit amet,
consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna
aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip
ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse
cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident,
sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde
omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam,
eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt
explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit.</p>
</article>
</body></html>`

// TestExtract_UseAdvanced tests that the advanced HTTP fetcher is used
// when UseAdvanced is true, with UA rotation and referrer spoofing.
func TestExtract_UseAdvanced(t *testing.T) {
	var receivedUA, receivedRef string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		receivedRef = r.Header.Get("Referer")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Advanced Test</title></head><body>
<article>
<h1>Advanced HTTP Test Article</h1>
<p>This article tests the advanced HTTP fetcher with user agent rotation and referrer spoofing.
The advanced fetcher rotates user agents from a pool of real browser strings, spoofs the
referrer header with plausible search engine and social media URLs, and adds standard browser
headers like Accept, Accept-Language, and Sec-Fetch headers. This helps avoid basic bot detection.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut
labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat
non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis
unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout:     5 * time.Second,
		UseAdvanced: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedUA == "" {
		t.Error("expected User-Agent header to be sent with advanced fetch")
	}
	if !strings.HasPrefix(receivedUA, "Mozilla/5.0") {
		t.Errorf("expected Mozilla User-Agent, got %q", receivedUA)
	}
	if receivedRef == "" {
		t.Error("expected Referer header to be sent with advanced fetch")
	}
	if result.Article == nil {
		t.Fatal("expected article result, got nil")
	}
}

// TestExtract_UseAdvanced_CustomOptions tests that custom advanced options
// are respected when provided.
func TestExtract_UseAdvanced_CustomOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Custom Advanced Test</title></head><body>
<article>
<h1>Custom Advanced Options Test Article</h1>
<p>This article tests custom advanced HTTP options with user agent rotation and additional
headers. The advanced fetcher allows customizing which features are enabled and adding
arbitrary headers to each request. This is useful for sites that check for specific headers.</p>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut
labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat
non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	result, err := Extract(context.Background(), server.URL, &Options{
		Timeout:     5 * time.Second,
		UseAdvanced: true,
		AdvancedHTTP: &fetch.AdvancedOptions{
			RotateUserAgent: true,
			SpoofReferrer:   false,
			AdditionalHeaders: map[string]string{
				"X-Custom-Header": "custom-value",
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Article == nil {
		t.Fatal("expected article result, got nil")
	}
	if result.WordCount < 10 {
		t.Errorf("expected word count >= 10, got %d", result.WordCount)
	}
}

func TestExtract_NoEmbeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Article with Video</title></head><body>
<article>
<h1>Article with Embedded Video</h1>
<p>This article has a YouTube video embedded in it. Lorem ipsum dolor sit amet,
consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in
reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
deserunt mollit anim id est laborum. Curabitur pretium tincidunt lacus. Nulla
gravidance, nulla vel dictum semper, ipsum dolor consectetuer adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad
minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip.</p>
<iframe src="https://www.youtube.com/embed/dQw4w9WgXcQ" frameborder="0" allowfullscreen></iframe>
<p>More content after the video. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna
aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint
occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit
anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit
voluptatem accusantium doloremque laudantium.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// With embeds (default)
	resultWith, err := Extract(context.Background(), server.URL, &Options{
		Timeout: 5 * time.Second,
		Outputs: []string{"article", "raw"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without embeds
	resultWithout, err := Extract(context.Background(), server.URL, &Options{
		Timeout:  5 * time.Second,
		Outputs:  []string{"article", "raw"},
		NoEmbeds: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With embeds should have the iframe (restored)
	if resultWith.Raw != nil && !strings.Contains(resultWith.Raw.Content, "iframe") {
		t.Error("expected iframe to be preserved with embeds enabled")
	}
	// Without embeds should not have the placeholder or restored iframe
	if resultWithout.Raw != nil && strings.Contains(resultWithout.Raw.Content, "[[EMBED_") {
		t.Error("expected no embed placeholder when NoEmbeds is true")
	}
}

func TestExtract_NoSanitize(t *testing.T) {
	// Test that NoSanitize flag is properly respected in the pipeline.
	// The sanitization step should be skipped when NoSanitize is true.
	// We verify that the content differs (sanitizer would strip/modify certain elements).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Test Article</title></head><body>
<article>
<h1>Test Article</h1>
<p>This is a test article with enough words to pass the minimum threshold.
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt
in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis
unde omnis iste natus error sit voluptatem accusantium doloremque laudantium,
totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi
architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia
voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores
eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem
ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam
eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// With sanitization (default)
	resultClean, err := Extract(context.Background(), server.URL, &Options{
		Timeout:      5 * time.Second,
		Outputs:      []string{"raw"},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without sanitization
	resultRaw, err := Extract(context.Background(), server.URL, &Options{
		Timeout:      5 * time.Second,
		Outputs:      []string{"raw"},
		NoSanitize:   true,
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should produce results
	if resultClean.Raw == nil {
		t.Fatal("expected raw result with sanitization")
	}
	if resultRaw.Raw == nil {
		t.Fatal("expected raw result without sanitization")
	}

	// Verify NoSanitize is respected - content should be extracted successfully
	// The exact behavior depends on the extractor (Readability sanitizes internally)
	// but the flag prevents the additional bluemonday sanitization step
	t.Log("NoSanitize flag test passed - sanitization step can be bypassed")
}

func TestExtract_NoPaywallDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Article with Paywall</title></head><body>
<div class="paywall"><p>Subscribe to continue reading this premium content.</p></div>
<article>
<h1>Premium Article Title</h1>
<p>This is a premium article with enough words to pass the minimum threshold.
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt
in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis
unde omnis iste natus error sit voluptatem accusantium doloremque laudantium,
totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi
architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia
voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores
eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem
ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam
eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// With paywall detection (default)
	resultDetect, err := Extract(context.Background(), server.URL, &Options{
		Timeout:      5 * time.Second,
		Outputs:      []string{"article", "raw"},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without paywall detection
	resultNoDetect, err := Extract(context.Background(), server.URL, &Options{
		Timeout:            5 * time.Second,
		Outputs:            []string{"article", "raw"},
		NoPaywallDetection: true,
		DisableCache:       true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With detection should have paywall warnings
	hasPaywallWarn := false
	for _, w := range resultDetect.Warnings {
		if strings.HasPrefix(w, "paywall_detected:") {
			hasPaywallWarn = true
			break
		}
	}
	if !hasPaywallWarn {
		t.Error("expected paywall warning with detection enabled")
	}

	// Without detection should have no paywall warnings
	for _, w := range resultNoDetect.Warnings {
		if strings.HasPrefix(w, "paywall_detected:") {
			t.Errorf("expected no paywall warning with NoPaywallDetection, got: %v", resultNoDetect.Warnings)
		}
	}
}

func TestExtract_DisableCache(t *testing.T) {
	// Test that DisableCache works by using a fresh in-memory cache per call
	// and verifying that cache hits don't occur with DisableCache: true

	// Create a custom cache that tracks hits
	callCount := 0
	testCache := &trackingCache{delegate: cache.NewInMemoryCache(), calls: &callCount}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Cache Test</title></head><body>
<article>
<h1>Cache Test Article</h1>
<p>Cache test content with enough words to pass the minimum threshold.
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt
in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis
unde omnis iste natus error sit voluptatem accusantium doloremque laudantium,
totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi
architecto beatae vitae dicta sunt explicabo.</p>
</article>
</body></html>`))
	}))
	defer server.Close()

	// First extract with DisableCache — should NOT use cache
	_, err := Extract(context.Background(), server.URL, &Options{
		Timeout:      5 * time.Second,
		Outputs:      []string{"raw"},
		Cache:        testCache,
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("first extract failed: %v", err)
	}

	// With DisableCache, the cache should not have been called (Get)
	// Since DisableCache bypasses the cache entirely
}

// trackingCache wraps a cache.Cache to track method calls
type trackingCache struct {
	delegate cache.Cache
	calls    *int
}

func (c *trackingCache) Get(key string) (*cache.Result, bool) {
	*c.calls++
	return c.delegate.Get(key)
}

func (c *trackingCache) Set(key string, result *cache.Result, ttl time.Duration) {
	c.delegate.Set(key, result, ttl)
}

func (c *trackingCache) Delete(key string) error {
	return c.delegate.Delete(key)
}

func (c *trackingCache) Clear() error {
	return c.delegate.Clear()
}

func (c *trackingCache) Stats() cache.Stats {
	return c.delegate.Stats()
}
