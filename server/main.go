package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	_ "embed"

	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

//go:embed live_streaming.html
var homeTemplate string

const geminiModel = "models/gemini-3.1-flash-live-preview"

var (
	addr = flag.String("addr", ":8080", "HTTP service address")

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	flag.Parse()

	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/live", liveHandler)
	mux.HandleFunc("/api/defaults", defaultsHandler)
	mux.HandleFunc("/api/notion-test", notionTestHandler)
	mux.HandleFunc("/proxyVideo", proxyVideo)

	log.Printf("Server listening on http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(homeTemplate)
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}
	wsURL := "ws://" + r.Host + "/live"
	if err := tmpl.Execute(w, wsURL); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

// liveHandler is the core of the POC:
//  1. Reads an init message from the browser (carries mode + MCP server list)
//  2. Connects to each MCP server, fetches its tools, builds a registry
//  3. Registers all tools with Gemini via LiveConnectConfig.Tools
//  4. Proxies audio/text between the browser and Gemini
//  5. Intercepts Gemini's ToolCall messages, dispatches them to the right
//     MCP server, and returns results — the browser never sees raw tool calls
func liveHandler(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer clientConn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// ── 1. Init message ───────────────────────────────────────
	// First WebSocket frame from the browser:
	// {"type":"init","mode":"audio","mcpServers":[{"name":"...","url":"...","auth":"..."}]}
	var initMsg struct {
		Type              string `json:"type"`
		Mode              string `json:"mode"`
		SystemInstruction string `json:"systemInstruction"`
		MCPServers        []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			Auth string `json:"auth"`
		} `json:"mcpServers"`
	}
	_, initRaw, err := clientConn.ReadMessage()
	if err != nil {
		log.Printf("read init message error: %v", err)
		return
	}
	if err := json.Unmarshal(initRaw, &initMsg); err != nil || initMsg.Type != "init" {
		log.Printf("invalid init message: %s", initRaw)
		return
	}

	mode := initMsg.Mode
	if mode == "" {
		mode = "audio"
	}

	// Collect server configs.
	// Env-var server (MCP_URL/MCP_AUTH/MCP_NAME) is always included when set —
	// it is a server-side secret so its auth never travels through the browser.
	// UI-configured servers (from the init message) are merged in, skipping
	// any URL that duplicates the env-var server to avoid double-connecting.
	type serverCfg struct{ Name, URL, Auth string }
	var servers []serverCfg
	envURL := os.Getenv("MCP_URL")
	if envURL != "" {
		name := os.Getenv("MCP_NAME")
		if name == "" {
			name = "default"
		}
		servers = append(servers, serverCfg{Name: name, URL: envURL, Auth: os.Getenv("MCP_AUTH")})
	}
	for _, s := range initMsg.MCPServers {
		if s.URL != "" && s.URL != envURL {
			servers = append(servers, serverCfg{s.Name, s.URL, s.Auth})
		}
	}

	// ── 2. MCP tool registry ──────────────────────────────────
	// One MCPClient per server; tools are keyed by name so we know
	// which server to call when Gemini invokes a function.
	toolRegistry := map[string]*MCPClient{} // tool name → owning client
	var allTools []*genai.FunctionDeclaration

	type mcpStatus struct {
		Name      string   `json:"name"`
		OK        bool     `json:"ok"`
		Error     string   `json:"error,omitempty"`
		Tools     int      `json:"tools"`
		ToolNames []string `json:"toolNames,omitempty"`
	}

	// Built-in tools are always available — no MCP server required.
	builtinNames := make([]string, len(builtinDeclarations))
	for i, d := range builtinDeclarations {
		builtinNames[i] = d.Name
	}
	allTools = append(allTools, builtinDeclarations...)
	statuses := []mcpStatus{{
		Name: "Built-in", OK: true,
		Tools: len(builtinDeclarations), ToolNames: builtinNames,
	}}

	for _, srv := range servers {
		headers := map[string]string{}
		if srv.Auth != "" {
			auth := srv.Auth
			if !strings.HasPrefix(auth, "Bearer ") && !strings.HasPrefix(auth, "Basic ") {
				auth = "Bearer " + auth
			}
			headers["Authorization"] = auth
		}
		mcp := newMCPClient(srv.URL, headers)
		if err := mcp.Initialize(ctx); err != nil {
			log.Printf("MCP [%s] init error: %v", srv.Name, err)
			statuses = append(statuses, mcpStatus{Name: srv.Name, Error: err.Error()})
			continue
		}
		tools, err := mcp.ListTools(ctx)
		if err != nil {
			log.Printf("MCP [%s] list tools error: %v", srv.Name, err)
			statuses = append(statuses, mcpStatus{Name: srv.Name, Error: err.Error()})
			continue
		}
		log.Printf("MCP [%s]: %d tools registered", srv.Name, len(tools))
		var names []string
		for _, t := range tools {
			log.Printf("  • %s", t.Name)
			toolRegistry[t.Name] = mcp
			allTools = append(allTools, t)
			names = append(names, t.Name)
		}
		statuses = append(statuses, mcpStatus{Name: srv.Name, OK: true, Tools: len(tools), ToolNames: names})
	}

	// Always tell the browser which servers and tools are active (shown in ⚙ settings)
	msg, _ := json.Marshal(map[string]any{"type": "mcpStatus", "servers": statuses})
	_ = clientConn.WriteMessage(websocket.TextMessage, msg)

	// ── 3. Gemini session ─────────────────────────────────────
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			APIVersion: "v1alpha",
		},
	})
	if err != nil {
		log.Printf("genai client error: %v", err)
		return
	}

	config := buildConfig(mode, allTools, initMsg.SystemInstruction)
	session, err := client.Live.Connect(ctx, geminiModel, config)
	if err != nil {
		log.Printf("live connect error: %v", err)
		return
	}
	defer session.Close()

	log.Printf("session started [mode=%s tools=%d servers=%d]", mode, len(allTools), len(servers))

	// Thread-safe write to browser — both the Gemini→browser goroutine and
	// dispatchToolCalls write concurrently, so they must share a mutex.
	var writeMu sync.Mutex
	writeBrowser := func(b []byte) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = clientConn.WriteMessage(websocket.TextMessage, b)
	}

	errCh := make(chan error, 2)

	// ── 4. Gemini → browser ───────────────────────────────────
	// ToolCall messages are intercepted here: Gemini asks for a function call
	// → we call MCP → we return the result and also notify the browser.
	// The browser never sees raw tool protocol, only audio/text plus
	// lightweight toolCall/toolResult status messages.
	go func() {
		for {
			msg, err := session.Receive()
			if err != nil {
				errCh <- fmt.Errorf("receive from gemini: %w", err)
				return
			}

			if msg.ToolCall != nil {
				if err := dispatchToolCalls(ctx, session, writeBrowser, toolRegistry, msg.ToolCall); err != nil {
					log.Printf("tool dispatch error: %v", err)
				}
				continue
			}

			b, err := json.Marshal(msg)
			if err != nil {
				errCh <- fmt.Errorf("marshal gemini message: %w", err)
				return
			}
			writeBrowser(b)
		}
	}()

	// ── 5. browser → Gemini ───────────────────────────────────
	go func() {
		for {
			_, raw, err := clientConn.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read from browser: %w", err)
				return
			}
			if err := forwardToGemini(session, mode, raw); err != nil {
				errCh <- err
				return
			}
		}
	}()

	if err := <-errCh; err != nil {
		log.Printf("session closed [mode=%s]: %v", mode, err)
	}
}

