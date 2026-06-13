package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/genai"
)

// notionAPIClient is initialised at startup when NOTION_API_KEY is set.
var notionAPIClient *notionDirectClient

func init() {
	key := os.Getenv("NOTION_API_KEY")
	if key == "" {
		log.Printf("Notion: NOTION_API_KEY not set — notion tools disabled")
		return
	}
	// Log enough chars to diagnose format issues (never log the full secret)
	preview := key
	if len(preview) > 12 {
		preview = preview[:12] + "…"
	}
	log.Printf("Notion: API key loaded (%s), registering notion_search + notion_get_page", preview)
	notionAPIClient = &notionDirectClient{apiKey: key, http: &http.Client{}}
	builtinDeclarations = append(builtinDeclarations, notionBuiltinDeclarations...)
}

var notionBuiltinDeclarations = []*genai.FunctionDeclaration{
	{
		Name:        "notion_search",
		Description: "Search the Notion workspace for pages matching a keyword or topic. Use this to find company information, policies, documentation, SOPs, and meeting notes. Returns page titles and IDs.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "Keywords to search for (e.g. 'leave policy', 'onboarding', 'product roadmap')",
				},
			},
			Required: []string{"query"},
		},
	},
	{
		Name:        "notion_get_page",
		Description: "Read the full text content of a Notion page by its ID. Call this after notion_search to retrieve complete information from a specific page.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"page_id": {
					Type:        genai.TypeString,
					Description: "Notion page ID returned by notion_search (with or without dashes)",
				},
			},
			Required: []string{"page_id"},
		},
	},
}

// notionDirectClient calls https://api.notion.com using an Integration secret.
type notionDirectClient struct {
	apiKey string
	http   *http.Client
}

func (c *notionDirectClient) do(ctx context.Context, method, path string, body any) (map[string]any, error) {
	var buf *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewReader(b)
	} else {
		buf = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.notion.com/v1"+path, buf)
	if err != nil {
		return nil, err
	}
	// Show only first 12 chars of key to catch format issues (e.g. accidental quotes)
	keyPreview := c.apiKey
	if len(keyPreview) > 12 {
		keyPreview = keyPreview[:12] + "…"
	}
	log.Printf("Notion → %s %s  [key: %s]", method, path, keyPreview)

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	log.Printf("Notion ← HTTP %d  body: %.300s", resp.StatusCode, rawBody)

	var result map[string]any
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return nil, fmt.Errorf("notion: decode response: %w", err)
	}
	if resp.StatusCode >= 400 {
		if msg, ok := result["message"].(string); ok {
			return nil, fmt.Errorf("notion API %d: %s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("notion API %d", resp.StatusCode)
	}
	return result, nil
}

func (c *notionDirectClient) search(ctx context.Context, query string) (string, error) {
	result, err := c.do(ctx, "POST", "/search", map[string]any{
		"query":     query,
		"page_size": 8,
		"filter":    map[string]any{"value": "page", "property": "object"},
		"sort":      map[string]any{"direction": "descending", "timestamp": "last_edited_time"},
	})
	if err != nil {
		return "", err
	}

	results, _ := result["results"].([]any)
	if len(results) == 0 {
		return "No pages found in Notion matching your query.", nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d page(s) in Notion:\n\n", len(results))
	for _, r := range results {
		page, _ := r.(map[string]any)
		title := notionPageTitle(page)
		id := strings.ReplaceAll(fmt.Sprint(page["id"]), "-", "")
		url, _ := page["url"].(string)
		fmt.Fprintf(&sb, "• %s\n  ID: %s\n  URL: %s\n\n", title, id, url)
	}
	return sb.String(), nil
}

func (c *notionDirectClient) getPage(ctx context.Context, pageID string) (string, error) {
	pageID = normalisePageID(pageID)

	page, err := c.do(ctx, "GET", "/pages/"+pageID, nil)
	if err != nil {
		return "", err
	}
	title := notionPageTitle(page)

	blocks, err := c.do(ctx, "GET", "/blocks/"+pageID+"/children?page_size=100", nil)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", title)
	for _, b := range asSlice(blocks["results"]) {
		if line := notionBlockText(asMap(b)); line != "" {
			sb.WriteString(line + "\n")
		}
	}
	return sb.String(), nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func normalisePageID(id string) string {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) == 32 {
		return id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	}
	return id
}

func notionPageTitle(page map[string]any) string {
	props, _ := page["properties"].(map[string]any)
	for _, key := range []string{"title", "Title", "Name", "name"} {
		prop, ok := props[key].(map[string]any)
		if !ok {
			continue
		}
		for _, richKey := range []string{"title", "rich_text"} {
			arr := asSlice(prop[richKey])
			if len(arr) > 0 {
				if t, ok := asMap(arr[0])["plain_text"].(string); ok && t != "" {
					return t
				}
			}
		}
	}
	return "(Untitled)"
}

func notionBlockText(block map[string]any) string {
	blockType, _ := block["type"].(string)
	content := asMap(block[blockType])

	var sb strings.Builder
	for _, rt := range asSlice(content["rich_text"]) {
		if t, ok := asMap(rt)["plain_text"].(string); ok {
			sb.WriteString(t)
		}
	}
	t := sb.String()
	if t == "" {
		return ""
	}

	switch blockType {
	case "heading_1":
		return "# " + t
	case "heading_2":
		return "## " + t
	case "heading_3":
		return "### " + t
	case "bulleted_list_item":
		return "• " + t
	case "numbered_list_item":
		return "1. " + t
	case "to_do":
		if checked, _ := content["checked"].(bool); checked {
			return "[x] " + t
		}
		return "[ ] " + t
	case "quote":
		return "> " + t
	case "code":
		return "```\n" + t + "\n```"
	case "divider":
		return "---"
	default:
		return t
	}
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func asSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}
