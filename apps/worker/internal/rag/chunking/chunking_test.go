package chunking

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeEmbedder is a deterministic EmbeddingClient for tests.
// It assigns orthogonal-ish unit vectors based on a hash of the text.
type fakeEmbedder struct {
	calls atomic.Int32
	dim int
}

func newFakeEmbedder() *fakeEmbedder { return &fakeEmbedder{dim: 4} }

func (f *fakeEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	out, err := f.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func (f *fakeEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	f.calls.Add(1)
	out := make([][]float32, len(texts))
	for i, t := range texts {
		// Hash by rune sum so identical sentences get identical vectors.
		var sum int
		for _, r := range t {
			sum += int(r)
		}
		v := make([]float32, f.dim)
		v[sum%f.dim] = 1
		out[i] = v
	}
	return out, nil
}

func TestNew_DefaultsAndErrors(t *testing.T) {
	if _, err := New(Options{}); err != nil {
		t.Errorf("default opts should succeed: %v", err)
	}
	if _, err := New(Options{TargetChildTokens: 10, OverlapTokens: 10}); err == nil {
		t.Error("overlap >= target should error")
	}
	if _, err := New(Options{TargetChildTokens: 10, OverlapTokens: -1}); err == nil {
		t.Error("negative overlap should error")
	}
}

func TestSplitSentences(t *testing.T) {
	cases := []struct {
		name, in string
		want int
	}{
		{"empty", "", 0},
		{"single", "Hello world.", 1},
		{"two", "First sentence. Second sentence.", 2},
		{"paragraphs", "Para one.\n\nPara two.", 2},
		{"multipunct", "Wow! Is this working? Yes it is.", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitSentences(tc.in)
			if len(got) != tc.want {
				t.Errorf("got %d sentences (%q), want %d", len(got), got, tc.want)
			}
		})
	}
}

