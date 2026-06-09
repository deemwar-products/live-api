package gemini

import (
	"sort"
	"strings"

	"live-api/internal/mcp"
)

type RequestContext struct {
	MCPContext     mcp.MCPContext
	RecentMessages []string
	CustomerFacts  map[string]string
}

func BuildContext(input RequestContext) string {
	var b strings.Builder
	b.WriteString(ToolUseInstruction())
	b.WriteString("\n\nTenant:\n")
	b.WriteString("org_id=" + input.MCPContext.OrgID + "\n")
	if input.MCPContext.SessionID != "" {
		b.WriteString("session_id=" + input.MCPContext.SessionID + "\n")
	}
	if len(input.CustomerFacts) > 0 {
		b.WriteString("\nCustomer facts:\n")
		keys := make([]string, 0, len(input.CustomerFacts))
		for key := range input.CustomerFacts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := input.CustomerFacts[key]
			b.WriteString(key + "=" + value + "\n")
		}
	}
	if len(input.RecentMessages) > 0 {
		b.WriteString("\nRecent conversation:\n")
		for _, msg := range input.RecentMessages {
			b.WriteString(msg + "\n")
		}
	}
	return b.String()
}
