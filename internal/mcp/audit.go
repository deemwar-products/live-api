package mcp

import (
	"context"
	"sync"
	"time"
)

type ToolExecution struct {
	ToolName   string        `json:"tool_name"`
	OrgID      string        `json:"org_id"`
	UserID     string        `json:"user_id,omitempty"`
	SessionID  string        `json:"session_id,omitempty"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	ExecutedAt time.Time     `json:"executed_at"`
}

type AuditSink interface {
	Record(ctx context.Context, execution ToolExecution) error
}

type NoopAuditSink struct{}

func (NoopAuditSink) Record(context.Context, ToolExecution) error { return nil }

type InMemoryAuditSink struct {
	mu         sync.RWMutex
	executions []ToolExecution
	maxEntries int
}

func NewInMemoryAuditSink() *InMemoryAuditSink {
	return &InMemoryAuditSink{}
}

func NewBoundedAuditSink(maxEntries int) *InMemoryAuditSink {
	return &InMemoryAuditSink{maxEntries: maxEntries}
}

func (s *InMemoryAuditSink) Record(_ context.Context, execution ToolExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions = append(s.executions, execution)
	if s.maxEntries > 0 && len(s.executions) > s.maxEntries {
		overflow := len(s.executions) - s.maxEntries
		copy(s.executions, s.executions[overflow:])
		s.executions = s.executions[:s.maxEntries]
	}
	return nil
}

func (s *InMemoryAuditSink) List() []ToolExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ToolExecution, len(s.executions))
	copy(out, s.executions)
	return out
}
