package gemini

import (
	"context"
	"fmt"

	mcpctx "live-api/internal/mcp"
	"live-api/internal/mcp/registry"

	"google.golang.org/genai"
)

type Orchestrator struct {
	Gemini *Service
	Router ToolRouter
	Safety SafetyConfig
}

func (o *Orchestrator) Respond(ctx context.Context, req RequestContext, userText string) (string, error) {
	if o == nil || o.Gemini == nil {
		return "", fmt.Errorf("gemini orchestrator is not configured")
	}
	ctx = mcpctx.WithMCPContext(ctx, req.MCPContext)
	tools := o.Router.Declarations(registry.UseCaseSupport)
	systemContext := BuildContext(req)
	contents := []*genai.Content{
		genai.NewContentFromText(systemContext, genai.RoleUser),
		genai.NewContentFromText(userText, genai.RoleUser),
	}

	maxCalls := o.Safety.EffectiveMaxToolCalls()
	for i := 0; i <= maxCalls; i++ {
		resp, err := o.Gemini.Client.Models.GenerateContent(ctx, o.Gemini.Model, contents, &genai.GenerateContentConfig{
			Tools: []*genai.Tool{{FunctionDeclarations: tools}},
		})
		if err != nil {
			return "", err
		}
		calls := resp.FunctionCalls()
		if len(calls) == 0 {
			return resp.Text(), nil
		}
		if i == maxCalls {
			return "", fmt.Errorf("max gemini tool calls exceeded")
		}
		for _, call := range calls {
			result, err := o.Gemini.ExecuteFunctionCall(ctx, call)
			if err != nil {
				result = map[string]any{"error": err.Error()}
			}
			contents = append(contents,
				genai.NewContentFromFunctionCall(call.Name, call.Args, genai.RoleUser),
				genai.NewContentFromFunctionResponse(call.Name, result, genai.RoleUser),
			)
		}
	}
	return "", fmt.Errorf("gemini orchestration ended unexpectedly")
}
