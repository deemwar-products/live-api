package embedding

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// GeminiClient implements chunking.EmbeddingClient using the Gemini API.
type GeminiClient struct {
	client *genai.Client
	model  string
}

// NewGeminiClient creates a Gemini embedding client.
func NewGeminiClient(ctx context.Context, apiKey, model string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiClient{client: client, model: model}, nil
}

// Embed returns a single embedding vector for the given text.
func (g *GeminiClient) Embed(ctx context.Context, text string) ([]float32, error) {
	result, err := g.client.Models.EmbedContent(ctx, g.model, genai.Text(text), nil)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("embed: empty response")
	}
	return result.Embeddings[0].Values, nil
}

// EmbedBatch returns embedding vectors for each text, in order.
// Calls Embed sequentially — sufficient for PoC workloads.
func (g *GeminiClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		emb, err := g.Embed(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("embed[%d]: %w", i, err)
		}
		out[i] = emb
	}
	return out, nil
}
