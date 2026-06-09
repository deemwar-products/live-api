# MCP Gateway

This directory contains the single MCP execution layer used by Gemini and the
local MCP dashboard.

## Architecture

```text
Gemini / HTTP dashboard
          |
          v
     MCP Gateway
          |
          +-- tenant context validation
          +-- tool registry lookup
          +-- role policy
          +-- per-org/tool rate limit
          +-- tool execution
          +-- audit recording
          |
          v
 RAG / session / presence / notification contracts
```

`Gateway.Execute` is the only local tool execution entry point. Every request
must include an `MCPContext` with an `OrgID`.

Execution order:

1. Validate tenant context.
2. Find the registered tool.
3. Apply the role policy.
4. apply the per-organization tool quota.
5. Execute the tool.
6. Record success, failure, identity, and duration in the audit sink.

## Core Tools

Exactly eight coarse tools are exposed to Gemini:

| Tool | Purpose | Required arguments |
|---|---|---|
| `retrieve_knowledge` | Search organization knowledge through RAG | `query` |
| `create_ticket` | Create an unresolved-support ticket | `title`, `description` |
| `send_notification` | Send email, SMS, WhatsApp, or in-app messages | `channel`, `recipient`, `message` |
| `create_escalation` | Request human-agent intervention | `reason` |
| `flag_knowledge_gap` | Record missing organization knowledge | `topic` |
| `execute_workflow` | Run an approved multi-step workflow | `workflow`, `summary`, `reason` |
| `query_connector` | Route an operation to an organization connector | `connector`, `operation` |
| `get_customer_context` | Load tenant-scoped session/customer context | None |

Provider-specific and low-level operations stay behind these tools in Go.
Tool definitions and implementations live in `coretools/core_tools.go`.

## Package Guide

| Path | Responsibility |
|---|---|
| `gateway.go` | Tool interface, registry, and execution pipeline |
| `context.go` | Tenant, user, session, and role context |
| `policy/` | Gemini-facing tool allowlist |
| `rate_limit.go` | In-memory quota per organization and tool |
| `audit.go` | No-op and bounded in-memory audit sinks |
| `coretools/` | Eight coarse tool implementations and gateway wiring |
| `contracts/` | RAG, session, presence, and notification interfaces |
| `registry/` | Use-case-based tool discovery |
| `remote/` | Client for external MCP command servers |
| `workflow/` | Sequential multi-tool workflows |
| `memory/` | Short-term conversation store |
| `analytics/` | Tool metrics and conversation scoring |

## Production Wiring

Implement the service contracts, then construct one gateway:

```go
audit := mcp.NewBoundedAuditSink(500)
limiter := mcp.NewRateLimiter(1000)

gateway, err := coretools.NewCoreGateway(
    ragService,
    sessionStore,
    presenceService,
    notificationService,
    connectorRouter,
    audit,
    limiter,
    policy.MVPPolicy(),
)
if err != nil {
    return err
}
```

Execute a tool with tenant context:

```go
ctx := mcp.WithMCPContext(ctx, mcp.MCPContext{
    OrgID:     "org-123",
    UserID:    "user-456",
    SessionID: "session-789",
    Role:      "ai",
})

result, err := gateway.Execute(ctx, "create_ticket", map[string]any{
    "title":       "Payment failed",
    "description": "Customer could not complete checkout",
    "priority":    "high",
})
```

Do not accept `OrgID` from tool arguments. Populate it from authenticated
middleware and inject it through `WithMCPContext`.

## Local Dashboard

From the repository root:

```bash
go run ./cmd/mcp-ui
```

Open `http://localhost:8080`.

Available endpoints:

| Endpoint | Purpose |
|---|---|
| `GET /api/health` | Gateway health and registered-tool count |
| `GET /api/tools` | Tool names, descriptions, and input schemas |
| `POST /api/execute` | Execute a registered tool |
| `GET /api/audit` | View bounded in-memory execution records |

Example:

```bash
curl -X POST http://localhost:8080/api/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "tool_name": "create_ticket",
    "org_id": "demo-org",
    "session_id": "demo-session",
    "role": "ai",
    "arguments": {
      "title": "Password reset",
      "description": "Customer cannot access the account"
    }
  }'
```

The dashboard wires nil external services. `create_ticket`,
`create_escalation`, `flag_knowledge_gap`, and `execute_workflow` can
demonstrate local behavior.
`retrieve_knowledge` requires a RAG implementation. Notification dispatch
requires a notification implementation; without one, the tool validates and
returns a local success response but sends nothing.
`query_connector` and `get_customer_context` require their corresponding
production services.

## Verification

Run commands from the repository root and use `./...` with three dots:

```bash
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
```


Tests cover every package, including tenant propagation, policy enforcement,
quota isolation, audit behavior, tool validation, HTTP handlers, workflows,
memory concurrency, Gemini declarations, and remote-client failure handling.

## Current Boundaries

- Quotas and audit records are in memory and reset on process restart.
- Service contracts must be backed by production infrastructure.
- Remote MCP integration uses command transport.
- Authentication middleware is outside this package.
- Tool schemas are declarations; full JSON Schema validation is not yet part
  of `Gateway.Execute`.