func TestSlidingWindow_FitsAsOne(t *testing.T) {
	e, _ := New(Options{TargetChildTokens: 100, OverlapTokens: 10})
	chunks, err := e.slidingWindow("This is a short block of text.")
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestSlidingWindow_ProducesOverlappingChunks(t *testing.T) {
	// Make a long string that requires multiple windows.
	word := "alpha "
	text := strings.Repeat(word, 200) // ~1000 chars, ~250 tokens
	e, _ := New(Options{TargetChildTokens: 50, OverlapTokens: 10})
	chunks, err := e.slidingWindow(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks, got %d", len(chunks))
	}
	// Adjacent chunks should overlap (share some words).
	overlap := sharedWords(chunks[0], chunks[1])
	if overlap < 1 {
		t.Errorf("expected overlap between adjacent chunks, got %d", overlap)
	}
}

func sharedWords(a, b string) int {
	aw := strings.Fields(a)
	bw := strings.Fields(b)
	set := make(map[string]struct{}, len(aw))
	for _, w := range aw {
		set[w] = struct{}{}
	}
	var n int
	for _, w := range bw {
		if _, ok := set[w]; ok {
			n++
		}
	}
	return n
}

func TestSlidingWindow_EmptyInput(t *testing.T) {
	e, _ := New(Options{TargetChildTokens: 100, OverlapTokens: 10})
	chunks, err := e.slidingWindow("")
	if err != nil {
		t.Fatal(err)
	}
	if chunks != nil {
		t.Errorf("expected nil chunks for empty input, got %v", chunks)
	}
}

func TestTokenCount(t *testing.T) {
	e, _ := New(Options{})
	n := e.TokenCount("Hello, world!")
	if n == 0 {
		t.Error("expected non-zero token count")
	}
}

func TestSplitDocumentSemantically_Empty(t *testing.T) {
	e, _ := New(Options{})
	e.SetEmbedder(newFakeEmbedder())
	got, err := e.SplitDocumentSemantically(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSplitDocumentSemantically_SingleSentence(t *testing.T) {
	e, _ := New(Options{})
	e.SetEmbedder(newFakeEmbedder())
	got, err := e.SplitDocumentSemantically(context.Background(), "Just one sentence here.")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 block, got %d", len(got))
	}
}

func TestSplitDocumentSemantically_Threshold(t *testing.T) {
	// Two sentences with completely different content -> hash gives
	// different unit vectors -> cosine distance is high -> split.
	e, _ := New(Options{SemanticThreshold: 0.5})
	e.SetEmbedder(newFakeEmbedder())
	// Use sentences whose first rune is different to land on different
	// hash buckets, producing orthogonal vectors.
	in := "Apple is red. Zebra has stripes."
	got, err := e.SplitDocumentSemantically(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 {
		t.Errorf("expected split, got %d blocks", len(got))
	}
}

func TestSplitDocumentSemantically_NoSplitWhenSimilar(t *testing.T) {
	// Identical sentences -> identical vectors -> distance 0 -> no split.
	e, _ := New(Options{SemanticThreshold: 0.1})
	e.SetEmbedder(newFakeEmbedder())
	in := "Hello there. Hello there. Hello there."
	got, err := e.SplitDocumentSemantically(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 block for similar sentences, got %d", len(got))
	}
}

func TestSplitDocumentSemantically_EmbedderError(t *testing.T) {
	e, _ := New(Options{})
	e.SetEmbedder(errorEmbedder{})
	_, err := e.SplitDocumentSemantically(context.Background(), "First. Second.")
	if err == nil {
		t.Error("expected error from embedder")
	}
}

type errorEmbedder struct{}

func (errorEmbedder) Embed(ctx context.Context, t string) ([]float32, error) {
	return nil, errors.New("nope")
}
func (errorEmbedder) EmbedBatch(ctx context.Context, ts []string) ([][]float32, error) {
	return nil, errors.New("nope")
}

func TestIngestDocument_RequiresDocID(t *testing.T) {
	e, _ := New(Options{})
	_, _, err := e.IngestDocument(context.Background(), "", "anything")
	if err == nil {
		t.Error("expected error for empty docID")
	}
}

func TestIngestDocument_EmptyMarkdown(t *testing.T) {
	e, _ := New(Options{})
	e.SetEmbedder(newFakeEmbedder())
	parents, children, err := e.IngestDocument(context.Background(), "doc1", "")
	if err != nil {
		t.Fatal(err)
	}
	if parents != nil {
		t.Errorf("expected nil parents, got %v", parents)
	}
	if children != nil {
		t.Errorf("expected nil children, got %v", children)
	}
}

func TestIngestDocument_IdentityMapping(t *testing.T) {
	// Single short sentence: token count <= target => identity mapping.
	e, _ := New(Options{TargetChildTokens: 100, OverlapTokens: 10})
	e.SetEmbedder(newFakeEmbedder())
	parents, children, err := e.IngestDocument(context.Background(), "doc1", "Just one short sentence.")
	if err != nil {
		t.Fatal(err)
	}
	if len(parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(parents))
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child (identity), got %d", len(children))
	}
	if parents[0].Content != children[0].Content {
		t.Error("identity mapping: parent and child content should match")
	}
	if parents[0].ID != children[0].ParentID {
		t.Error("child.ParentID should equal parent.ID")
	}
}

func TestIngestDocument_ProducesParentsAndChildren(t *testing.T) {
	// Build a long-ish text that should produce multiple semantic blocks
	// each large enough to need sliding window.
	var b strings.Builder
	for i := 0; i < 20; i++ {
		b.WriteString("The quick brown fox jumps over the lazy dog. ")
	}
	text := b.String()
	e, _ := New(Options{TargetChildTokens: 20, OverlapTokens: 5})
	e.SetEmbedder(newFakeEmbedder())
	parents, children, err := e.IngestDocument(context.Background(), "doc1", text)
	if err != nil {
		t.Fatal(err)
	}
	if len(parents) == 0 {
		t.Fatal("expected at least 1 parent")
	}
	if len(children) == 0 {
		t.Fatal("expected at least 1 child")
	}
	// Every child should point to a real parent.
	for _, c := range children {
		found := false
		for _, p := range parents {
			if p.ID == c.ParentID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("child %s has no matching parent", c.ID)
		}
	}
	// Children should have nil Embedding (caller fills it).
	for _, c := range children {
		if c.Embedding != nil {
			t.Error("expected nil embedding in chunking output")
		}
	}
}

func TestSplitDocumentSemantically_NoEmbedder(t *testing.T) {
	e, _ := New(Options{})
	_, err := e.SplitDocumentSemantically(context.Background(), "First. Second.")
	if err == nil {
		t.Error("expected error when no embedder configured")
	}
}

func TestIngestDocument_NoEmbedder(t *testing.T) {
	e, _ := New(Options{})
	_, _, err := e.IngestDocument(context.Background(), "doc1", "First. Second.")
	if err == nil {
		t.Error("expected error when no embedder configured")
	}
}

func TestBatchedEmbed_Empty(t *testing.T) {
	e, _ := New(Options{})
	e.SetEmbedder(newFakeEmbedder())
	got, err := e.batchedEmbed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSplitSentences_EmptyParagraphs(t *testing.T) {
	// Whitespace-only paragraph between content paragraphs.
	got := SplitSentences("\n\n\n\nHello.\n\n   \n\nWorld.\n\n")
	if len(got) != 2 {
		t.Errorf("expected 2 sentences, got %d: %q", len(got), got)
	}
}
