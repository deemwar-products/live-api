// Package fetcher retrieves document content from a source and returns plain text.
package fetcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deemwar/live-api/apps/worker/internal/rag/document"
)

// Fetcher retrieves and returns the plain text content of a document.
type Fetcher interface {
	Fetch(ctx context.Context, sourceName, sourceType string) (string, error)
}

// LocalFetcher reads documents from the local filesystem under a base directory.
type LocalFetcher struct {
	baseDir  string
	readFile func(string) ([]byte, error) // injectable for testing
}

// NewLocalFetcher returns a Fetcher that reads files from baseDir.
func NewLocalFetcher(baseDir string) *LocalFetcher {
	return &LocalFetcher{
		baseDir:  baseDir,
		readFile: os.ReadFile,
	}
}

// Fetch reads sourceName from the base directory and parses it into plain text.
func (f *LocalFetcher) Fetch(_ context.Context, sourceName, sourceType string) (string, error) {
	path := filepath.Join(f.baseDir, sourceName)
	data, err := f.readFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", sourceName, err)
	}
	text, err := document.Parse(data, sourceType)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", sourceName, err)
	}
	return text, nil
}
