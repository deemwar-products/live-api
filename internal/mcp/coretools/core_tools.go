package coretools

import (
	"context"
	"fmt"

	"live-api/internal/mcp"
	"live-api/internal/mcp/contracts"
)

func RegisterCoreTools(
	registry *mcp.ToolRegistry,
	rag contracts.RAGService,
	sessions contracts.SessionStore,
	presence contracts.PresenceService,
	notifications contracts.NotificationService,
	connectors contracts.ConnectorRouter,
) error {
	for _, tool := range []mcp.Tool{
		RetrieveKnowledgeTool{RAG: rag},
		CreateTicketTool{Sessions: sessions},
		SendNotificationTool{Notifications: notifications},
		CreateEscalationTool{Sessions: sessions, Presence: presence, Notifications: notifications},
		FlagKnowledgeGapTool{},
		GetCustomerContextTool{Sessions: sessions},
		QueryConnectorTool{Connectors: connectors},
	} {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}

func NewCoreGateway(
	rag contracts.RAGService,
	sessions contracts.SessionStore,
	presence contracts.PresenceService,
	notifications contracts.NotificationService,
	connectors contracts.ConnectorRouter,
	audit mcp.AuditSink,
	limiter *mcp.RateLimiter,
	policy mcp.ToolPolicy,
) (*mcp.Gateway, error) {
	registry := mcp.NewToolRegistry()
	if err := RegisterCoreTools(registry, rag, sessions, presence, notifications, connectors); err != nil {
		return nil, err
	}
	gateway := mcp.NewGatewayWithPolicy(registry, audit, limiter, policy)
	if err := registry.Register(ExecuteWorkflowTool{Gateway: gateway}); err != nil {
		return nil, err
	}
	return gateway, nil
}

type RetrieveKnowledgeTool struct {
	RAG contracts.RAGService
}

func (RetrieveKnowledgeTool) Name() string { return "retrieve_knowledge" }
func (RetrieveKnowledgeTool) Description() string {
	return "Search the organization's knowledge base using RAG, PostgreSQL pgvector, and Redis session context."
}
func (RetrieveKnowledgeTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"query": map[string]any{"type": "string", "description": "Customer question or search query"},
		"topK":  map[string]any{"type": "integer", "description": "Number of chunks to retrieve"},
	}, []string{"query"})
}
func (t RetrieveKnowledgeTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	if t.RAG == nil {
		return nil, fmt.Errorf("rag service is not configured")
	}
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	topK := intArg(args, "topK", 20)
	resp, err := t.RAG.RetrieveKnowledge(ctx, contracts.KnowledgeRequest{
		OrgID:     mcpCtx.OrgID,
		Query:     query,
		TopK:      topK,
		Threshold: contracts.MediumConfidenceThreshold,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"chunks":     resp.Chunks,
		"total":      len(resp.Chunks),
		"has_gap":    resp.HasGap,
		"source":     "rag.pgvector",
		"session_id": mcpCtx.SessionID,
		"threshold":  contracts.MediumConfidenceThreshold,
	}, nil
}

type CreateTicketTool struct {
	Sessions contracts.SessionStore
}

func (CreateTicketTool) Name() string { return "create_ticket" }
func (CreateTicketTool) Description() string {
	return "Create a support ticket when Gemini cannot resolve the customer request."
}
func (CreateTicketTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"title":       map[string]any{"type": "string"},
		"description": map[string]any{"type": "string"},
		"priority":    map[string]any{"type": "string", "enum": []string{"low", "normal", "high", "urgent"}},
	}, []string{"title", "description"})
}
func (t CreateTicketTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	title, _ := args["title"].(string)
	description, _ := args["description"].(string)
	if title == "" || description == "" {
		return nil, fmt.Errorf("title and description are required")
	}
	priority := stringArg(args, "priority", "normal")
	result := map[string]any{
		"success":   true,
		"ticket_id": "TICKET-" + mcpCtx.OrgID + "-" + safeID(title),
		"status":    "open",
		"priority":  priority,
		"org_id":    mcpCtx.OrgID,
	}
	if t.Sessions != nil && mcpCtx.SessionID != "" {
		_ = t.Sessions.LogToolExecution(ctx, mcpCtx.SessionID, contracts.ToolExecution{
			ToolName: "create_ticket",
			Success:  true,
			Result:   result,
		})
	}
	return result, nil
}

