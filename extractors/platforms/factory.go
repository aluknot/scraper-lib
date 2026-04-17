package platforms

import (
	"context"
	"fmt"
	"net/http"
)

type Extractor interface {
	Name() string
	CanProcess(url string) bool
	Metadata(ctx context.Context, url string) (interface{}, error)
	Content(ctx context.Context, url string) (interface{}, error)
	Profile(ctx context.Context, url string) (interface{}, error)
}

type extractorEntry struct {
	name       string
	canProcess func(url string) bool
	new        func(client *http.Client) Extractor
}

var registeredExtractors []extractorEntry

func Register(name string, canProcess func(url string) bool, new func(client *http.Client) Extractor) {
	registeredExtractors = append(registeredExtractors, extractorEntry{
		name:       name,
		canProcess: canProcess,
		new:        new,
	})
}

func Get(client *http.Client, url string) (Extractor, error) {
	if client == nil {
		client = &http.Client{}
	}

	for _, entry := range registeredExtractors {
		if entry.canProcess(url) {
			return entry.new(client), nil
		}
	}

	return nil, fmt.Errorf("no extractor found for URL: %s", url)
}

func GetByName(client *http.Client, name string) (Extractor, error) {
	if client == nil {
		client = &http.Client{}
	}

	for _, entry := range registeredExtractors {
		if entry.name == name {
			return entry.new(client), nil
		}
	}

	return nil, fmt.Errorf("extractor not found: %s", name)
}

func ListExtractors() []string {
	names := make([]string, len(registeredExtractors))
	for i, entry := range registeredExtractors {
		names[i] = entry.name
	}
	return names
}
