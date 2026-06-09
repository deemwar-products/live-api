package contracts

import "context"

// NotificationService - worker dispatches these
// From TRD Section 15: Notifications
type NotificationService interface {
	// Send dispatches a notification via specified channel
	Send(ctx context.Context, req NotificationRequest) error

	// SendBatch sends multiple notifications efficiently
	SendBatch(ctx context.Context, reqs []NotificationRequest) error
}

type NotificationRequest struct {
	OrgID     string
	Channel   string // "sms", "whatsapp", "email", "slack", "webhook", "inapp"
	Recipient string // phone, email, user_id, or webhook URL
	Template  string // template name
	Data      map[string]any
	Priority  string // "normal", "high", "urgent"
}

// From TRD Section 15.1: Notification Types
const (
	ChannelSMS      = "sms"
	ChannelWhatsApp = "whatsapp"
	ChannelEmail    = "email"
	ChannelSlack    = "slack"
	ChannelWebhook  = "webhook"
	ChannelInApp    = "inapp"
)

// Notification templates for common events
const (
	TemplateEscalationCreated = "escalation_created"
	TemplateTicketCreated     = "ticket_created"
	TemplateKnowledgeGapFound = "knowledge_gap"
	TemplateCreditAlert       = "credit_alert"
	TemplateAgentAvailable    = "agent_available"
	TemplateCallbackScheduled = "callback_scheduled"
)

// NotificationMetrics tracks notification delivery
type NotificationMetrics struct {
	TemplateID string
	Channel    string
	SentAt     int64
	Success    bool
	Error      string
	LatencyMs  int64
}