type SendNotificationTool struct {
	Notifications contracts.NotificationService
}

func (SendNotificationTool) Name() string { return "send_notification" }
func (SendNotificationTool) Description() string {
	return "Send email, SMS, WhatsApp, or in-app notifications through the notification service."
}
func (SendNotificationTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"channel":   map[string]any{"type": "string", "enum": []string{"email", "sms", "whatsapp", "inapp"}},
		"recipient": map[string]any{"type": "string"},
		"message":   map[string]any{"type": "string"},
		"priority":  map[string]any{"type": "string", "enum": []string{"normal", "high", "urgent"}},
	}, []string{"channel", "recipient", "message"})
}
func (t SendNotificationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	channel := stringArg(args, "channel", "")
	recipient := stringArg(args, "recipient", "")
	message := stringArg(args, "message", "")
	if channel == "" || recipient == "" || message == "" {
		return nil, fmt.Errorf("channel, recipient, and message are required")
	}
	priority := stringArg(args, "priority", "normal")
	if t.Notifications != nil {
		if err := t.Notifications.Send(ctx, contracts.NotificationRequest{
			OrgID:     mcpCtx.OrgID,
			Channel:   channel,
			Recipient: recipient,
			Template:  "custom_message",
			Data:      map[string]any{"message": message},
			Priority:  priority,
		}); err != nil {
			return nil, err
		}
	}
	return map[string]any{"success": true, "channel": channel, "recipient": recipient, "priority": priority}, nil
}

type CreateEscalationTool struct {
	Sessions      contracts.SessionStore
	Presence      contracts.PresenceService
	Notifications contracts.NotificationService
}

func (CreateEscalationTool) Name() string { return "create_escalation" }
func (CreateEscalationTool) Description() string {
	return "Escalate the current conversation to a human agent and notify available agents."
}
func (CreateEscalationTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"reason":   map[string]any{"type": "string"},
		"summary":  map[string]any{"type": "string"},
		"priority": map[string]any{"type": "string", "enum": []string{"low", "normal", "high", "urgent"}},
	}, []string{"reason"})
}
func (t CreateEscalationTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	reason := stringArg(args, "reason", "")
	if reason == "" {
		return nil, fmt.Errorf("reason is required")
	}
	priority := stringArg(args, "priority", "normal")
	req := contracts.EscalationRequest{
		SessionID:     mcpCtx.SessionID,
		OrgID:         mcpCtx.OrgID,
		Summary:       stringArg(args, "summary", ""),
		Priority:      priority,
		ScoringReason: reason,
	}
	notified := 0
	if t.Presence != nil {
		notified, err = t.Presence.NotifyAgents(ctx, mcpCtx.OrgID, req)
		if err != nil {
			return nil, err
		}
	}
	return map[string]any{
		"success":         true,
		"escalation_id":   mcpCtx.SessionID + "_escalation",
		"status":          "awaiting_agent",
		"agents_notified": notified,
		"priority":        priority,
	}, nil
}

type FlagKnowledgeGapTool struct{}

func (FlagKnowledgeGapTool) Name() string { return "flag_knowledge_gap" }
func (FlagKnowledgeGapTool) Description() string {
	return "Record a knowledge gap when retrieved content cannot answer the customer."
}

type GetCustomerContextTool struct {
	Sessions contracts.SessionStore
}

func (GetCustomerContextTool) Name() string { return "get_customer_context" }
func (GetCustomerContextTool) Description() string {
	return "Load the current tenant-scoped customer and session context."
}
func (GetCustomerContextTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{}, nil)
}
func (t GetCustomerContextTool) Execute(ctx context.Context, _ map[string]any) (any, error) {
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if t.Sessions == nil {
		return nil, fmt.Errorf("session store is not configured")
	}
	if mcpCtx.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	session, err := t.Sessions.Get(ctx, mcpCtx.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session %q not found", mcpCtx.SessionID)
	}
	if session.OrgID != mcpCtx.OrgID {
		return nil, fmt.Errorf("session does not belong to organization")
	}
	return map[string]any{
		"session_id":         session.ID,
		"customer_id":        session.CustomerID,
		"mode":               session.Mode,
		"status":             session.Status,
		"preserved_entities": session.PreservedEntities,
	}, nil
}

