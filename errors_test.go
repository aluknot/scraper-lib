package scraperlib

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aluknot/scraper-lib/types"
)

func TestErrFetchFailed_Error(t *testing.T) {
	err := ErrFetchFailed{
		URL: "https://example.com/test",
		Attempts: []types.StrategyAttempt{
			{Strategy: "http_simple", Status: "error", DurationMs: 100, Error: "timeout"},
		},
		LastError: fmt.Errorf("connection timeout"),
	}

	msg := err.Error()
	if !strings.Contains(msg, "fetch failed") {
		t.Errorf("expected error message to contain 'fetch failed', got %q", msg)
	}
	if !strings.Contains(msg, "https://example.com/test") {
		t.Errorf("expected error message to contain URL, got %q", msg)
	}
}

func TestErrAllStrategiesFailed_Error(t *testing.T) {
	err := ErrAllStrategiesFailed{
		URL: "https://example.com/test",
		Attempts: []types.StrategyAttempt{
			{Strategy: "http_simple", Status: "error"},
			{Strategy: "http_advanced", Status: "blocked"},
		},
	}

	msg := err.Error()
	if !strings.Contains(msg, "all strategies failed") {
		t.Errorf("expected 'all strategies failed', got %q", msg)
	}
	if !strings.Contains(msg, "https://example.com/test") {
		t.Errorf("expected URL in error message, got %q", msg)
	}
	if !strings.Contains(msg, "2 attempts") {
		t.Errorf("expected attempt count, got %q", msg)
	}
}

func TestErrCircuitOpen_Error(t *testing.T) {
	err := ErrCircuitOpen{
		Domain:            "example.com",
		CooldownRemaining: 2 * time.Minute,
	}

	msg := err.Error()
	if !strings.Contains(msg, "circuit breaker open") {
		t.Errorf("expected 'circuit breaker open', got %q", msg)
	}
	if !strings.Contains(msg, "example.com") {
		t.Errorf("expected domain in error message, got %q", msg)
	}
}

func TestErrFetchFailed_IsTarget(t *testing.T) {
	err := ErrFetchFailed{
		URL:       "https://example.com",
		LastError: fmt.Errorf("timeout"),
	}

	var fetchErr ErrFetchFailed
	if !errors.As(err, &fetchErr) {
		t.Error("expected errors.As to match ErrFetchFailed")
	}
	if fetchErr.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", fetchErr.URL)
	}
}

func TestErrAllStrategiesFailed_IsTarget(t *testing.T) {
	err := ErrAllStrategiesFailed{
		URL: "https://example.com",
	}

	var stratErr ErrAllStrategiesFailed
	if !errors.As(err, &stratErr) {
		t.Error("expected errors.As to match ErrAllStrategiesFailed")
	}
	if stratErr.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", stratErr.URL)
	}
}
