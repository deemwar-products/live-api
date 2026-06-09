package contracts

import "context"

// SessionStore - implemented by session/core team
// From TRD Section 7: Session Management
type SessionStore interface {
	// Get retrieves session state from Redis
	Get(ctx context.Context, sessionID string) (*Session, error)

	// GetOrgConfig retrieves org settings
	GetOrgConfig(ctx context.Context, orgID string) (*OrgConfig, error)

	// UpdateToolState updates MCP tool state during session
	UpdateToolState(ctx context.Context, sessionID string, toolState map[string]any) error

	// LogToolExecution logs a tool call for audit
	LogToolExecution(ctx context.Context, sessionID string, execution ToolExecution) error
}

type Session struct {
	ID         string
	OrgID      string
	CustomerID string
	Mode       string // "voice" or "chat"
	Status     string // "active", "escalated", "completed", "failed"
	StartedAt  int64  // unix timestamp

	// From TRD 7.2: Session state in Redis
	ConversationHistory []ConversationTurn
	ContextChunks       []string // RAG chunk IDs retrieved
	MCPToolState        map[string]any
	PreservedEntities   map[string]string // order IDs, account numbers, etc.
}

type ConversationTurn struct {
	Role      string // "customer", "ai", "agent"
	Content   string
	Timestamp int64
	TurnIndex int
}

type OrgConfig struct {
	OrgID                 string
	Name                  string
	EscalationMode        string // "live" or "async" (from TRD Section 3.4 - Escalation Strategy)
	EscalationEnabled     bool
	EscalationThreshold   float64 // conversation scoring threshold (default 0.3, red threshold)
	MCPServers            []MCPServerConfig
	AudioRecordingEnabled bool
	SystemPrompt          string // from TRD 4.3
}

type MCPServerConfig struct {
	ID               string
	OrgID            string
	URL              string
	AuthType         string // "api_key", "oauth2", "mtls"
	EncryptedCreds   []byte // AES-256-GCM encrypted (from TRD 6.4)
	ConnectionStatus string // "healthy", "degraded", "down"
}

type ToolExecution struct {
	ToolName  string
	Arguments map[string]any
	Success   bool
	Error     string
	LatencyMs int64
	Result    map[string]any
}
