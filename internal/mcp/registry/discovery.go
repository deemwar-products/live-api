package registry

import "live-api/internal/mcp"

type UseCase string

const (
	UseCaseSupport UseCase = "support"
	UseCaseVoice   UseCase = "voice"
	UseCaseAdmin   UseCase = "admin"
)

type Discovery struct {
	Registry *mcp.ToolRegistry
}

func NewDiscovery(registry *mcp.ToolRegistry) *Discovery {
	return &Discovery{Registry: registry}
}

func (d *Discovery) GetToolsForUseCase(useCase UseCase) []mcp.Tool {
	if d == nil || d.Registry == nil {
		return nil
	}
	all := d.Registry.List()
	allowed := map[string]bool{}
	switch useCase {
	case UseCaseAdmin:
		allowed = allow("retrieve_knowledge", "create_ticket", "send_notification", "create_escalation", "flag_knowledge_gap", "execute_workflow", "query_connector", "get_customer_context")
	case UseCaseVoice, UseCaseSupport:
		allowed = allow("retrieve_knowledge", "create_ticket", "create_escalation", "flag_knowledge_gap", "execute_workflow", "query_connector", "get_customer_context")
	default:
		allowed = allow("retrieve_knowledge", "create_ticket", "send_notification", "create_escalation", "flag_knowledge_gap", "execute_workflow", "query_connector", "get_customer_context")
	}
	out := make([]mcp.Tool, 0, len(all))
	for _, tool := range all {
		if allowed[tool.Name()] {
			out = append(out, tool)
		}
	}
	return out
}

func allow(names ...string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		out[name] = true
	}
	return out
}
