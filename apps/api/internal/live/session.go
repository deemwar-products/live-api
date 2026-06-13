package live

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

// LiveSession holds the state for one active client ↔ Gemini session.
type LiveSession struct {
	ID         string
	ClientConn *websocket.Conn
	GeminiSess *genai.Session
	StartedAt  time.Time
}

// SessionManager holds all active sessions in memory.
// It is safe for concurrent use.
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*LiveSession
	maxSessions int
}

// NewSessionManager creates a manager capped at maxSessions concurrent sessions.
func NewSessionManager(maxSessions int) *SessionManager {
	return &SessionManager{
		sessions:    make(map[string]*LiveSession),
		maxSessions: maxSessions,
	}
}

// Add stores a session. Returns an error if the cap is reached.
func (m *SessionManager) Add(s *LiveSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sessions) >= m.maxSessions {
		return fmt.Errorf("max concurrent sessions (%d) reached", m.maxSessions)
	}
	m.sessions[s.ID] = s
	return nil
}

// Get retrieves a session by ID.
func (m *SessionManager) Get(id string) (*LiveSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

// Remove deletes the session and closes both connections.
func (m *SessionManager) Remove(id string) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	if ok {
		_ = s.GeminiSess.Close()
		_ = s.ClientConn.Close()
	}
}

// Count returns the number of active sessions.
func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
