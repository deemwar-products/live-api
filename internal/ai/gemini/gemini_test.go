package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/genai"

	"live-api/internal/mcp"
)

type gatewayStub struct {
	result any
	err    error
	name   string
	args   map[string]any
}

func (g *gatewayStub) Execute(_ context.Context, name string, args map[string]any) (any, error) {
	g.name, g.args = name, args
	return g.result, g.err
}

type declarationTool struct{ name string }

func (t declarationTool) Name() string                                       { return t.name }
func (declarationTool) Description() string                                  { return "description" }
func (declarationTool) InputSchema() map[string]any                          { return map[string]any{"type": "object"} }
func (declarationTool) Execute(context.Context, map[string]any) (any, error) { return nil, nil }

func TestBuildContextIsCompleteAndDeterministic(t *testing.T) {
	input := RequestContext{
		MCPContext:     mcp.MCPContext{OrgID: "org-1", SessionID: "session-1"},
		CustomerFacts:  map[string]string{"zeta": "last", "alpha": "first"},
		RecentMessages: []string{"customer: hello", "ai: hi"},
	}
	got := BuildContext(input)
	for _, value := range []string{"org_id=org-1", "session_id=session-1", "customer: hello"} {
		if !strings.Contains(got, value) {
			t.Fatalf("context missing %q: %s", value, got)
		}
	}
	if strings.Index(got, "alpha=first") > strings.Index(got, "zeta=last") {
		t.Fatalf("customer facts are not sorted: %s", got)
	}
}

func TestSafetyConfigBoundsToolCalls(t *testing.T) {
	tests := []struct{ input, want int }{{0, 3}, {-1, 3}, {2, 2}, {20, 5}}
	for _, tt := range tests {
		if got := (SafetyConfig{MaxToolCalls: tt.input}).EffectiveMaxToolCalls(); got != tt.want {
			t.Errorf("input %d: got %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestAIVisibleCoreToolsExposeOnlyEightTools(t *testing.T) {
	tools := AIVisibleCoreTools()
	if len(tools) != 8 {
		t.Fatalf("got %d tools", len(tools))
	}
	for _, tool := range tools {
		schema, ok := tool.ParametersJsonSchema.(map[string]any)
		if tool.Name == "" || !ok || schema["additionalProperties"] != false {
			t.Fatalf("invalid declaration: %#v", tool)
		}
	}
}

func TestCoreFunctionDeclarations(t *testing.T) {
	got := CoreFunctionDeclarations([]mcp.Tool{declarationTool{name: "echo"}})
	if len(got) != 1 || got[0].Name != "echo" || got[0].Description != "description" {
		t.Fatalf("unexpected declarations: %#v", got)
	}
}

func TestExecuteFunctionCall(t *testing.T) {
	gateway := &gatewayStub{result: struct {
		OK bool `json:"ok"`
	}{OK: true}}
	service := &Service{Gateway: gateway}
	got, err := service.ExecuteFunctionCall(context.Background(), &genai.FunctionCall{
		Name: "echo", Args: map[string]any{"value": "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["ok"] != true || gateway.name != "echo" {
		t.Fatalf("unexpected result %#v or gateway call %#v", got, gateway)
	}

	gateway.err = errors.New("failed")
	if _, err := service.ExecuteFunctionCall(context.Background(), &genai.FunctionCall{Name: "echo"}); err == nil {
		t.Fatal("expected gateway error")
	}
	if _, err := (&Service{}).ExecuteFunctionCall(context.Background(), &genai.FunctionCall{}); err == nil {
		t.Fatal("expected missing gateway error")
	}
}
