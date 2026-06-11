package chunking

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/deemwar/live-api/apps/worker/internal/models"
	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
)

// EmbeddingClient produces vector embeddings for text. The chunking engine
// uses this during semantic splitting to find topic-shift boundaries.
type EmbeddingClient interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// ChunkingEngine splits documents into parent-child chunks using a
// semantic-then-structural strategy. It is safe for concurrent use after
// construction.
type ChunkingEngine struct {
	encoder *tiktoken.Tiktoken
	targetChildTokens int
	overlapTokens int
	// semanticThreshold is the cosine distance above which a new semantic
	// block is started between adjacent sentences. Range [0, 2].
	semanticThreshold float32
	// batchSize caps the number of sentences sent to EmbedBatch in one
	// call to avoid huge API requests.
	batchSize int
	// embedder is the configured embedding client. Must be set via
	// SetEmbedder before calling methods that need it.
	embedder EmbeddingClient
}

// Options configures a ChunkingEngine.
type Options struct {
	TargetChildTokens int
	OverlapTokens int
	SemanticThreshold float32
	BatchSize int
}

// New constructs a ChunkingEngine using the cl100k_base tokenizer
// (the BPE used by GPT-4 / GPT-3.5-turbo, the closest public match to
// Gemini's tokenizer).
func New(opts Options) (*ChunkingEngine, error) {
	if opts.TargetChildTokens <= 0 {
		opts.TargetChildTokens = 150
	}
	if opts.OverlapTokens < 0 {
		return nil, fmt.Errorf("OverlapTokens must be >= 0, got %d", opts.OverlapTokens)
	}
	if opts.OverlapTokens >= opts.TargetChildTokens {
		return nil, fmt.Errorf("OverlapTokens (%d) must be < TargetChildTokens (%d)", opts.OverlapTokens, opts.TargetChildTokens)
	}
	if opts.SemanticThreshold <= 0 {
		opts.SemanticThreshold = 0.5
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 64
	}
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("load tiktoken cl100k_base: %w", err)
	}
	return &ChunkingEngine{
		encoder: enc,
		targetChildTokens: opts.TargetChildTokens,
		overlapTokens: opts.OverlapTokens,
		semanticThreshold: opts.SemanticThreshold,
		batchSize: opts.BatchSize,
	}, nil
}

// sentenceSep matches a sentence terminator followed by whitespace or
// a newline. Go's regexp (RE2) doesn't support lookaheads, so this is
// the next-best: it matches the boundary char + the following space,
// which we then trim off in the splitter.
var sentenceSep = regexp.MustCompile(`[.!?](?:\s+|\n+)`)

