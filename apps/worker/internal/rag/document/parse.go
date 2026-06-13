// Package document extracts plain text from various document formats.
package document

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

// Parse extracts plain text from document bytes.
// Supported sourceType values: "markdown", "text", "html", "pdf", "docx".
func Parse(data []byte, sourceType string) (string, error) {
	switch sourceType {
	case "markdown", "text", "html":
		return string(data), nil
	case "pdf":
		return parsePDF(data)
	case "docx":
		return parseDOCX(data)
	default:
		return "", fmt.Errorf("unsupported source type: %q", sourceType)
	}
}

func parsePDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	// GetPlainText only errors on corrupt per-page font maps; skip that error
	// since partial extraction is better than rejecting the whole document.
	plain, _ := r.GetPlainText()
	var buf bytes.Buffer
	buf.ReadFrom(plain) // bytes.Buffer.ReadFrom never errors
	return buf.String(), nil
}

var xmlTagRe = regexp.MustCompile(`<[^>]+>`)

func parseDOCX(data []byte) (string, error) {
	r, err := docx.ReadDocxFromMemory(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}
	defer r.Close()
	raw := r.Editable().GetContent()
	text := xmlTagRe.ReplaceAllString(raw, " ")
	return strings.Join(strings.Fields(text), " "), nil
}
