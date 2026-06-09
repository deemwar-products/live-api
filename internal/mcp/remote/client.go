package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	official "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Client struct {
	session *official.ClientSession
}

func ConnectCommand(ctx context.Context, name string, args ...string) (*Client, error) {
	client := official.NewClient(&official.Implementation{
		Name:    "live-api-mcp-gateway",
		Version: "v1.0.0",
	}, nil)
	session, err := client.Connect(ctx, &official.CommandTransport{
		Command: exec.Command(name, args...),
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Client{session: session}, nil
}

func (c *Client) Close() error {
	if c == nil || c.session == nil {
		return nil
	}
	return c.session.Close()
}

func (c *Client) DiscoverTools(ctx context.Context) ([]*official.Tool, error) {
	if c == nil || c.session == nil {
		return nil, fmt.Errorf("remote MCP client is not connected")
	}
	result, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

func (c *Client) ExecuteTool(ctx context.Context, name string, args map[string]any) (map[string]any, error) {
	if c == nil || c.session == nil {
		return nil, fmt.Errorf("remote MCP client is not connected")
	}
	if name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	result, err := c.session.CallTool(ctx, &official.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, err
	}
	if result.IsError {
		return nil, result.GetError()
	}
	if structured, ok := result.StructuredContent.(map[string]any); ok {
		return structured, nil
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("mcp result was not an object: %w", err)
	}
	return out, nil
}