// SplitSentences splits a document into sentences, trimming whitespace
// from each. Splits first on paragraph breaks, then on sentence
// terminators.
func SplitSentences(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	var out []string
	for _, para := range strings.Split(text, "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		// Split on sentence terminators. The regex consumes the
		// trailing whitespace, so we trim the result.
		locs := sentenceSep.FindAllStringIndex(para, -1)
		if len(locs) == 0 {
			out = append(out, para)
			continue
		}
		start := 0
		for _, m := range locs {
			// m[1] is just past the whitespace, so include it.
			end := m[1]
			if end > len(para) {
				end = len(para)
			}
			s := strings.TrimSpace(para[start:end])
			if s != "" {
				out = append(out, s)
			}
			start = end
		}
		if start < len(para) {
			s := strings.TrimSpace(para[start:])
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

// SplitDocumentSemantically partitions text into organic blocks by
// computing cosine distance between adjacent sentence embeddings and
// cutting where the distance exceeds the configured threshold.
func (e *ChunkingEngine) SplitDocumentSemantically(ctx context.Context, rawMarkdown string) ([]string, error) {
	sentences := SplitSentences(rawMarkdown)
	if len(sentences) == 0 {
		return nil, nil
	}
	if len(sentences) == 1 {
		return []string{sentences[0]}, nil
	}

	embeddings, err := e.batchedEmbed(ctx, sentences)
	if err != nil {
		return nil, fmt.Errorf("embed sentences: %w", err)
	}

	var blocks []string
	var current []string
	for i, s := range sentences {
		if i == 0 {
			current = append(current, s)
			continue
		}
		dist := CosineDistance(embeddings[i-1], embeddings[i])
		if dist > e.semanticThreshold && len(current) > 0 {
			blocks = append(blocks, strings.Join(current, " "))
			current = current[:0]
		}
		current = append(current, s)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, " "))
	}
	return blocks, nil
}

// batchedEmbed calls EmbedBatch in chunks of e.batchSize. Empty input
// returns nil, nil.
func (e *ChunkingEngine) batchedEmbed(ctx context.Context, sentences []string) ([][]float32, error) {
	if len(sentences) == 0 {
		return nil, nil
	}
	if e.embedder == nil {
		return nil, fmt.Errorf("embedding client not configured (call SetEmbedder)")
	}
	var result [][]float32
	for start := 0; start < len(sentences); start += e.batchSize {
		end := start + e.batchSize
		if end > len(sentences) {
			end = len(sentences)
		}
		batch, err := e.embedder.EmbedBatch(ctx, sentences[start:end])
		if err != nil {
			return nil, err
		}
		result = append(result, batch...)
	}
	return result, nil
}

// SetEmbedder wires the engine to a real EmbeddingClient. Must be
// called before SplitDocumentSemantically or IngestDocument.
func (e *ChunkingEngine) SetEmbedder(c EmbeddingClient) {
	e.embedder = c
}

// IngestDocument runs the full chunking pipeline: semantic split, then
// parent-child decomposition with identity mapping for short blocks and
// sliding window for large blocks. It does NOT generate embeddings for
// the children — the caller is responsible for embedding them and
// assigning each child an Embedding. (This lets the caller batch
// embeddings across all parents in one API call.)
func (e *ChunkingEngine) IngestDocument(ctx context.Context, docID, rawMarkdown string) ([]models.ParentChunk, []models.ChildChunk, error) {
	if docID == "" {
		return nil, nil, fmt.Errorf("docID is required")
	}
	blocks, err := e.SplitDocumentSemantically(ctx, rawMarkdown)
	if err != nil {
		return nil, nil, fmt.Errorf("semantic split: %w", err)
	}
	var parents []models.ParentChunk
	var children []models.ChildChunk
	for i, block := range blocks {
		parent := models.ParentChunk{
			ID: uuid.NewString(),
			DocumentID: docID,
			Content: block,
			TokenCount: e.TokenCount(block),
			Position: i,
		}
		parents = append(parents, parent)

		// Identity mapping: if the block already fits, parent and child
		// are the same text. This prevents fragmentation of small
		// semantic blocks into multiple overlapping children.
		if parent.TokenCount <= e.targetChildTokens {
			children = append(children, models.ChildChunk{
				ID: uuid.NewString(),
				ParentID: parent.ID,
				Content: block,
				TokenCount: parent.TokenCount,
				Position: 0,
				Embedding: nil, // caller must set this
			})
			continue
		}

		// Sliding window for larger blocks.
		chunks, err := e.slidingWindow(block)
		if err != nil {
			return nil, nil, fmt.Errorf("sliding window block %d: %w", i, err)
		}
		for j, c := range chunks {
			children = append(children, models.ChildChunk{
				ID: uuid.NewString(),
				ParentID: parent.ID,
				Content: c,
				TokenCount: e.TokenCount(c),
				Position: j,
				Embedding: nil,
			})
		}
	}
	return parents, children, nil
}

// slidingWindow tokenizes text and produces overlapping windows of
// approximately targetChildTokens tokens with overlapTokens of overlap.
func (e *ChunkingEngine) slidingWindow(text string) ([]string, error) {
	tokens := e.encoder.Encode(text, nil, nil)
	n := len(tokens)
	if n == 0 {
		return nil, nil
	}
	if n <= e.targetChildTokens {
		return []string{text}, nil
	}
	step := e.targetChildTokens - e.overlapTokens
	if step <= 0 {
		return nil, fmt.Errorf("sliding window step must be > 0 (target=%d, overlap=%d)", e.targetChildTokens, e.overlapTokens)
	}
	var chunks []string
	for start := 0; start < n; start += step {
		end := start + e.targetChildTokens
		if end > n {
			end = n
		}
		if end <= start {
			break
		}
		// Decode the slice back to a string. The encoder tolerates
		// arbitrary token ranges.
		chunk := e.encoder.Decode(tokens[start:end])
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
		if end == n {
			break
		}
	}
	return chunks, nil
}

// TokenCount returns the number of cl100k_base tokens in text.
func (e *ChunkingEngine) TokenCount(text string) int {
	return len(e.encoder.Encode(text, nil, nil))
}
