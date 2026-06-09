package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	Execute(ctx context.Context, args map[string]any) (any, error)
}

type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]Tool)}
}

func (r *ToolRegistry) Register(tool Tool) error {
	if tool == nil || tool.Name() == "" {
		return fmt.Errorf("invalid tool")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %q is already registered", tool.Name())
	}
	r.tools[tool.Name()] = tool
	return nil
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *ToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		out = append(out, tool)
	}
	return out
}

func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

type Gateway struct {
	Registry    *ToolRegistry
	Audit       AuditSink
	RateLimiter *RateLimiter
	Policy      ToolPolicy
}

func NewGateway(registry *ToolRegistry, audit AuditSink, limiter *RateLimiter) *Gateway {
	return NewGatewayWithPolicy(registry, audit, limiter, nil)
}

func NewGatewayWithPolicy(registry *ToolRegistry, audit AuditSink, limiter *RateLimiter, policy ToolPolicy) *Gateway {
	if registry == nil {
		registry = NewToolRegistry()
	}
	if audit == nil {
		audit = NoopAuditSink{}
	}
	return &Gateway{Registry: registry, Audit: audit, RateLimiter: limiter, Policy: policy}
}

func (g *Gateway) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
	mcpCtx, err := RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	tool, ok := g.Registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	if g.Policy != nil {
		if err := g.Policy.CanUseTool(ctx, mcpCtx, name); err != nil {
			return nil, err
		}
	}
	if g.RateLimiter != nil {
		if err := g.RateLimiter.Allow(mcpCtx.OrgID, name); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	result, execErr := tool.Execute(ctx, args)
	_ = g.Audit.Record(ctx, ToolExecution{
		ToolName:   name,
		OrgID:      mcpCtx.OrgID,
		UserID:     mcpCtx.UserID,
		SessionID:  mcpCtx.SessionID,
		Duration:   time.Since(start),
		Success:    execErr == nil,
		Error:      errorString(execErr),
		ExecutedAt: time.Now(),
	})
	return result, execErr
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
