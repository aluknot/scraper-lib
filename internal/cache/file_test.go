package cache

import (
	"os"
	"testing"
	"time"
)

func TestFileCache_GetSet(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	result := &Result{URL: "https://example.com", Title: "FileTest", Content: "Hello from file"}
	c.Set("https://example.com", result, 1*time.Hour)

	got, ok := c.Get("https://example.com")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Title != "FileTest" {
		t.Errorf("expected title 'FileTest', got %q", got.Title)
	}
	if got.Content != "Hello from file" {
		t.Errorf("expected content 'Hello from file', got %q", got.Content)
	}
}

func TestFileCache_Miss(t *testing.T) {
	dir := t.TempDir()
	c, _ := NewFileCache(dir)

	_, ok := c.Get("https://missing.com")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestFileCache_TTLExpiration(t *testing.T) {
	dir := t.TempDir()
	c, _ := NewFileCache(dir)

	result := &Result{URL: "https://example.com", Title: "Expiring"}
	c.Set("https://example.com", result, 100*time.Millisecond)

	// Should be present immediately
	got, ok := c.Get("https://example.com")
	if !ok {
		t.Fatalf("expected hit before expiration")
	}
	if got.Title != "Expiring" {
		t.Errorf("expected title 'Expiring', got %q", got.Title)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	_, ok = c.Get("https://example.com")
	if ok {
		t.Error("expected miss after TTL expiration")
	}
}

func TestFileCache_Delete(t *testing.T) {
	dir := t.TempDir()
	c, _ := NewFileCache(dir)

	c.Set("https://example.com", &Result{URL: "https://example.com"}, 1*time.Hour)

	if err := c.Delete("https://example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := c.Get("https://example.com")
	if ok {
		t.Error("expected miss after delete")
	}
}

func TestFileCache_Clear(t *testing.T) {
	dir := t.TempDir()
	c, _ := NewFileCache(dir)

	c.Set("https://a.com", &Result{URL: "https://a.com"}, 1*time.Hour)
	c.Set("https://b.com", &Result{URL: "https://b.com"}, 1*time.Hour)

	if err := c.Clear(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("expected size 0 after clear, got %d", stats.Size)
	}
}

func TestFileCache_Stats(t *testing.T) {
	dir := t.TempDir()
	c, _ := NewFileCache(dir)

	c.Set("https://a.com", &Result{URL: "https://a.com"}, 1*time.Hour)
	c.Set("https://b.com", &Result{URL: "https://b.com"}, 1*time.Hour)

	c.Get("https://a.com")
	c.Get("https://missing.com")

	stats := c.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 2 {
		t.Errorf("expected size 2, got %d", stats.Size)
	}
}

func TestFileCache_CreatesDir(t *testing.T) {
	dir := t.TempDir() + "/subdir/cache"
	c, err := NewFileCache(dir)
	if err != nil {
		t.Fatalf("expected dir creation to succeed, got error: %v", err)
	}

	_, err = os.Stat(dir)
	if err != nil {
		t.Errorf("expected dir to exist, got error: %v", err)
	}
	_ = c
}

func TestFileCache_DefaultDir(t *testing.T) {
	c, err := NewFileCache("")
	if err != nil {
		t.Fatalf("expected default dir to succeed, got error: %v", err)
	}

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("expected empty default cache, got size %d", stats.Size)
	}
}