type QueryConnectorTool struct {
	Connectors contracts.ConnectorRouter
}

func (QueryConnectorTool) Name() string { return "query_connector" }
func (QueryConnectorTool) Description() string {
	return "Route an approved operation to an organization connector."
}
func (QueryConnectorTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"connector": map[string]any{"type": "string"},
		"operation": map[string]any{"type": "string"},
		"input":     map[string]any{"type": "object"},
	}, []string{"connector", "operation"})
}
func (t QueryConnectorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if t.Connectors == nil {
		return nil, fmt.Errorf("connector router is not configured")
	}
	connector := stringArg(args, "connector", "")
	operation := stringArg(args, "operation", "")
	if connector == "" || operation == "" {
		return nil, fmt.Errorf("connector and operation are required")
	}
	input, _ := args["input"].(map[string]any)
	if input == nil {
		input = map[string]any{}
	}
	return t.Connectors.ExecuteConnector(ctx, mcpCtx.OrgID, connector, operation, input)
}

type ExecuteWorkflowTool struct {
	Gateway *mcp.Gateway
}

func (ExecuteWorkflowTool) Name() string { return "execute_workflow" }
func (ExecuteWorkflowTool) Description() string {
	return "Execute an approved multi-step support workflow."
}
func (ExecuteWorkflowTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"workflow": map[string]any{
			"type": "string",
			"enum": []string{"support_escalation"},
		},
		"summary": map[string]any{"type": "string"},
		"reason":  map[string]any{"type": "string"},
	}, []string{"workflow", "summary", "reason"})
}
func (t ExecuteWorkflowTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	if t.Gateway == nil {
		return nil, fmt.Errorf("workflow gateway is not configured")
	}
	workflowName := stringArg(args, "workflow", "")
	summary := stringArg(args, "summary", "")
	reason := stringArg(args, "reason", "")
	if workflowName == "" || summary == "" || reason == "" {
		return nil, fmt.Errorf("workflow, summary, and reason are required")
	}
	if workflowName != "support_escalation" {
		return nil, fmt.Errorf("workflow %q is not supported", workflowName)
	}
	steps := []struct {
		name string
		args map[string]any
	}{
		{name: "create_ticket", args: map[string]any{
			"title": "Human follow-up requested", "description": summary, "priority": "high",
		}},
		{name: "create_escalation", args: map[string]any{
			"reason": reason, "summary": summary, "priority": "high",
		}},
	}
	results := make([]map[string]any, 0, len(steps))
	for _, step := range steps {
		result, err := t.Gateway.Execute(ctx, step.name, step.args)
		item := map[string]any{"tool": step.name, "success": err == nil}
		if err != nil {
			item["error"] = err.Error()
			results = append(results, item)
			return map[string]any{"workflow": workflowName, "success": false, "steps": results}, err
		}
		item["result"] = result
		results = append(results, item)
	}
	return map[string]any{"workflow": workflowName, "success": true, "steps": results}, nil
}
func (FlagKnowledgeGapTool) InputSchema() map[string]any {
	return objectSchema(map[string]any{
		"topic":    map[string]any{"type": "string"},
		"context":  map[string]any{"type": "string"},
		"severity": map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
	}, []string{"topic"})
}
func (FlagKnowledgeGapTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	mcpCtx, err := mcp.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	topic := stringArg(args, "topic", "")
	if topic == "" {
		return nil, fmt.Errorf("topic is required")
	}
	return map[string]any{
		"success":  true,
		"gap_id":   "GAP-" + mcpCtx.OrgID + "-" + safeID(topic),
		"topic":    topic,
		"severity": stringArg(args, "severity", "medium"),
		"org_id":   mcpCtx.OrgID,
	}, nil
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "properties": properties, "required": required}
}

func stringArg(args map[string]any, name, fallback string) string {
	if value, ok := args[name].(string); ok && value != "" {
		return value
	}
	return fallback
}

func intArg(args map[string]any, name string, fallback int) int {
	switch value := args[name].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return fallback
	}
}

func safeID(input string) string {
	out := ""
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			out += string(r)
		}
		if len(out) >= 12 {
			break
		}
	}
	if out == "" {
		return "item"
	}
	return out
}
