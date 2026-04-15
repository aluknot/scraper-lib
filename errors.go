package scraperlib

import (
	"fmt"
	"time"

	"github.com/aluknot/scraper-lib/types"
)

// ErrFetchFailed is returned when the HTTP fetcher cannot retrieve the page
// after all retries.
type ErrFetchFailed struct {
	URL       string
	Attempts  []types.StrategyAttempt
	LastError error
}

func (e ErrFetchFailed) Error() string {
	return fmt.Sprintf("fetch failed for %s: %v", e.URL, e.LastError)
}

// ErrAllStrategiesFailed is returned when all fetch strategies have been
// exhausted without success (e.g., http_simple, http_advanced, browser, archive).
type ErrAllStrategiesFailed struct {
	URL      string
	Attempts []types.StrategyAttempt
}

func (e ErrAllStrategiesFailed) Error() string {
	return fmt.Sprintf("all strategies failed for %s (%d attempts)", e.URL, len(e.Attempts))
}

// ErrCircuitOpen is returned when the circuit breaker for a domain is open,
// meaning recent failures have triggered a cooldown period.
type ErrCircuitOpen struct {
	Domain            string
	CooldownRemaining time.Duration
}

func (e ErrCircuitOpen) Error() string {
	return fmt.Sprintf("circuit breaker open for %s, retry in %v", e.Domain, e.CooldownRemaining)
}
