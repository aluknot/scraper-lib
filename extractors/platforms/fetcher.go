package platforms

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) (*goquery.Document, *http.Response, error)
}

type GoQueryFetcher struct {
	Client *http.Client
}

func NewGoQueryFetcher(client *http.Client) *GoQueryFetcher {
	if client == nil {
		client = &http.Client{}
	}
	return &GoQueryFetcher{Client: client}
}

func (f *GoQueryFetcher) Fetch(ctx context.Context, url string) (*goquery.Document, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; scraper-lib/1.0)")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("http %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("parse html: %w", err)
	}

	return doc, resp, nil
}

// InMemoryFetcher parses pre-fetched HTML without making HTTP requests.
// Used when the caller already has the HTML (e.g., from scraperlib.Extract).
type InMemoryFetcher struct {
	html string
}

func NewInMemoryFetcher(html string) *InMemoryFetcher {
	return &InMemoryFetcher{html: html}
}

func (f *InMemoryFetcher) Fetch(ctx context.Context, url string) (*goquery.Document, *http.Response, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(f.html))
	if err != nil {
		return nil, nil, fmt.Errorf("parse html: %w", err)
	}
	return doc, nil, nil
}

func CanProcess(url string, domains []string) bool {
	lowerURL := strings.ToLower(url)
	for _, domain := range domains {
		if strings.Contains(lowerURL, strings.ToLower(domain)) {
			return true
		}
	}
	return false
}
