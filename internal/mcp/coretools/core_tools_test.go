package coretools

import (
	"context"
	"testing"

	"live-api/internal/mcp"
	"live-api/internal/mcp/contracts"
)

type ragStub struct {
	request contracts.KnowledgeRequest
}

func (s *ragStub) RetrieveKnowledge(_ context.Context, req contracts.KnowledgeRequest) (*contracts.KnowledgeResponse, error) {
	s.request = req
	return &contracts.KnowledgeResponse{
		Chunks: []contracts.KnowledgeChunk{{ID: "chunk-1", Score: 0.9}},
	}, nil
}

type notificationStub struct {
	request contracts.NotificationRequest
}

type sessionStub struct {
	session *contracts.Session
}

func (s *sessionStub) Get(context.Context, string) (*contracts.Session, error) { return s.session, nil }
func (*sessionStub) GetOrgConfig(context.Context, string) (*contracts.OrgConfig, error) {
	return nil, nil
}
func (*sessionStub) UpdateToolState(context.Context, string, map[string]any) error { return nil }
func (*sessionStub) LogToolExecution(context.Context, string, contracts.ToolExecution) error {
	return nil
}

type connectorStub struct {
	orgID, connector, operation string
}

func (s *connectorStub) ExecuteConnector(
	_ context.Context,
	orgID, connector, operation string,
	_ map[string]any,
) (map[string]any, error) {
	s.orgID, s.connector, s.operation = orgID, connector, operation
	return map[string]any{"found": true}, nil
}

func (s *notificationStub) Send(_ context.Context, req contracts.NotificationRequest) error {
	s.request = req
	return nil
}
func (*notificationStub) SendBatch(context.Context, []contracts.NotificationRequest) error {
	return nil
}

func TestNewCoreGatewayRegistersEnhancedToolSet(t *testing.T) {
	gateway, err := NewCoreGateway(nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if gateway.Registry.Count() != 8 {
		t.Fatalf("expected 8 Gemini-facing tools, got %d", gateway.Registry.Count())
	}
	for _, name := range []string{
		"retrieve_knowledge", "create_ticket", "send_notification",
		"create_escalation", "flag_knowledge_gap", "execute_workflow",
		"query_connector", "get_customer_context",
	} {
		if _, ok := gateway.Registry.Get(name); !ok {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestGetCustomerContextEnforcesTenant(t *testing.T) {
	sessions := &sessionStub{session: &contracts.Session{
		ID: "session-1", OrgID: "org-1", CustomerID: "customer-1", Status: "active",
	}}
	tool := GetCustomerContextTool{Sessions: sessions}
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{
		OrgID: "org-1", SessionID: "session-1",
	})
	result, err := tool.Execute(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.(map[string]any)["customer_id"] != "customer-1" {
		t.Fatalf("unexpected context: %#v", result)
	}
	badCtx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{
		OrgID: "other-org", SessionID: "session-1",
	})
	if _, err := tool.Execute(badCtx, nil); err == nil {
		t.Fatal("expected cross-tenant session rejection")
	}
}

func TestQueryConnectorPassesTenantToRouter(t *testing.T) {
	connectors := &connectorStub{}
	tool := QueryConnectorTool{Connectors: connectors}
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{OrgID: "org-2"})
	result, err := tool.Execute(ctx, map[string]any{
		"connector": "zendesk", "operation": "search_customer",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.(map[string]any)["found"] != true || connectors.orgID != "org-2" ||
		connectors.connector != "zendesk" || connectors.operation != "search_customer" {
		t.Fatalf("unexpected connector call: %#v %#v", connectors, result)
	}
}

func TestExecuteWorkflowRunsApprovedSteps(t *testing.T) {
	gateway, err := NewCoreGateway(nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{
		OrgID: "org-1", SessionID: "session-1", Role: "ai",
	})
	result, err := gateway.Execute(ctx, "execute_workflow", map[string]any{
		"workflow": "support_escalation",
		"summary":  "Customer needs human support",
		"reason":   "customer_request",
	})
	if err != nil {
		t.Fatal(err)
	}
	out := result.(map[string]any)
	if out["success"] != true || len(out["steps"].([]map[string]any)) != 2 {
		t.Fatalf("unexpected workflow result: %#v", out)
	}
}

func TestRetrieveKnowledgeUsesTenantAndDefaults(t *testing.T) {
	rag := &ragStub{}
	tool := RetrieveKnowledgeTool{RAG: rag}
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{
		OrgID: "org-7", SessionID: "session-9",
	})
	result, err := tool.Execute(ctx, map[string]any{"query": "reset password"})
	if err != nil {
		t.Fatal(err)
	}
	if rag.request.OrgID != "org-7" || rag.request.TopK != 20 {
		t.Fatalf("unexpected RAG request: %#v", rag.request)
	}
	out := result.(map[string]any)
	if out["session_id"] != "session-9" || out["total"] != 1 {
		t.Fatalf("unexpected result: %#v", out)
	}
}

func TestSendNotificationPassesTenantEnvelope(t *testing.T) {
	notifications := &notificationStub{}
	tool := SendNotificationTool{Notifications: notifications}
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{OrgID: "org-8"})
	_, err := tool.Execute(ctx, map[string]any{
		"channel": "email", "recipient": "user@example.com", "message": "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if notifications.request.OrgID != "org-8" || notifications.request.Priority != "normal" {
		t.Fatalf("unexpected notification: %#v", notifications.request)
	}
}

func TestCoreToolsValidateRequiredArguments(t *testing.T) {
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{OrgID: "org-1"})
	tests := []mcp.Tool{
		CreateTicketTool{},
		SendNotificationTool{},
		CreateEscalationTool{},
		FlagKnowledgeGapTool{},
		QueryConnectorTool{},
		ExecuteWorkflowTool{},
	}
	for _, tool := range tests {
		t.Run(tool.Name(), func(t *testing.T) {
			if _, err := tool.Execute(ctx, map[string]any{}); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