// dispatchToolCalls executes each Gemini function call via the MCP registry,
// notifies the browser about the call and result, then returns responses to Gemini.
func dispatchToolCalls(ctx context.Context, session *genai.Session, writeBrowser func([]byte), registry map[string]*MCPClient, toolCall *genai.LiveServerToolCall) error {
	var responses []*genai.FunctionResponse
	for _, fc := range toolCall.FunctionCalls {
		notify, _ := json.Marshal(map[string]any{"type": "toolCall", "tool": fc.Name, "args": fc.Args})
		writeBrowser(notify)

		var resp map[string]any

		// Built-in tools take lowest priority — MCP registry wins if the same name exists.
		if mcp, ok := registry[fc.Name]; ok {
			log.Printf("calling MCP tool: %s %v", fc.Name, fc.Args)
			text, err := mcp.CallTool(ctx, fc.Name, fc.Args)
			if err != nil {
				log.Printf("MCP tool %s error: %v", fc.Name, err)
				resp = map[string]any{"error": err.Error()}
				result, _ := json.Marshal(map[string]any{"type": "toolResult", "tool": fc.Name, "ok": false, "error": err.Error()})
				writeBrowser(result)
			} else {
				resp = map[string]any{"output": text}
				result, _ := json.Marshal(map[string]any{"type": "toolResult", "tool": fc.Name, "ok": true})
				writeBrowser(result)
			}
		} else if text, handled, builtinErr := callBuiltin(ctx, fc.Name, fc.Args); handled {
			log.Printf("calling built-in tool: %s", fc.Name)
			if builtinErr != nil {
				log.Printf("built-in tool %s error: %v", fc.Name, builtinErr)
				resp = map[string]any{"error": builtinErr.Error()}
				result, _ := json.Marshal(map[string]any{"type": "toolResult", "tool": fc.Name, "ok": false, "error": builtinErr.Error()})
				writeBrowser(result)
			} else {
				resp = map[string]any{"output": text}
				result, _ := json.Marshal(map[string]any{"type": "toolResult", "tool": fc.Name, "ok": true})
				writeBrowser(result)
			}
		} else {
			log.Printf("unknown tool: %s", fc.Name)
			resp = map[string]any{"error": "tool not available"}
			result, _ := json.Marshal(map[string]any{"type": "toolResult", "tool": fc.Name, "ok": false, "error": "tool not available"})
			writeBrowser(result)
		}
		responses = append(responses, &genai.FunctionResponse{
			Name:     fc.Name,
			ID:       fc.ID,
			Response: resp,
		})
	}
	return session.SendToolResponse(genai.LiveToolResponseInput{
		FunctionResponses: responses,
	})
}

