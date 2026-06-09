package contracts

import "context"

// PresenceService - implemented by agent management team
// From TRD Section 7.3: Escalation Flow
type PresenceService interface {
	// GetAvailableAgents returns agents who can take escalations
	GetAvailableAgents(ctx context.Context, orgID string) ([]Agent, error)

	// NotifyAgents notifies available agents of an escalation
	// Returns notified agent count
	NotifyAgents(ctx context.Context, orgID string, escalation EscalationRequest) (int, error)

	// UpdateAgentStatus updates agent status (online, away, busy, etc.)
	UpdateAgentStatus(ctx context.Context, agentID string, status string) error

	// GetAgentCapacity returns current workload capacity (0-100%)
	GetAgentCapacity(ctx context.Context, agentID string) (int, error)
}

// Agent represents a human support agent
// From TRD Section 3.1: Agent role hierarchy
type Agent struct {
	ID       string
	Name     string
	Status   string // "available", "busy", "away", "offline"
	Capacity int    // 0-100, current workload percentage
	Channel  string // "voice", "chat", "both"
}

type EscalationRequest struct {
	SessionID       string
	OrgID           string
	CustomerID      string
	Summary         string
	Priority        string // "low", "normal", "high", "urgent"
	Transcript      []ConversationTurn
	ClassifierScore float64 // From TRD Section 17: Conversation Scoring
	ScoringReason   string  // Why it's being escalated
}

// EscalationReason explains why escalation is needed
// From TRD Section 7.3
const (
	EscalationReasonCustomerRequest = "customer_request"
	EscalationReasonScoreThreshold  = "score_threshold"
	EscalationReasonToolFailure     = "tool_failure"
	EscalationReasonKnowledgeGap    = "knowledge_gap"
)
