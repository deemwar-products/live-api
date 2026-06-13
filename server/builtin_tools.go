package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// builtinDeclarations are always registered with Gemini, regardless of MCP servers.
// notion.go's init() appends Notion tools when NOTION_API_KEY is set.
var builtinDeclarations = []*genai.FunctionDeclaration{
	{
		Name:        "get_current_time",
		Description: "Returns the current date and time in UTC. Call this when the user asks what time or date it is.",
	},
	{
		Name:        "echo",
		Description: "Echoes a message back verbatim. Useful for testing the tool-calling pipeline.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"message": {
					Type:        genai.TypeString,
					Description: "The message to echo back.",
				},
			},
			Required: []string{"message"},
		},
	},
}

// callBuiltin dispatches a built-in tool call.
// Returns (output, true, nil) on success, ("", true, err) on tool error, ("", false, nil) if unknown.
func callBuiltin(ctx context.Context, name string, args map[string]any) (string, bool, error) {
	switch name {
	case "get_current_time":
		return time.Now().UTC().Format("Monday, 02 Jan 2006 15:04:05 UTC"), true, nil
	case "echo":
		msg, _ := args["message"].(string)
		return fmt.Sprintf("echo: %s", msg), true, nil
	case "notion_search":
		if notionAPIClient == nil {
			return "", true, fmt.Errorf("Notion not configured — set NOTION_API_KEY in server/.env")
		}
		query, _ := args["query"].(string)
		text, err := notionAPIClient.search(ctx, query)
		return text, true, err
	case "notion_get_page":
		if notionAPIClient == nil {
			return "", true, fmt.Errorf("Notion not configured — set NOTION_API_KEY in server/.env")
		}
		pageID, _ := args["page_id"].(string)
		text, err := notionAPIClient.getPage(ctx, pageID)
		return text, true, err
	}
	return "", false, nil
}
