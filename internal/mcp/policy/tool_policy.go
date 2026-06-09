package policy

import (
	"context"
	"fmt"

	"live-api/internal/mcp"
)

type StaticToolPolicy struct {
	AllowedByRole map[string]map[string]bool
}

func MVPPolicy() *StaticToolPolicy {
	common := map[string]bool{
		"retrieve_knowledge":   true,
		"create_ticket":        true,
		"send_notification":    true,
		"create_escalation":    true,
		"flag_knowledge_gap":   true,
		"execute_workflow":     true,
		"query_connector":      true,
		"get_customer_context": true,
	}
	return &StaticToolPolicy{AllowedByRole: map[string]map[string]bool{
		"":      common,
		"ai":    common,
		"agent": common,
		"admin": common,
	}}
}

func (p *StaticToolPolicy) CanUseTool(_ context.Context, mcpCtx mcp.MCPContext, toolName string) error {
	if p == nil {
		return nil
	}
	allowed := p.AllowedByRole[mcpCtx.Role]
	if allowed == nil {
		allowed = p.AllowedByRole[""]
	}
	if allowed[toolName] {
		return nil
	}
	return fmt.Errorf("role %q cannot use tool %q", mcpCtx.Role, toolName)
}
