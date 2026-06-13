package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"google.golang.org/genai"
)

// MCPClient talks to an MCP server over the Streamable HTTP transport.
type MCPClient struct {
	url       string
	headers   map[string]string
	sessionID string
	seq       atomic.Int64
	http      *http.Client
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"` // may be object or string
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newMCPClient(url string, headers map[string]string) *MCPClient {
	return &MCPClient{
		url:     url,
		headers: headers,
		http:    &http.Client{},
	}
}

func (c *MCPClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	req := rpcRequest{
		JSONRPC: "2.0",
		ID:      c.seq.Add(1),
		Method:  method,
		Params:  params,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range c.headers {
		r.Header.Set(k, v)
	}
	if c.sessionID != "" {
		r.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.sessionID = sid
	}

	// Check HTTP-level errors before trying to parse JSON
	switch {
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return nil, fmt.Errorf("HTTP %d: authentication failed — check your token", resp.StatusCode)
	case resp.StatusCode == 404:
		return nil, fmt.Errorf("HTTP %d: endpoint not found — check the MCP URL", resp.StatusCode)
	case resp.StatusCode == 204:
		return nil, nil // no content — treat as success with no result
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	var rpc rpcResponse
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "text/event-stream") {
		rpc, err = parseSSEResponse(resp.Body)
	} else {
		err = json.NewDecoder(resp.Body).Decode(&rpc)
	}
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(rpc.Error) > 0 && string(rpc.Error) != "null" {
		// Error may be a JSON object {"code":...,"message":...} or a plain string
		var errObj rpcError
		if json.Unmarshal(rpc.Error, &errObj) == nil && errObj.Message != "" {
			return nil, fmt.Errorf("MCP %s error %d: %s", method, errObj.Code, errObj.Message)
		}
		var errStr string
		if json.Unmarshal(rpc.Error, &errStr) == nil {
			return nil, fmt.Errorf("MCP %s error: %s", method, errStr)
		}
		return nil, fmt.Errorf("MCP %s error: %s", method, rpc.Error)
	}
	return rpc.Result, nil
}

// notify sends a JSON-RPC notification (no response expected).
func (c *MCPClient) notify(ctx context.Context, method string, params any) error {
	req := rpcRequest{JSONRPC: "2.0", Method: method, Params: params}
	body, _ := json.Marshal(req)
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		r.Header.Set(k, v)
	}
	if c.sessionID != "" {
		r.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	resp, err := c.http.Do(r)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func parseSSEResponse(r io.Reader) (rpcResponse, error) {
	var rpc rpcResponse
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if err := json.Unmarshal([]byte(data), &rpc); err != nil {
				return rpc, err
			}
			return rpc, nil
		}
	}
	return rpc, fmt.Errorf("no data line in SSE stream")
}

// Initialize performs the MCP protocol handshake.
func (c *MCPClient) Initialize(ctx context.Context) error {
	_, err := c.call(ctx, "initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "gemini-live-poc", "version": "1.0.0"},
	})
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	return c.notify(ctx, "notifications/initialized", nil)
}

type mcpTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ListTools fetches the tool list and converts it to Gemini FunctionDeclarations.
func (c *MCPClient) ListTools(ctx context.Context) ([]*genai.FunctionDeclaration, error) {
	result, err := c.call(ctx, "tools/list", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("tools/list: %w", err)
	}
	var r struct {
		Tools []mcpTool `json:"tools"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		return nil, err
	}

	var decls []*genai.FunctionDeclaration
	for _, t := range r.Tools {
		decl := &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
		}
		if len(t.InputSchema) > 0 {
			var schema map[string]any
			if json.Unmarshal(t.InputSchema, &schema) == nil {
				decl.Parameters = jsonSchemaToGenai(schema)
			}
		}
		decls = append(decls, decl)
	}
	return decls, nil
}

// CallTool executes a named tool and returns the combined text output.
func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	result, err := c.call(ctx, "tools/call", map[string]any{
		"name":      name,
		"arguments": args,
	})
	if err != nil {
		return "", fmt.Errorf("tools/call %s: %w", name, err)
	}
	var r struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		return "", err
	}
	if r.IsError {
		return "", fmt.Errorf("tool %s returned an error", name)
	}
	var parts []string
	for _, item := range r.Content {
		if item.Type == "text" && item.Text != "" {
			parts = append(parts, item.Text)
		}
	}
	return strings.Join(parts, "\n"), nil
}

// jsonSchemaToGenai converts a JSON Schema map to a genai.Schema.
func jsonSchemaToGenai(s map[string]any) *genai.Schema {
	gs := &genai.Schema{}

	if t, _ := s["type"].(string); t != "" {
		gs.Type = genai.Type(strings.ToUpper(t))
	}
	if d, _ := s["description"].(string); d != "" {
		gs.Description = d
	}
	if props, ok := s["properties"].(map[string]any); ok {
		gs.Properties = make(map[string]*genai.Schema)
		for k, v := range props {
			if vm, ok := v.(map[string]any); ok {
				gs.Properties[k] = jsonSchemaToGenai(vm)
			}
		}
	}
	if req, ok := s["required"].([]any); ok {
		for _, r := range req {
			if rs, ok := r.(string); ok {
				gs.Required = append(gs.Required, rs)
			}
		}
	}
	if items, ok := s["items"].(map[string]any); ok {
		gs.Items = jsonSchemaToGenai(items)
	}
	if enums, ok := s["enum"].([]any); ok {
		for _, e := range enums {
			if es, ok := e.(string); ok {
				gs.Enum = append(gs.Enum, es)
			}
		}
	}
	return gs
}
