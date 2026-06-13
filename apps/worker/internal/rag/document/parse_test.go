package document

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed testdata/test.pdf
var testPDF []byte

//go:embed testdata/test.docx
var testDOCX []byte

// --- Parse (dispatch) ---

func TestParse_Markdown(t *testing.T) {
	got, err := Parse([]byte("# Hello"), "markdown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "# Hello" {
		t.Errorf("got %q want %q", got, "# Hello")
	}
}

func TestParse_Text(t *testing.T) {
	got, err := Parse([]byte("plain text"), "text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "plain text" {
		t.Errorf("got %q", got)
	}
}

func TestParse_HTML(t *testing.T) {
	got, err := Parse([]byte("<p>hello</p>"), "html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "<p>hello</p>" {
		t.Errorf("got %q", got)
	}
}

func TestParse_UnknownType(t *testing.T) {
	_, err := Parse([]byte("data"), "unknown")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unsupported source type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParse_EmptyType(t *testing.T) {
	_, err := Parse([]byte("data"), "")
	if err == nil {
		t.Fatal("expected error for empty type")
	}
}

// --- PDF ---

func TestParse_PDF_Valid(t *testing.T) {
	got, err := Parse(testPDF, "pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty text from PDF")
	}
}

func TestParse_PDF_InvalidBytes(t *testing.T) {
	_, err := Parse([]byte("not a pdf"), "pdf")
	if err == nil {
		t.Fatal("expected error for invalid PDF")
	}
}

func TestParse_PDF_Empty(t *testing.T) {
	_, err := Parse([]byte{}, "pdf")
	if err == nil {
		t.Fatal("expected error for empty PDF bytes")
	}
}

// --- DOCX ---

func TestParse_DOCX_Valid(t *testing.T) {
	got, err := Parse(testDOCX, "docx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DOCX content should be non-empty after XML stripping
	if len(got) == 0 {
		t.Error("expected non-empty text from DOCX")
	}
}

func TestParse_DOCX_InvalidBytes(t *testing.T) {
	_, err := Parse([]byte("not a docx"), "docx")
	if err == nil {
		t.Fatal("expected error for invalid DOCX")
	}
}

func TestParse_DOCX_Empty(t *testing.T) {
	_, err := Parse([]byte{}, "docx")
	if err == nil {
		t.Fatal("expected error for empty DOCX bytes")
	}
}
