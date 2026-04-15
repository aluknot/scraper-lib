package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRandomUserAgent_NotEmpty(t *testing.T) {
	ua := randomUserAgent()
	if ua == "" {
		t.Error("expected non-empty user agent")
	}
	if !strings.HasPrefix(ua, "Mozilla/5.0") {
		t.Errorf("expected Mozilla user agent prefix, got %q", ua)
	}
}

func TestRandomUserAgent_Variety(t *testing.T) {
	// Run multiple times to ensure we get variety (not always the same)
	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		ua := randomUserAgent()
		seen[ua] = true
	}
	// With 8 user agents and 20 samples, we should see at least 2 different ones
	if len(seen) < 2 {
		t.Errorf("expected user agent variety, got only %d unique from 20 samples", len(seen))
	}
}

func TestRandomReferrer_NotEmpty(t *testing.T) {
	ref := randomReferrer()
	if ref == "" {
		t.Error("expected non-empty referrer")
	}
	if !strings.HasPrefix(ref, "https://") {
		t.Errorf("expected https referrer, got %q", ref)
	}
}

func TestApplyAdvancedOptions_UserAgent(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	ApplyAdvancedOptions(req, &AdvancedOptions{
		RotateUserAgent: true,
	})

	ua := req.Header.Get("User-Agent")
	if ua == "" {
		t.Error("expected User-Agent header to be set")
	}
	if !strings.HasPrefix(ua, "Mozilla/5.0") {
		t.Errorf("expected Mozilla user agent, got %q", ua)
	}
}

func TestApplyAdvancedOptions_Referrer(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	ApplyAdvancedOptions(req, &AdvancedOptions{
		SpoofReferrer: true,
	})

	ref := req.Header.Get("Referer")
	if ref == "" {
		t.Error("expected Referer header to be set")
	}
}

func TestApplyAdvancedOptions_AcceptLanguage(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	ApplyAdvancedOptions(req, &AdvancedOptions{
		SpoofAcceptLanguage: true,
	})

	al := req.Header.Get("Accept-Language")
	if al == "" {
		t.Error("expected Accept-Language header to be set")
	}
}

func TestApplyAdvancedOptions_AdditionalHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	ApplyAdvancedOptions(req, &AdvancedOptions{
		AdditionalHeaders: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	})

	if req.Header.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected custom header, got %q", req.Header.Get("X-Custom-Header"))
	}
}

func TestApplyAdvancedOptions_NilOptions(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	req.Header.Set("User-Agent", "test-agent")
	ApplyAdvancedOptions(req, nil)

	// Should not modify User-Agent when opts is nil
	if req.Header.Get("User-Agent") != "test-agent" {
		t.Errorf("expected User-Agent unchanged, got %q", req.Header.Get("User-Agent"))
	}
}

func TestNewHTTPClientAdvanced_WithCookies(t *testing.T) {
	client := NewHTTPClientAdvanced(5*time.Second, true)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Jar == nil {
		t.Error("expected cookie jar to be set")
	}
	if client.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", client.Timeout)
	}
}

func TestNewHTTPClientAdvanced_NoCookies(t *testing.T) {
	client := NewHTTPClientAdvanced(5*time.Second, false)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Jar != nil {
		t.Error("expected no cookie jar")
	}
}

func TestGetHTMLAdvanced_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><p>Advanced fetch test content with enough words to pass validation. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident.</p></body></html>`))
	}))
	defer server.Close()

	ctx := context.Background()
	client := NewHTTPClientAdvanced(5*time.Second, true)

	html, err := GetHTMLAdvanced(ctx, client, server.URL, &AdvancedOptions{
		RotateUserAgent:     true,
		SpoofReferrer:       true,
		SpoofAcceptLanguage: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Advanced fetch test content") {
		t.Error("expected test content in response")
	}
}

func TestGetHTMLAdvanced_HeadersReceived(t *testing.T) {
	var receivedUA, receivedRef string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		receivedRef = r.Header.Get("Referer")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><p>OK</p></body></html>`))
	}))
	defer server.Close()

	ctx := context.Background()
	client := NewHTTPClientAdvanced(5*time.Second, true)

	_, err := GetHTMLAdvanced(ctx, client, server.URL, &AdvancedOptions{
		RotateUserAgent: true,
		SpoofReferrer:   true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedUA == "" {
		t.Error("expected User-Agent header to be sent")
	}
	if receivedRef == "" {
		t.Error("expected Referer header to be sent")
	}
}

func TestGetHTMLAdvanced_CookiePersistence(t *testing.T) {
	cookieSet := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cookieSet {
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: "abc123",
				Path:  "/",
			})
			cookieSet = true
		}
		// Check if cookie was sent back
		cookies := r.Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "session_id" && c.Value == "abc123" {
				found = true
				break
			}
		}
		w.WriteHeader(http.StatusOK)
		if found {
			w.Write([]byte(`<html><body><p>Cookie received and sent back. Enough words for validation. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p></body></html>`))
		} else {
			w.Write([]byte(`<html><body><p>No cookie yet. Enough words for validation. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p></body></html>`))
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := NewHTTPClientAdvanced(5*time.Second, true)

	// First request — server sets cookie
	_, err := GetHTMLAdvanced(ctx, client, server.URL, &AdvancedOptions{
		RotateUserAgent: true,
	})
	if err != nil {
		t.Fatalf("first request error: %v", err)
	}

	// Second request — client should send cookie back
	html, err := GetHTMLAdvanced(ctx, client, server.URL, &AdvancedOptions{
		RotateUserAgent: true,
	})
	if err != nil {
		t.Fatalf("second request error: %v", err)
	}
	if !strings.Contains(html, "Cookie received") {
		t.Error("expected cookie to be persisted on second request")
	}
}

func TestNewRequestWithUA(t *testing.T) {
	ctx := context.Background()
	req, err := NewRequestWithUA(ctx, "GET", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ua := req.Header.Get("User-Agent")
	if ua == "" {
		t.Error("expected User-Agent to be set")
	}
	ref := req.Header.Get("Referer")
	if ref == "" {
		t.Error("expected Referer to be set")
	}
	al := req.Header.Get("Accept-Language")
	if al == "" {
		t.Error("expected Accept-Language to be set")
	}
}

func TestCalculateJitter(t *testing.T) {
	delay := 100 * time.Millisecond
	jitter := calculateJitter(delay)
	if jitter < 0 {
		t.Errorf("expected non-negative jitter, got %v", jitter)
	}
	if jitter > delay {
		t.Errorf("expected jitter <= delay, got %v > %v", jitter, delay)
	}
}

func TestParseURL_Valid(t *testing.T) {
	u := ParseURL("https://example.com/path?query=value")
	if u == nil {
		t.Fatal("expected non-nil URL")
	}
	if u.Host != "example.com" {
		t.Errorf("expected host 'example.com', got %q", u.Host)
	}
}

func TestParseURL_Invalid(t *testing.T) {
	u := ParseURL("://invalid-url")
	if u != nil {
		t.Errorf("expected nil for invalid URL, got %v", u)
	}
}
