// Package fetch provides HTTP fetching with retry logic and backoff.
package fetch

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// userAgentPool lists common browser user agents for rotation.
var userAgentPool = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0",
}

// referrerPool lists plausible referrer URLs for header spoofing.
var referrerPool = []string{
	"https://www.google.com/",
	"https://www.google.com/search?q=example",
	"https://www.bing.com/search?q=example",
	"https://t.co/shortlink",
	"https://www.reddit.com/",
	"https://news.ycombinator.com/",
	"https://www.facebook.com/",
	"https://twitter.com/",
}

// acceptLanguages lists common Accept-Language header values.
var acceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-US,en;q=0.9,es;q=0.8",
	"en-GB,en;q=0.9,en-US;q=0.8",
	"es-ES,es;q=0.9,en;q=0.8",
	"en-US,en;q=0.5",
}

// randomUserAgent returns a random user agent from the pool.
func randomUserAgent() string {
	return userAgentPool[rand.Intn(len(userAgentPool))]
}

// randomReferrer returns a random referrer from the pool.
func randomReferrer() string {
	return referrerPool[rand.Intn(len(referrerPool))]
}

// randomAcceptLanguage returns a random Accept-Language value.
func randomAcceptLanguage() string {
	return acceptLanguages[rand.Intn(len(acceptLanguages))]
}

// calculateJitter returns a random delay between 0 and the given delay.
func calculateJitter(delay time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(delay)))
}

// AdvancedOptions configures the advanced HTTP fetcher.
type AdvancedOptions struct {
	// RotateUserAgent enables random user agent selection from a pool.
	RotateUserAgent bool

	// SpoofReferrer sets a plausible Referer header from a pool.
	SpoofReferrer bool

	// SpoofAcceptLanguage sets a random Accept-Language header.
	SpoofAcceptLanguage bool

	// EnableCookies enables a cookie jar to maintain session state.
	EnableCookies bool

	// AdditionalHeaders are extra headers to include in the request.
	AdditionalHeaders map[string]string
}

// NewHTTPClientAdvanced creates an http.Client with advanced options.
// When rotateUA is true, each request will use a random user agent.
// When enableCookies is true, a cookie jar is created to persist cookies.
func NewHTTPClientAdvanced(timeout time.Duration, enableCookies bool) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	if enableCookies {
		jar, err := cookiejar.New(nil)
		if err == nil {
			client.Jar = jar
		}
	}

	return client
}

// ApplyAdvancedOptions applies advanced options to the request.
// Mutates req in place.
func ApplyAdvancedOptions(req *http.Request, opts *AdvancedOptions) {
	if opts == nil {
		return
	}

	if opts.RotateUserAgent {
		req.Header.Set("User-Agent", randomUserAgent())
	}

	if opts.SpoofReferrer {
		req.Header.Set("Referer", randomReferrer())
	}

	if opts.SpoofAcceptLanguage {
		req.Header.Set("Accept-Language", randomAcceptLanguage())
	}

	// Common browser headers that sites expect
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	for k, v := range opts.AdditionalHeaders {
		req.Header.Set(k, v)
	}
}

// GetHTMLAdvanced fetches the HTML content from a URL with advanced options.
// It uses user agent rotation, cookie jars, and referrer spoofing when enabled.
// Falls back to FetchWithRetry for the actual HTTP request with backoff.
func GetHTMLAdvanced(ctx context.Context, client *http.Client, url string, opts *AdvancedOptions) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	// Apply advanced options (UA, referrer, etc.)
	ApplyAdvancedOptions(req, opts)

	// If no UA rotation, use a default scraper UA
	if opts == nil || !opts.RotateUserAgent {
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ScraperLib/1.0)")
	}

	return FetchWithRetry(ctx, client, req)
}

// NewRequestWithUA creates a request with a random user agent and common browser headers.
// Useful when the caller needs direct control over the request.
func NewRequestWithUA(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	ApplyAdvancedOptions(req, &AdvancedOptions{
		RotateUserAgent:     true,
		SpoofReferrer:       true,
		SpoofAcceptLanguage: true,
	})
	return req, nil
}

// ParseURL safely parses a URL string, returning nil on error.
func ParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	return u
}
