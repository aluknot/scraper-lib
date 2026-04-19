//go:build integration

package scraperlib

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/aluknot/scraper-lib/extractors"
)

func TestDebug_MinWordsZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Configurar logging para ver todo a stderr
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	testURL := "https://ansitype.com/"
	t.Logf("Testing URL: %s", testURL)
	t.Logf("Test 1: MinWords=100 (default - should fail)")

	result, err := Extract(context.Background(), testURL, &Options{
		Timeout:     30 * time.Second,
		Debug:       true,
		UseAdvanced: true,
		MinWords:    100, // default
	})
	_ = result // unused in test

	t.Logf("Test 1 result: err=%v", err)
	if err != nil {
		t.Logf("Test 1 failed as expected")
	}

	t.Logf("\n--- Test 2: MinWords=0 (should work) ---")
	result2, err2 := Extract(context.Background(), testURL, &Options{
		Timeout:     30 * time.Second,
		Debug:       true,
		UseAdvanced: true,
		MinWords:    0, // Accept any result
	})

	if err2 != nil {
		t.Logf("Test 2 ERROR: %v", err2)
		t.Fail()
		return
	}

	t.Logf("Test 2 SUCCESS!")
	t.Logf("ExtractorUsed: %s", result2.ExtractorUsed)
	t.Logf("WordCount: %d", result2.WordCount)
	t.Logf("Title: %s", result2.Article.Title)

	if result2.Article != nil && result2.Article.Content != "" {
		t.Logf("Content preview: %s", result2.Article.Content[:200])
	}
}

// TestDebug_MetadataExtractor tests the new MetadataExtractor
func TestDebug_MetadataExtractor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	testURL := "https://ansitype.com/"

	t.Logf("--- Test: MetadataExtractor (direct) ---")
	metadata, err := extractors.NewMetadataExtractor().Extract(context.Background(), "test", testURL)
	if err != nil {
		t.Logf("MetadataExtractor ERROR: %v", err)
		t.Fail()
		return
	}

	t.Logf("MetadataExtractor SUCCESS!")
	t.Logf("Title: %s", metadata.Title)
	t.Logf("WordCount: %d", metadata.WordCount)
	t.Logf("Author: %s", metadata.Author)
}
