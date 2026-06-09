package workflow

import (
	"context"
	"fmt"

	"live-api/internal/mcp"
)

type Step struct {
	ToolName string
	Args     map[string]any
}

type Result struct {
	Steps []StepResult `json:"steps"`
}

type StepResult struct {
	ToolName string `json:"tool_name"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Result   any    `json:"result,omitempty"`
}

type Engine struct {
	Gateway *mcp.Gateway
}

func (e *Engine) Execute(ctx context.Context, steps []Step) Result {
	result := Result{Steps: make([]StepResult, 0, len(steps))}
	if e == nil || e.Gateway == nil {
		result.Steps = append(result.Steps, StepResult{
			ToolName: "workflow",
			Success:  false,
			Error:    "workflow gateway is not configured",
		})
		return result
	}
	for _, step := range steps {
		if step.ToolName == "" {
			result.Steps = append(result.Steps, StepResult{
				ToolName: "workflow",
				Success:  false,
				Error:    fmt.Sprintf("step %d has no tool name", len(result.Steps)),
			})
			break
		}
		value, err := e.Gateway.Execute(ctx, step.ToolName, step.Args)
		item := StepResult{ToolName: step.ToolName, Success: err == nil, Result: value}
		if err != nil {
			item.Error = err.Error()
			result.Steps = append(result.Steps, item)
			break
		}
		result.Steps = append(result.Steps, item)
	}
	return result
}
