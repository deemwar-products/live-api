package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/mux"

	"live-api/internal/mcp"
	"live-api/internal/mcp/coretools"
	"live-api/internal/mcp/policy"
)

type app struct {
	gateway *mcp.Gateway
	audit   *mcp.InMemoryAuditSink
}

type executeRequest struct {
	ToolName  string         `json:"tool_name"`
	Arguments map[string]any `json:"arguments"`
	OrgID     string         `json:"org_id"`
	UserID    string         `json:"user_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Role      string         `json:"role,omitempty"`
}

func main() {
	audit := mcp.NewBoundedAuditSink(500)
	gateway, err := coretools.NewCoreGateway(
		nil,
		nil,
		nil,
		nil,
		nil,
		audit,
		mcp.NewRateLimiter(1000),
		policy.MVPPolicy(),
	)
	if err != nil {
		log.Fatalf("initialize MCP gateway: %v", err)
	}

	server := &app{gateway: gateway, audit: audit}
	router := mux.NewRouter()
	router.HandleFunc("/", server.home).Methods(http.MethodGet)
	router.HandleFunc("/api/tools", server.listTools).Methods(http.MethodGet)
	router.HandleFunc("/api/execute", server.executeTool).Methods(http.MethodPost)
	router.HandleFunc("/api/audit", server.listAudit).Methods(http.MethodGet)
	router.HandleFunc("/api/health", server.health).Methods(http.MethodGet)

	addr := envOr("MCP_UI_ADDR", ":8080")
	log.Printf("MCP dashboard listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("serve MCP dashboard: %v", err)
	}
}

func (a *app) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (a *app) listTools(w http.ResponseWriter, _ *http.Request) {
	tools := a.gateway.Registry.List()
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name() < tools[j].Name() })
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		out = append(out, map[string]any{
			"name":         tool.Name(),
			"description":  tool.Description(),
			"input_schema": tool.InputSchema(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": out, "count": len(out)})
}

func (a *app) executeTool(w http.ResponseWriter, r *http.Request) {
	var req executeRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	req.ToolName = strings.TrimSpace(req.ToolName)
	req.OrgID = strings.TrimSpace(req.OrgID)
	if req.ToolName == "" || req.OrgID == "" {
		writeError(w, http.StatusBadRequest, "tool_name and org_id are required", nil)
		return
	}
	if req.Arguments == nil {
		req.Arguments = map[string]any{}
	}

	ctx := mcp.WithMCPContext(r.Context(), mcp.MCPContext{
		OrgID: req.OrgID, UserID: req.UserID, SessionID: req.SessionID, Role: req.Role,
	})
	result, err := a.gateway.Execute(ctx, req.ToolName, req.Arguments)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "quota exceeded") {
			status = http.StatusTooManyRequests
		} else if strings.Contains(err.Error(), "cannot use tool") {
			status = http.StatusForbidden
		}
		writeError(w, status, "tool execution failed", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tool": req.ToolName, "result": result})
}

func (a *app) listAudit(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"executions": a.audit.List()})
}

func (a *app) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":           "healthy",
		"tools_registered": a.gateway.Registry.Count(),
		"architecture":     "mcp_gateway",
	})
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	body := map[string]any{"error": message}
	if err != nil {
		body["detail"] = err.Error()
	}
	writeJSON(w, status, body)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil && !errors.Is(err, http.ErrHandlerTimeout) {
		log.Printf("encode response: %v", err)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>MCP Gateway</title>
<style>
*{box-sizing:border-box}body{margin:0;font:14px system-ui,sans-serif;background:#f4f6f8;color:#17212b}
header{padding:18px 24px;background:#17212b;color:#fff}header h1{margin:0;font-size:20px}
main{display:grid;grid-template-columns:300px 1fr;min-height:calc(100vh - 60px)}
aside{padding:18px;border-right:1px solid #d8dee4;background:#fff}.tool{width:100%;padding:10px;text-align:left;background:#fff;border:1px solid #d8dee4;border-radius:6px;margin-bottom:8px;cursor:pointer}.tool:hover,.tool.active{border-color:#087f5b;background:#edf9f5}
section{padding:24px;max-width:900px}label{display:block;font-weight:600;margin:14px 0 6px}input,textarea{width:100%;padding:10px;border:1px solid #b8c2cc;border-radius:6px;font:inherit}textarea{min-height:220px;font-family:ui-monospace,monospace}
button.run{margin-top:14px;padding:10px 16px;border:0;border-radius:6px;background:#087f5b;color:#fff;font-weight:700;cursor:pointer}pre{padding:16px;background:#17212b;color:#d8f3e8;border-radius:6px;overflow:auto;min-height:100px}
.muted{color:#66717d;font-size:13px}@media(max-width:700px){main{grid-template-columns:1fr}aside{border-right:0;border-bottom:1px solid #d8dee4}}
</style>
</head>
<body>
<header><h1>MCP Gateway</h1></header>
<main>
<aside><strong>Registered tools</strong><div id="tools" style="margin-top:12px"></div></aside>
<section>
<h2 id="name">Select a tool</h2><p id="description" class="muted"></p>
<label for="org">Organization ID</label><input id="org" value="demo-org">
<label for="args">Arguments (JSON)</label><textarea id="args">{}</textarea>
<button class="run" id="run">Run tool</button>
<label>Result</label><pre id="result">Ready.</pre>
</section>
</main>
<script>
let selected="";
fetch("/api/tools").then(r=>r.json()).then(data=>{
 const root=document.getElementById("tools");
 data.tools.forEach(tool=>{
  const b=document.createElement("button");b.className="tool";b.textContent=tool.name;
  b.onclick=()=>{selected=tool.name;document.querySelectorAll(".tool").forEach(x=>x.classList.remove("active"));b.classList.add("active");document.getElementById("name").textContent=tool.name;document.getElementById("description").textContent=tool.description;const sample={};for(const key of (tool.input_schema.required||[]))sample[key]="";document.getElementById("args").value=JSON.stringify(sample,null,2)};root.appendChild(b);
 });
});
document.getElementById("run").onclick=async()=>{
 const out=document.getElementById("result");
 if(!selected){out.textContent="Select a tool first.";return}
 try{
  const body={tool_name:selected,org_id:document.getElementById("org").value,role:"ai",arguments:JSON.parse(document.getElementById("args").value)};
  const response=await fetch("/api/execute",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(body)});
  out.textContent=JSON.stringify(await response.json(),null,2);
 }catch(error){out.textContent=error.message}
};
</script>
</body>
</html>`
