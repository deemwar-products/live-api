package contracts

import "context"

// RAGService - implemented by RAG team
// From TRD Section 5: RAG Pipeline
type RAGService interface {
	// RetrieveKnowledge performs semantic search
	// Called by MCP retrieve_knowledge tool
	// From TRD 5.3: Top-20 retrieval, relevance threshold 0.5
	RetrieveKnowledge(ctx context.Context, req KnowledgeRequest) (*KnowledgeResponse, error)
}

type KnowledgeRequest struct {
	OrgID     string
	Query     string
	TopK      int     // default: 20 (from TRD 5.3)
	Threshold float64 // default: 0.5 (from TRD 5.4)
}

type KnowledgeResponse struct {
	Chunks     []KnowledgeChunk
	HasGap     bool // true if all chunks < threshold
	TotalCount int
}

type KnowledgeChunk struct {
	ID         string
	Content    string
	Score      float64 // relevance score from vector search
	DocumentID string
	Source     string
	ChunkIndex int
}

// From TRD 5.4: Relevance Thresholds
// > 0.75: Include — confident answer
// 0.50–0.75: Include — AI should express some uncertainty
// < 0.50: Exclude — not relevant enough
// All chunks < 0.50: Flag as knowledge gap, escalate or return fallback
const (
	HighConfidenceThreshold   float64 = 0.75
	MediumConfidenceThreshold float64 = 0.50
	LowConfidenceThreshold    float64 = 0.30
)
