// Package fetch provides HTTP fetching with retry logic and backoff.
package fetch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	MaxRetries     = 3
	baseRetryDelay = 1 * time.Second
	maxRetryDelay  = 10 * time.Second
)

// calculateBackoff returns the delay for exponential backoff.
func calculateBackoff(attempt int) time.Duration {
	delay := baseRetryDelay * time.Duration(1<<attempt)
	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}
	return delay
}

// isRetriable checks if an error should trigger a retry.
func isRetriable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	// Network errors
	if containsAny(msg,
		"timeout",
		"connection refused",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"temporary failure",
		"server misbehaving",
	) {
		return true
	}

	// HTTP errors that are temporary
	if containsAny(msg, "429", "503", "502", "504") {
		return true
	}

	return false
}

// isPermanent checks if an error should NOT trigger a retry.
func isPermanent(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	// Client errors that won't be fixed by retrying
	if containsAny(msg, "http 401", "http 403", "http 404", "http 400") {
		return true
	}

	// Parse errors are usually permanent
	if containsAny(msg, "parse") {
		return true
	}

	return false
}

// FetchWithRetry performs an HTTP request with exponential backoff retry.
// Returns the raw HTML body on success.
func FetchWithRetry(ctx context.Context, client *http.Client, req *http.Request) (string, error) {
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		resp, err := client.Do(req)

		// Success
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// Close body when this function returns
				defer resp.Body.Close()

				buf := make([]byte, 0, 1024*512) // Preallocate 512KB
				for {
					chunk := make([]byte, 32*1024)
					n, readErr := resp.Body.Read(chunk)
					if n > 0 {
						buf = append(buf, chunk[:n]...)
					}
					if readErr != nil {
						break
					}
				}
				return string(buf), nil
			}

			// For 429 (rate limited), retry with delay
			if resp.StatusCode == 429 {
				lastErr = fmt.Errorf("http %d", resp.StatusCode)
				resp.Body.Close()
				delay := calculateBackoff(attempt)
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(delay):
					continue
				}
			}

			// For 5xx (server errors), retry
			if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				lastErr = fmt.Errorf("http %d", resp.StatusCode)
				resp.Body.Close()
				delay := calculateBackoff(attempt)
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(delay):
					continue
				}
			}

			// For other status codes, return the error
			resp.Body.Close()
			return "", fmt.Errorf("http %d", resp.StatusCode)
		}

		lastErr = err

		// Check if error is retriable
		if !isRetriable(err) || isPermanent(err) {
			return "", err
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt)

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return "", fmt.Errorf("max retries (%d) exceeded: %w", MaxRetries, lastErr)
}

// GetHTML fetches the HTML content from a URL with retry logic.
func GetHTML(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ScraperLib/1.0)")

	return FetchWithRetry(ctx, client, req)
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
