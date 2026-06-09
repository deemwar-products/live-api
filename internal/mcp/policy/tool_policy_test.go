package policy

import (
	"context"
	"testing"

	"live-api/internal/mcp"
)

func TestMVPPolicyAllowsCoreToolsForSupportedRoles(t *testing.T) {
	policy := MVPPolicy()
	for _, role := range []string{"", "ai", "agent", "admin", "unknown"} {
		for _, tool := range []string{
			"retrieve_knowledge", "create_ticket", "send_notification",
			"create_escalation", "flag_knowledge_gap", "execute_workflow",
			"query_connector", "get_customer_context",
		} {
			if err := policy.CanUseTool(context.Background(), mcp.MCPContext{Role: role}, tool); err != nil {
				t.Errorf("role %q tool %q: %v", role, tool, err)
			}
		}
	}
}

func TestMVPPolicyRejectsUnexposedTool(t *testing.T) {
	if err := MVPPolicy().CanUseTool(context.Background(), mcp.MCPContext{Role: "ai"}, "delete_customer"); err == nil {
		t.Fatal("expected policy rejection")
	}
}

func TestNilPolicyAllowsTool(t *testing.T) {
	var policy *StaticToolPolicy
	if err := policy.CanUseTool(context.Background(), mcp.MCPContext{}, "anything"); err != nil {
		t.Fatal(err)
	}
}
