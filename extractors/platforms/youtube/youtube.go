package youtube

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aluknot/scraper-lib/extractors/platforms"
	"github.com/aluknot/scraper-lib/internal/fetch"
)

// MetadataSource defines a source for fetching metadata.
type MetadataSource interface {
	FetchMetadata(ctx context.Context, url string) (*VideoMetadata, error)
	Priority() int
	Name() string
}

type Extractor struct {
	client        *http.Client
	fetcher       platforms.Fetcher
	oembedFetcher *fetch.OEmbedFetcher
	sources       []MetadataSource
}

func New(client *http.Client) *Extractor {
	if client == nil {
		client = &http.Client{}
	}
	e := &Extractor{
		client:        client,
		fetcher:       platforms.NewGoQueryFetcher(client),
		oembedFetcher: fetch.NewOEmbedFetcher(client),
	}
	e.initSources()
	return e
}

func NewWithFetcher(fetcher platforms.Fetcher) *Extractor {
	e := &Extractor{
		fetcher: fetcher,
	}
	e.initSources()
	return e
}

func (e *Extractor) initSources() {
	// oEmbed first (no IP blocks, fast) - only if fetcher is available
	if e.oembedFetcher != nil {
		e.sources = append(e.sources, &oembedSource{fetcher: e.oembedFetcher})
	}
	// HTML parse as fallback (may get 429)
	e.sources = append(e.sources, &htmlSource{fetcher: e.fetcher})
}

func (e *Extractor) Name() string { return "youtube" }

func (e *Extractor) CanProcess(url string) bool {
	return platforms.CanProcess(url, []string{
		"youtube.com",
		"youtu.be",
	})
}

func (e *Extractor) Metadata(ctx context.Context, url string) (interface{}, error) {
	contentType := DetectContentType(url)
	if contentType == ContentTypeChannel {
		doc, _, err := e.fetcher.Fetch(ctx, url)
		if err != nil {
			return nil, err
		}
		return ParseChannelMetadata(doc)
	}

	// Try each source in priority order
	var lastErr error
	for _, source := range e.sources {
		meta, err := source.FetchMetadata(ctx, url)
		if err != nil {
			lastErr = err
			continue
		}
		if meta != nil {
			return meta, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all metadata sources failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no metadata found")
}

func (e *Extractor) Content(ctx context.Context, url string) (interface{}, error) {
	doc, _, err := e.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	contentType := DetectContentType(url)

	switch contentType {
	case ContentTypeShort:
		return ParseShortContent(doc)
	default:
		return ParseVideoContent(doc)
	}
}

func (e *Extractor) Profile(ctx context.Context, url string) (interface{}, error) {
	if !strings.Contains(url, "/channel/") && !strings.Contains(url, "/@") && !strings.Contains(url, "/c/") {
		return nil, nil
	}

	doc, _, err := e.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	return ParseChannelMetadata(doc)
}

// oembedSource fetches metadata from YouTube's oEmbed API.
type oembedSource struct {
	fetcher *fetch.OEmbedFetcher
}

func (s *oembedSource) Name() string  { return "oembed" }
func (s *oembedSource) Priority() int { return 10 } // Highest priority

func (s *oembedSource) FetchMetadata(ctx context.Context, url string) (*VideoMetadata, error) {
	oembed, err := s.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	videoID := extractVideoID(url)
	meta := &VideoMetadata{
		Title:        oembed.Title,
		ChannelName:  oembed.AuthorName,
		ChannelURL:   oembed.AuthorURL,
		ThumbnailURL: oembed.ThumbnailURL,
		VideoID:      videoID,
	}

	return meta, nil
}

// htmlSource fetches metadata by parsing the HTML page.
type htmlSource struct {
	fetcher platforms.Fetcher
}

func (s *htmlSource) Name() string  { return "html" }
func (s *htmlSource) Priority() int { return 5 }

func (s *htmlSource) FetchMetadata(ctx context.Context, url string) (*VideoMetadata, error) {
	doc, _, err := s.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	return ParseVideoMetadata(doc, url)
}

func ParseShortContent(doc *goquery.Document) (*ShortContent, error) {
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(strings.TrimSuffix(title, "- YouTube"))

	description := extractDescription(doc)

	return &ShortContent{
		Title:       title,
		Description: description,
	}, nil
}

func init() {
	platforms.Register("youtube",
		func(url string) bool {
			return strings.Contains(url, "youtube.com") ||
				strings.Contains(url, "youtu.be")
		},
		func(client *http.Client) platforms.Extractor {
			return New(client)
		},
	)
}
