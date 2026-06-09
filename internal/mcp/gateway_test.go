package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type testTool struct {
	name string
	run  func(context.Context, map[string]any) (any, error)
}

func (t testTool) Name() string                { return t.name }
func (t testTool) Description() string         { return "test tool" }
func (t testTool) InputSchema() map[string]any { return map[string]any{"type": "object"} }
func (t testTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	return t.run(ctx, args)
}

type denyingPolicy struct{}

func (denyingPolicy) CanUseTool(context.Context, MCPContext, string) error {
	return errors.New("denied")
}

func TestGatewayExecuteRequiresTenantContext(t *testing.T) {
	gateway := NewGateway(nil, nil, nil)
	_, err := gateway.Execute(context.Background(), "missing", nil)
	if err == nil || !strings.Contains(err.Error(), "org_id") {
		t.Fatalf("expected org_id error, got %v", err)
	}
}

func TestGatewayExecuteAuditsSuccessAndFailure(t *testing.T) {
	registry := NewToolRegistry()
	if err := registry.Register(testTool{name: "echo", run: func(_ context.Context, args map[string]any) (any, error) {
		if args["fail"] == true {
			return nil, errors.New("boom")
		}
		return args, nil
	}}); err != nil {
		t.Fatal(err)
	}
	audit := NewInMemoryAuditSink()
	gateway := NewGateway(registry, audit, nil)
	ctx := WithMCPContext(context.Background(), MCPContext{
		OrgID: "org-1", UserID: "user-1", SessionID: "session-1",
	})

	if _, err := gateway.Execute(ctx, "echo", map[string]any{"value": "ok"}); err != nil {
		t.Fatalf("execute success: %v", err)
	}
	if _, err := gateway.Execute(ctx, "echo", map[string]any{"fail": true}); err == nil {
		t.Fatal("expected tool failure")
	}

	entries := audit.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(entries))
	}
	if !entries[0].Success || entries[0].OrgID != "org-1" || entries[0].UserID != "user-1" {
		t.Fatalf("unexpected success audit: %#v", entries[0])
	}
	if entries[1].Success || entries[1].Error != "boom" {
		t.Fatalf("unexpected failure audit: %#v", entries[1])
	}
}

func TestGatewayPolicyAndRateLimitRunBeforeTool(t *testing.T) {
	registry := NewToolRegistry()
	calls := 0
	_ = registry.Register(testTool{name: "echo", run: func(context.Context, map[string]any) (any, error) {
		calls++
		return "ok", nil
	}})
	ctx := WithMCPContext(context.Background(), MCPContext{OrgID: "org-1"})

	denied := NewGatewayWithPolicy(registry, nil, nil, denyingPolicy{})
	if _, err := denied.Execute(ctx, "echo", nil); err == nil {
		t.Fatal("expected policy denial")
	}

	limited := NewGateway(registry, nil, NewRateLimiter(1))
	if _, err := limited.Execute(ctx, "echo", nil); err != nil {
		t.Fatalf("first execution: %v", err)
	}
	if _, err := limited.Execute(ctx, "echo", nil); err == nil {
		t.Fatal("expected quota error")
	}
	if calls != 1 {
		t.Fatalf("expected one tool call, got %d", calls)
	}
}

func TestToolRegistryRejectsDuplicateNames(t *testing.T) {
	registry := NewToolRegistry()
	tool := testTool{name: "echo", run: func(context.Context, map[string]any) (any, error) {
		return nil, nil
	}}
	if err := registry.Register(tool); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(tool); err == nil {
		t.Fatal("expected duplicate registration error")
	}
	if registry.Count() != 1 {
		t.Fatalf("expected one registered tool, got %d", registry.Count())
	}
}

func TestBoundedAuditSinkKeepsNewestEntries(t *testing.T) {
	sink := NewBoundedAuditSink(2)
	for _, name := range []string{"one", "two", "three"} {
		_ = sink.Record(context.Background(), ToolExecution{ToolName: name})
	}
	entries := sink.List()
	if len(entries) != 2 || entries[0].ToolName != "two" || entries[1].ToolName != "three" {
		t.Fatalf("unexpected bounded entries: %#v", entries)
	}
}
