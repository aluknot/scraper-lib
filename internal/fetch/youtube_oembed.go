package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// OEmbedResponse represents the YouTube oEmbed API response.
type OEmbedResponse struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	ProviderName string `json:"provider_name"`
	Type         string `json:"type"`
	HTML         string `json:"html"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

// OEmbedFetcher fetches metadata from YouTube's oEmbed endpoint.
type OEmbedFetcher struct {
	client *http.Client
}

// NewOEmbedFetcher creates a new OEmbedFetcher with the given HTTP client.
func NewOEmbedFetcher(client *http.Client) *OEmbedFetcher {
	if client == nil {
		client = &http.Client{}
	}
	return &OEmbedFetcher{client: client}
}

// Fetch retrieves oEmbed data for a YouTube URL.
func (f *OEmbedFetcher) Fetch(ctx context.Context, videoURL string) (*OEmbedResponse, error) {
	oembedURL := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json",
		url.QueryEscape(videoURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, oembedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create oembed request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; scraper-lib/1.0)")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch oembed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oembed HTTP %d: %s", resp.StatusCode, string(body))
	}

	var oembed OEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&oembed); err != nil {
		return nil, fmt.Errorf("decode oembed response: %w", err)
	}

	return &oembed, nil
}
