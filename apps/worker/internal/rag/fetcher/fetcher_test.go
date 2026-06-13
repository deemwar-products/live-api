package fetcher

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/test.pdf
var testPDF []byte

//go:embed testdata/test.docx
var testDOCX []byte

func TestNewLocalFetcher(t *testing.T) {
	f := NewLocalFetcher("/some/dir")
	if f == nil {
		t.Fatal("expected non-nil fetcher")
	}
	if f.baseDir != "/some/dir" {
		t.Errorf("baseDir: got %q want %q", f.baseDir, "/some/dir")
	}
}

func TestFetch_Markdown(t *testing.T) {
	dir := t.TempDir()
	content := "# Hello World"
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	f := NewLocalFetcher(dir)
	got, err := f.Fetch(context.Background(), "doc.md", "markdown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q want %q", got, content)
	}
}

func TestFetch_Text(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	f := NewLocalFetcher(dir)
	got, err := f.Fetch(context.Background(), "doc.txt", "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestFetch_PDF(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.pdf"), testPDF, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	f := NewLocalFetcher(dir)
	got, err := f.Fetch(context.Background(), "doc.pdf", "pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty PDF content")
	}
}

func TestFetch_DOCX(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "doc.docx"), testDOCX, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	f := NewLocalFetcher(dir)
	got, err := f.Fetch(context.Background(), "doc.docx", "docx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty DOCX content")
	}
}

func TestFetch_FileNotFound(t *testing.T) {
	f := NewLocalFetcher(t.TempDir())
	_, err := f.Fetch(context.Background(), "missing.txt", "text")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "missing.txt") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetch_ReadError(t *testing.T) {
	f := NewLocalFetcher("/any")
	f.readFile = func(string) ([]byte, error) {
		return nil, errors.New("disk error")
	}
	_, err := f.Fetch(context.Background(), "doc.txt", "text")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "disk error") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetch_ParseError(t *testing.T) {
	f := NewLocalFetcher("/any")
	f.readFile = func(string) ([]byte, error) {
		return []byte("data"), nil
	}
	_, err := f.Fetch(context.Background(), "doc.xyz", "unsupported_type")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported source type") {
		t.Errorf("unexpected error: %v", err)
	}
}
