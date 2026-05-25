package github

import (
	"context"
	"net/http"
	"strings"

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

func NewWithFetcher(fetcher platforms.Fetcher) *Extractor {
	return &Extractor{
		fetcher: fetcher,
	}
}

func (e *Extractor) Name() string { return "github" }

func (e *Extractor) CanProcess(url string) bool {
	return platforms.CanProcess(url, []string{"github.com"})
}

func (e *Extractor) Metadata(ctx context.Context, url string) (interface{}, error) {
	doc, _, err := e.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	contentType := DetectContentType(url)

	switch contentType {
	case ContentTypeProfile:
		return ParseProfileMetadata(doc)
	default:
		return ParseRepoMetadata(doc, url)
	}
}

func (e *Extractor) Content(ctx context.Context, url string) (interface{}, error) {
	doc, _, err := e.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	contentType := DetectContentType(url)

	switch contentType {
	case ContentTypeRepo:
		return ParseReadmeContent(doc)
	default:
		return nil, nil
	}
}

func (e *Extractor) Profile(ctx context.Context, url string) (interface{}, error) {
	if !isProfileURL(url) {
		return nil, nil
	}

	doc, _, err := e.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	return ParseProfileMetadata(doc)
}

func init() {
	platforms.Register("github",
		func(url string) bool {
			return strings.Contains(url, "github.com")
		},
		func(client *http.Client) platforms.Extractor {
			return New(client)
		},
	)
}
