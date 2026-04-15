package cache

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryCache_GetSet(t *testing.T) {
	c := NewInMemoryCache()

	result := &Result{URL: "https://example.com", Title: "Test", Content: "Hello"}
	c.Set("https://example.com", result, 1*time.Hour)

	got, ok := c.Get("https://example.com")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", got.Title)
	}
	if got.Content != "Hello" {
		t.Errorf("expected content 'Hello', got %q", got.Content)
	}
}

func TestInMemoryCache_Miss(t *testing.T) {
	c := NewInMemoryCache()

	_, ok := c.Get("https://missing.com")
	if ok {
		t.Error("expected cache miss")
	}

	stats := c.Stats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestInMemoryCache_TTLExpiration(t *testing.T) {
	c := NewInMemoryCache()

	result := &Result{URL: "https://example.com", Title: "Expiring"}
	c.Set("https://example.com", result, 50*time.Millisecond)

	// Should be present immediately after set
	_, ok := c.Get("https://example.com")
	if !ok {
		t.Fatal("expected hit before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	_, ok = c.Get("https://example.com")
	if ok {
		t.Error("expected miss after TTL expiration")
	}

	stats := c.Stats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss after expiration, got %d", stats.Misses)
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	c := NewInMemoryCache()

	result := &Result{URL: "https://example.com", Title: "ToDelete"}
	c.Set("https://example.com", result, 1*time.Hour)

	if err := c.Delete("https://example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := c.Get("https://example.com")
	if ok {
		t.Error("expected miss after delete")
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	c := NewInMemoryCache()

	c.Set("https://a.com", &Result{URL: "https://a.com"}, 1*time.Hour)
	c.Set("https://b.com", &Result{URL: "https://b.com"}, 1*time.Hour)

	if err := c.Clear(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("expected size 0 after clear, got %d", stats.Size)
	}
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("expected stats reset after clear, got hits=%d misses=%d", stats.Hits, stats.Misses)
	}
}

func TestInMemoryCache_Stats(t *testing.T) {
	c := NewInMemoryCache()

	c.Set("https://a.com", &Result{URL: "https://a.com"}, 1*time.Hour)
	c.Set("https://b.com", &Result{URL: "https://b.com"}, 1*time.Hour)

	// 1 hit + 1 miss
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

func TestInMemoryCache_ConcurrentAccess(t *testing.T) {
	c := NewInMemoryCache()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url := "https://example.com/" + string(rune(i))
			c.Set(url, &Result{URL: url}, 1*time.Hour)
			c.Get(url)
		}(i)
	}
	wg.Wait()

	stats := c.Stats()
	if stats.Size != 100 {
		t.Errorf("expected 100 entries after concurrent writes, got %d", stats.Size)
	}
}
