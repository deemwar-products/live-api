package memory

import (
	"context"
	"fmt"
	"sync"
)

type Message struct {
	Role    string
	Content string
}

type ShortTermStore interface {
	Append(ctx context.Context, sessionID string, message Message) error
	List(ctx context.Context, sessionID string, limit int) ([]Message, error)
}

type InMemoryStore struct {
	mu       sync.RWMutex
	messages map[string][]Message
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{messages: make(map[string][]Message)}
}

func (s *InMemoryStore) Append(ctx context.Context, sessionID string, message Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if message.Role == "" || message.Content == "" {
		return fmt.Errorf("message role and content are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[sessionID] = append(s.messages[sessionID], message)
	return nil
}

func (s *InMemoryStore) List(ctx context.Context, sessionID string, limit int) ([]Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	messages := s.messages[sessionID]
	start := 0
	if limit > 0 && len(messages) > limit {
		start = len(messages) - limit
	}
	out := make([]Message, len(messages)-start)
	copy(out, messages[start:])
	return out, nil
}
