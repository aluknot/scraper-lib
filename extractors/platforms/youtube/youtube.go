package youtube

import (
	"context"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aluknot/scraper-lib/extractors/platforms"
)

type Extractor struct {
	client  *http.Client
	fetcher platforms.Fetcher
}

func New(client *http.Client) *Extractor {
	if client == nil {
		client = &http.Client{}
	}
	return &Extractor{
		client:  client,
		fetcher: platforms.NewGoQueryFetcher(client),
	}
}

func (e *Extractor) Name() string { return "youtube" }

func (e *Extractor) CanProcess(url string) bool {
	return platforms.CanProcess(url, []string{
		"youtube.com",
		"youtu.be",
	})
}

func (e *Extractor) Metadata(ctx context.Context, url string) (interface{}, error) {
	doc, _, err := e.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	contentType := DetectContentType(url)

	switch contentType {
	case ContentTypeChannel:
		return ParseChannelMetadata(doc)
	default:
		return ParseVideoMetadata(doc, url)
	}
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