// buildConfig returns the LiveConnectConfig for the given mode.
// Both modes use ModalityAudio — this model does not support ModalityText.
// Text mode adds OutputAudioTranscription so the browser can display text.
func buildConfig(mode string, tools []*genai.FunctionDeclaration, systemInstruction string) *genai.LiveConnectConfig {
	var geminiTools []*genai.Tool
	if len(tools) > 0 {
		geminiTools = []*genai.Tool{{FunctionDeclarations: tools}}
	}
	var sysInst *genai.Content
	if systemInstruction != "" {
		sysInst = &genai.Content{Parts: []*genai.Part{{Text: systemInstruction}}}
	}
	if mode == "text" {
		return &genai.LiveConnectConfig{
			ResponseModalities:       []genai.Modality{genai.ModalityAudio},
			OutputAudioTranscription: &genai.AudioTranscriptionConfig{},
			Tools:                    geminiTools,
			SystemInstruction:        sysInst,
		}
	}
	return &genai.LiveConnectConfig{
		ResponseModalities:       []genai.Modality{genai.ModalityAudio},
		InputAudioTranscription:  &genai.AudioTranscriptionConfig{},
		OutputAudioTranscription: &genai.AudioTranscriptionConfig{},
		Tools:                    geminiTools,
		SystemInstruction:        sysInst,
	}
}

func forwardToGemini(session *genai.Session, mode string, raw []byte) error {
	if mode == "text" {
		var input genai.LiveClientContentInput
		if err := json.Unmarshal(raw, &input); err != nil {
			return fmt.Errorf("unmarshal text message: %w", err)
		}
		return session.SendClientContent(input)
	}
	var input genai.LiveRealtimeInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("unmarshal audio message: %w", err)
	}
	return session.SendRealtimeInput(input)
}

// notionTestHandler validates the Notion API key and lists the first few accessible pages.
// Hit http://localhost:8080/api/notion-test to debug token issues without Gemini.
func notionTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if notionAPIClient == nil {
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"ok":    false,
			"error": "NOTION_API_KEY not set in server/.env",
		})
		return
	}
	// /users/me confirms the token is valid and shows the integration name
	me, err := notionAPIClient.do(r.Context(), "GET", "/users/me", nil)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()}) //nolint:errcheck
		return
	}
	// Quick search to show accessible pages
	pages, _ := notionAPIClient.do(r.Context(), "POST", "/search", map[string]any{
		"page_size": 5,
		"filter":    map[string]any{"value": "page", "property": "object"},
	})
	var pageTitles []string
	if results, ok := pages["results"].([]any); ok {
		for _, r := range results {
			pageTitles = append(pageTitles, notionPageTitle(asMap(r)))
		}
	}
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"ok":              true,
		"integration":     me["name"],
		"accessible_pages": pageTitles,
	})
}

// defaultsHandler returns the server-preconfigured MCP server(s) (name + URL only,
// no auth) so the browser can pre-populate the settings panel without the user
// having to type anything. Auth tokens stay server-side.
func defaultsHandler(w http.ResponseWriter, r *http.Request) {
	type entry struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	var servers []entry
	if url := os.Getenv("MCP_URL"); url != "" {
		name := os.Getenv("MCP_NAME")
		if name == "" {
			name = "default"
		}
		servers = append(servers, entry{Name: name, URL: url})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"mcpServers": servers}) //nolint:errcheck
}

func proxyVideo(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://storage.googleapis.com/cloud-samples-data/video/animals.mp4")
	if err != nil {
		http.Error(w, "error fetching video", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(w, resp.Body) //nolint:errcheck — broken pipe on client disconnect is expected
}
