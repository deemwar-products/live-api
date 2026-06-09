package mcp

import "context"

type ToolPolicy interface {
	CanUseTool(ctx context.Context, mcpCtx MCPContext, toolName string) error
}
