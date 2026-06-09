package workflow

import (
	"context"
	"errors"
	"testing"

	"live-api/internal/mcp"
)

type workflowTool struct {
	name string
	err  error
}

func (t workflowTool) Name() string                { return t.name }
func (t workflowTool) Description() string         { return t.name }
func (t workflowTool) InputSchema() map[string]any { return nil }
func (t workflowTool) Execute(context.Context, map[string]any) (any, error) {
	return t.name + "-result", t.err
}

func TestEngineStopsAfterFailedStep(t *testing.T) {
	registry := mcp.NewToolRegistry()
	_ = registry.Register(workflowTool{name: "first"})
	_ = registry.Register(workflowTool{name: "second", err: errors.New("failed")})
	_ = registry.Register(workflowTool{name: "third"})
	engine := Engine{Gateway: mcp.NewGateway(registry, nil, nil)}
	ctx := mcp.WithMCPContext(context.Background(), mcp.MCPContext{OrgID: "org-1"})

	result := engine.Execute(ctx, []Step{
		{ToolName: "first"}, {ToolName: "second"}, {ToolName: "third"},
	})
	if len(result.Steps) != 2 {
		t.Fatalf("expected workflow to stop at failure, got %d steps", len(result.Steps))
	}
	if !result.Steps[0].Success || result.Steps[1].Success || result.Steps[1].Error != "failed" {
		t.Fatalf("unexpected workflow result: %#v", result)
	}
}

func TestSupportEscalationPlan(t *testing.T) {
	plan := SupportEscalationPlan("customer summary", "knowledge_gap")
	if len(plan) != 2 || plan[0].ToolName != "create_ticket" || plan[1].ToolName != "create_escalation" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}
