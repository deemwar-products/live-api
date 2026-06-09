package mcp

import (
	"context"
	"fmt"
)

// MCPContext carries the tenant/session envelope every tool must execute under.
type MCPContext struct {
	OrgID     string `json:"org_id"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Role      string `json:"role,omitempty"`
}

type mcpContextKey struct{}

func WithMCPContext(ctx context.Context, mcpCtx MCPContext) context.Context {
	return context.WithValue(ctx, mcpContextKey{}, mcpCtx)
}

func ContextFrom(ctx context.Context) (MCPContext, bool) {
	mcpCtx, ok := ctx.Value(mcpContextKey{}).(MCPContext)
	return mcpCtx, ok
}

func RequireContext(ctx context.Context) (MCPContext, error) {
	mcpCtx, ok := ContextFrom(ctx)
	if !ok || mcpCtx.OrgID == "" {
		return MCPContext{}, fmt.Errorf("mcp context with org_id is required")
	}
	return mcpCtx, nil
}
