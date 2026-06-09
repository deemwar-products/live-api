package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	domainmcp "live-api/internal/mcp"

	official "github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"
)

func CoreFunctionDeclarations(tools []domainmcp.Tool) []*genai.FunctionDeclaration {
	out := make([]*genai.FunctionDeclaration, 0, len(tools))
	for _, tool := range tools {
		out = append(out, &genai.FunctionDeclaration{
			Name:                 tool.Name(),
			Description:          tool.Description(),
			ParametersJsonSchema: tool.InputSchema(),
		})
	}
	return out
}

func MCPToolsToFunctionDeclarations(tools []*official.Tool) []*genai.FunctionDeclaration {
	out := make([]*genai.FunctionDeclaration, 0, len(tools))
	for _, tool := range tools {
		out = append(out, &genai.FunctionDeclaration{
			Name:                 tool.Name,
			Description:          tool.Description,
			ParametersJsonSchema: tool.InputSchema,
		})
	}
	return out
}

func AIVisibleCoreTools() []*genai.FunctionDeclaration {
	return []*genai.FunctionDeclaration{
		function("retrieve_knowledge", "Search organization knowledge using RAG, pgvector, and Redis session context.", objectSchema(map[string]any{
			"query": map[string]any{"type": "string", "description": "Customer question or retrieval query"},
			"topK":  map[string]any{"type": "integer", "description": "Maximum knowledge chunks to return"},
		}, []string{"query"})),
		function("create_ticket", "Create a support ticket when Gemini cannot resolve the customer request.", objectSchema(map[string]any{
			"title":       map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
			"priority":    map[string]any{"type": "string", "enum": []string{"low", "normal", "high", "urgent"}},
		}, []string{"title", "description"})),
		function("send_notification", "Send an email, SMS, WhatsApp, or in-app notification through the notification service.", objectSchema(map[string]any{
			"channel":   map[string]any{"type": "string", "enum": []string{"email", "sms", "whatsapp", "inapp"}},
			"recipient": map[string]any{"type": "string"},
			"message":   map[string]any{"type": "string"},
			"priority":  map[string]any{"type": "string", "enum": []string{"normal", "high", "urgent"}},
		}, []string{"channel", "recipient", "message"})),
		function("create_escalation", "Escalate the conversation to a human agent and notify available agents.", objectSchema(map[string]any{
			"reason":   map[string]any{"type": "string"},
			"summary":  map[string]any{"type": "string"},
			"priority": map[string]any{"type": "string", "enum": []string{"low", "normal", "high", "urgent"}},
		}, []string{"reason"})),
		function("flag_knowledge_gap", "Record a knowledge gap when retrieved content cannot answer the customer.", objectSchema(map[string]any{
			"topic":    map[string]any{"type": "string"},
			"context":  map[string]any{"type": "string"},
			"severity": map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
		}, []string{"topic"})),
		function("execute_workflow", "Execute an approved multi-step support workflow.", objectSchema(map[string]any{
			"workflow": map[string]any{"type": "string", "enum": []string{"support_escalation"}},
			"summary":  map[string]any{"type": "string"},
			"reason":   map[string]any{"type": "string"},
		}, []string{"workflow", "summary", "reason"})),
		function("query_connector", "Route an approved operation to an organization connector.", objectSchema(map[string]any{
			"connector": map[string]any{"type": "string"},
			"operation": map[string]any{"type": "string"},
			"input":     map[string]any{"type": "object"},
		}, []string{"connector", "operation"})),
		function("get_customer_context", "Load the current tenant-scoped customer and session context.", objectSchema(map[string]any{}, nil)),
	}
}

func (s *Service) ExecuteFunctionCall(ctx context.Context, call *genai.FunctionCall) (map[string]any, error) {
	if s.Gateway == nil {
		return nil, fmt.Errorf("gemini gateway is not configured")
	}
	result, err := s.Gateway.Execute(ctx, call.Name, call.Args)
	if err != nil {
		return nil, err
	}
	if out, ok := result.(map[string]any); ok {
		return out, nil
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func function(name, description string, schema map[string]any) *genai.FunctionDeclaration {
	return &genai.FunctionDeclaration{Name: name, Description: description, ParametersJsonSchema: schema}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}
