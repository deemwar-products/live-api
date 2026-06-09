package registry

import (
	"context"
	"testing"

	"live-api/internal/mcp"
)

type discoveryTool string

func (t discoveryTool) Name() string                                       { return string(t) }
func (t discoveryTool) Description() string                                { return string(t) }
func (discoveryTool) InputSchema() map[string]any                          { return nil }
func (discoveryTool) Execute(context.Context, map[string]any) (any, error) { return nil, nil }

func TestDiscoveryFiltersToolsByUseCase(t *testing.T) {
	registry := mcp.NewToolRegistry()
	for _, name := range []string{"retrieve_knowledge", "send_notification", "create_ticket", "private_tool"} {
		if err := registry.Register(discoveryTool(name)); err != nil {
			t.Fatal(err)
		}
	}
	discovery := NewDiscovery(registry)
	support := names(discovery.GetToolsForUseCase(UseCaseSupport))
	if support["send_notification"] || support["private_tool"] || !support["retrieve_knowledge"] {
		t.Fatalf("unexpected support tools: %#v", support)
	}
	admin := names(discovery.GetToolsForUseCase(UseCaseAdmin))
	if !admin["send_notification"] || admin["private_tool"] {
		t.Fatalf("unexpected admin tools: %#v", admin)
	}
}

func TestDiscoveryHandlesNilRegistry(t *testing.T) {
	if got := (&Discovery{}).GetToolsForUseCase(UseCaseSupport); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func names(tools []mcp.Tool) map[string]bool {
	out := make(map[string]bool)
	for _, tool := range tools {
		out[tool.Name()] = true
	}
	return out
}
