package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"live-api/internal/mcp"
	"live-api/internal/mcp/coretools"
	"live-api/internal/mcp/policy"
)

func newTestApp(t *testing.T, limit int) *app {
	t.Helper()
	audit := mcp.NewInMemoryAuditSink()
	gateway, err := coretools.NewCoreGateway(nil, nil, nil, nil, nil, audit, mcp.NewRateLimiter(limit), policy.MVPPolicy())
	if err != nil {
		t.Fatal(err)
	}
	return &app{gateway: gateway, audit: audit}
}

func TestHealthAndToolDiscoveryHandlers(t *testing.T) {
	server := newTestApp(t, 10)
	health := httptest.NewRecorder()
	server.health(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status: %d", health.Code)
	}
	var healthBody map[string]any
	_ = json.Unmarshal(health.Body.Bytes(), &healthBody)
	if healthBody["architecture"] != "mcp_gateway" || healthBody["tools_registered"] != float64(8) {
		t.Fatalf("unexpected health: %#v", healthBody)
	}

	tools := httptest.NewRecorder()
	server.listTools(tools, httptest.NewRequest(http.MethodGet, "/api/tools", nil))
	var toolsBody struct {
		Count int `json:"count"`
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(tools.Body.Bytes(), &toolsBody); err != nil {
		t.Fatal(err)
	}
	if toolsBody.Count != 8 || toolsBody.Tools[0].Name != "create_escalation" {
		t.Fatalf("unexpected tools response: %#v", toolsBody)
	}
}

func TestExecuteToolHandlerAndAudit(t *testing.T) {
	server := newTestApp(t, 10)
	body := []byte(`{"tool_name":"create_ticket","org_id":"org-1","session_id":"session-1","role":"ai","arguments":{"title":"Help","description":"Details"}}`)
	recorder := httptest.NewRecorder()
	server.executeTool(recorder, httptest.NewRequest(http.MethodPost, "/api/execute", bytes.NewReader(body)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body.String())
	}
	if entries := server.audit.List(); len(entries) != 1 || entries[0].ToolName != "create_ticket" {
		t.Fatalf("unexpected audit entries: %#v", entries)
	}
}

func TestExecuteToolHandlerRejectsBadRequests(t *testing.T) {
	server := newTestApp(t, 1)
	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "invalid JSON", body: `{`, want: http.StatusBadRequest},
		{name: "missing tenant", body: `{"tool_name":"create_ticket"}`, want: http.StatusBadRequest},
		{name: "unknown tool", body: `{"tool_name":"missing","org_id":"org-1"}`, want: http.StatusNotFound},
		{name: "invalid arguments", body: `{"tool_name":"create_ticket","org_id":"org-2","role":"ai","arguments":{}}`, want: http.StatusUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.executeTool(recorder, httptest.NewRequest(http.MethodPost, "/api/execute", bytes.NewBufferString(tt.body)))
			if recorder.Code != tt.want {
				t.Fatalf("got %d, want %d: %s", recorder.Code, tt.want, recorder.Body.String())
			}
		})
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("MCP_TEST_ADDR", ":9999")
	if got := envOr("MCP_TEST_ADDR", ":8080"); got != ":9999" {
		t.Fatalf("got %q", got)
	}
}
