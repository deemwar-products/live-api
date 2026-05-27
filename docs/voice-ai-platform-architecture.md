---
title: Voice AI Customer Support Platform — Solution Architecture
status: draft
created: 2026-05-24
updated: 2026-05-24
author: Sreyash Reddy
source_prd: prd-live-api-poc-2026-05-24/prd.md
---

# Voice AI Customer Support Platform — Solution Architecture

## Executive Summary

This document provides the solution architecture for the Voice AI Customer Support Platform PoC. It addresses the architectural concerns raised during PRD review and provides concrete decisions for the technical implementation.

---

## 1. Architecture Overview

### 1.1 High-Level System Topology

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                          │
│  │  Web App    │  │  Mobile SDK │  │  Embed SDK  │                          │
│  │  (Voice+    │  │  (Voice+    │  │  (Embedded  │                          │
│  │   Text)     │  │   Text)     │  │   Widget)   │                          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                          │
└─────────┼────────────────┼────────────────┼─────────────────────────────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            API GATEWAY                                      │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │  Rate Limiting │ Auth │ Tenant Context │ Request Routing │ Logging   │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          CORE SERVICES                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐    │
│  │ Voice       │  │ RAG         │  │ MCP         │  │ Analytics       │    │
│  │ Service     │  │ Service     │  │ Service     │  │ Service         │    │
│  │ (Gemini     │  │ (Retrieval  │  │ (Tool       │  │ (Scoring,       │    │
│  │  Live)      │  │  Pipeline)  │  │  Execution) │  │  Aggregation)   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
          │                │                │                │
          ▼                ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DATA LAYER                                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐    │
│  │ Vector DB   │  │ Document    │  │ Session     │  │ Analytics       │    │
│  │ (Per-Org    │  │ Store       │  │ Store       │  │ Store           │    │
│  │  Namespace) │  │ (S3 Per-Org)│  │ (Redis)     │  │ (TimescaleDB)   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        EXTERNAL INTEGRATIONS                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐    │
│  │ Gemini     │  │ Org MCP     │  │ Human       │  │ Notification    │    │
│  │ Live API   │  │ Servers     │  │ Agent Sys   │  │ (Email/Slack)   │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Core Principles

| Principle | Application |
|-----------|-------------|
| **Multi-Tenant Isolation** | Every service enforces tenant context; no shared state across orgs |
| **Real-Time First** | Voice pipeline optimized for sub-second latency |
| **Fault Tolerance** | Graceful degradation on every failure path |
| **Observability** | Full tracing from customer to response |

---

## 2. Voice Pipeline Architecture

### 2.1 Audio Flow

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Browser    │───▶│   WebRTC     │───▶│   Voice      │───▶│   Gemini     │
│   Microphone │    │   Media      │    │   Gateway    │    │   Live API   │
│              │◀───│   Stream     │◀───│   Service    │◀───│              │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
                                                        │
                                                        ▼
                                                  ┌──────────────┐
                                                  │   Response   │
                                                  │   Audio      │
                                                  │   Stream     │
                                                  └──────────────┘
```

### 2.2 Audio Buffering Strategy

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| **Chunk Duration** | 100ms | Balance between latency and context; 100ms is standard for real-time voice |
| **Buffer Size** | 3 chunks (300ms) | Ensures smooth audio even with minor network jitter |
| **Overlap** | 50ms | Prevents word boundary clipping during streaming |

### 2.3 Backchannel Handling (Customer Interruption)

```
Customer Speaks ──▶ AI Processing ──▶ AI Response Starts
                                              │
                   Customer Interrupts ◀──────┘
                   │
                   ▼
              ┌─────────────────┐
              │ Interruption    │
              │ Detection       │
              │ (VAD + Context) │
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │ Cancel Pending  │  OR  │ Continue (if priority is low) │
              │ Audio Response  │      └───────────────────────────────┘
              └─────────────────┘
```

**Interruption Detection:**
- Voice Activity Detection (VAD) triggers on sound above -40dB threshold
- If customer sound > AI sound for >200ms, trigger interruption
- Preserve conversation context up to interruption point

**Response Cancellation:**
- Stop audio stream immediately on interruption
- Do NOT send partial/incorrect response
- Wait for customer's new query, then continue

### 2.4 WebRTC Reconnection Strategy

| Event | Behavior |
|-------|----------|
| **Connection Drop < 5s** | Auto-reconnect; preserve session context |
| **Connection Drop 5-30s** | Show "reconnecting" UI; attempt reconnect 3x |
| **Connection Drop > 30s** | End session gracefully; flag as incomplete |
| **Connection Drop > 2min** | Mark conversation for review; notify admin |

**Session Persistence:**
- Session state stored in Redis with 30-minute TTL
- Browser refresh preserves session via session_id cookie
- Infrastructure restart does NOT lose active sessions

### 2.5 Session State Management

```
┌─────────────────────────────────────────────────────────────────┐
│                      SESSION STORE (Redis)                       │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ Session:{session_id}                                       │ │
│  │   org_id: string                                           │ │
│  │   customer_id: string                                      │ │
│  │   conversation_history: [{role, content, timestamp}]        │ │
│  │   context_chunks: [embedded_doc_ids]                       │ │
│  │   mcp_tool_state: {}                                       │ │
│  │   started_at: timestamp                                    │ │
│  │   last_activity: timestamp                                 │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**TTL Policy:** Sessions expire 30 minutes after last activity.

---

## 3. Multi-Tenant Isolation Architecture

### 3.1 Isolation Strategy: Soft Multi-Tenancy with Logical Enforcement

**Decision: Soft multi-tenancy with strict logical enforcement**

Rationale: Hard multi-tenancy (separate DB instances per org) adds operational complexity without proportional security benefit for this use case. The risk of logical filtering bugs is acceptable with proper controls.

### 3.2 Isolation Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLIENT REQUEST                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   API GATEWAY                                    │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │ Tenant Context Middleware                                 │   │
│  │   - Extract org_id from JWT token                        │   │
│  │   - Inject org_id into request context                   │   │
│  │   - Validate org_id matches user's organization          │   │
│  └───────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   SERVICE LAYER                                  │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │ Tenant Context Validation                                │   │
│  │   - Every service validates org_id on every request      │   │
│  │   - No service accepts cross-org data references         │   │
│  │   - All queries include org_id filter                    │   │
│  └───────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   Vector DB      │  │   Document      │  │   Analytics     │
│   (Namespaced)   │  │   Store (S3)    │  │   Store         │
│                 │  │                 │  │                 │
│ namespace:      │  │ prefix:         │  │ tenant_id on    │
│ org_id          │  │ orgs/{org_id}/  │  │ every record    │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### 3.3 Threat Model & Mitigation

| Threat | Mitigation |
|--------|------------|
| **JWT forged with different org_id** | Validate JWT signature; verify org_id claim matches authenticated user |
| **SQL/NoSQL injection to access cross-org** | All queries parameterized; org_id always enforced at query level |
| **Vector DB namespace collision** | Namespace enforced at SDK level, not query level |
| **S3 prefix traversal** | S3 keys constructed server-side only; no user input in path |
| **MCP credential cross-org access** | MCP configs scoped by org_id; credential store enforces scope |

### 3.4 Verification Strategy

| Verification | Method |
|--------------|--------|
| **Row-level security** | Database-level RLS policies on all tenant-scoped tables |
| **Unit tests** | Test that queries without org_id fail; test cross-org access attempts return 403 |
| **Integration tests** | Test with multiple orgs in same DB; verify no data leakage |
| **Security audit** | Quarterly review of all data access patterns |

---

## 4. RAG Pipeline Architecture

### 4.1 Document Ingestion Pipeline

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Document   │───▶│  Validation  │───▶│  Chunking     │───▶│  Embedding    │
│   Upload     │    │  (Type,Size, │    │  (Semantic   │    │  (Batch,      │
│   (Org-Scoped)│   │   Virus Scan) │    │   Overlap)    │    │   Async)      │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
                                                                │
                                                                ▼
                                                          ┌──────────────┐
                                                          │  Vector DB   │
                                                          │  (Indexed)   │
                                                          └──────────────┘
```

### 4.2 Chunking Strategy

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| **Chunk Size** | 512 tokens | Optimal for semantic search; preserves context within chunk |
| **Chunk Overlap** | 64 tokens | Ensures concepts at boundaries aren't lost |
| **Chunking Method** | Semantic (sentence-aware) | Prefer splitting on sentence boundaries over arbitrary cuts |
| **Min Chunk Size** | 100 tokens | Discard chunks below threshold (likely incomplete) |
| **Max Chunk Size** | 1024 tokens | Split larger chunks to prevent context overflow |

### 4.3 Embedding Pipeline

| Parameter | Value |
|-----------|-------|
| **Model** | TBD (recommend text-embedding-3-small or Vertex AI) |
| **Batch Size** | 100 chunks per API call |
| **Processing** | Async (background job queue) |
| **SLA** | Documents searchable within 5 minutes of upload |
| **Update Behavior** | On document update: delete old embeddings, create new |
| **Delete Behavior** | On document delete: cascade delete all embeddings |

### 4.4 Retrieval Strategy

```
Query ──▶ Intent Classification ──▶ Query Embedding
                                        │
                                        ▼
                              ┌─────────────────────┐
                              │ Vector Similarity  │
                              │ Search (top-20)    │
                              └──────────┬──────────┘
                                         │
                          ┌──────────────┴──────────────┐
                          ▼                              ▼
                    ┌───────────┐                ┌───────────┐
                    │ Relevance │                │ Relevance │
                    │ > 0.75   │                │ < 0.75    │
                    │ Use      │                │ Discard   │
                    └───────────┘                └───────────┘
                          │
                          ▼
                  ┌───────────────────┐
                  │ Top-K Chunks      │
                  │ (Max 10 chunks)   │
                  │ + Context Header  │
                  └───────────────────┘
                          │
                          ▼
                  ┌───────────────────┐
                  │ LLM Prompt        │
                  │ (System + Context  │
                  │  + Query)          │
                  └───────────────────┘
```

### 4.5 Relevance Gate

| Score Range | Action |
|-------------|--------|
| **> 0.75** | Include in context; confident response |
| **0.5 - 0.75** | Include but flag uncertainty |
| **< 0.5** | Exclude from context; do NOT answer with this data |

**"No Relevant Context" Path:**
- If all retrieved chunks < 0.5, agent should NOT fabricate answer
- Trigger escalation or "I don't have information about that" response

---

## 5. MCP Integration Architecture

### 5.1 Tool Calling Strategy

```
Query ──▶ Intent Classification ──▶ Decision
                                    │
            ┌───────────────────────┼───────────────────────┐
            ▼                       ▼                       ▼
     ┌──────────┐            ┌──────────┐            ┌──────────┐
     │ RAG Only │            │ MCP Only │            │ RAG+MCP  │
     │ Needed   │            │ Needed   │            │ Needed   │
     └──────────┘            └──────────┘            └──────────┘
            │                       │                       │
            ▼                       ▼                       ▼
     Use Vector DB           Call MCP Tool           Parallel Fetch
     for context             for data               Both → Merge
```

**Tool Selection Criteria:**
- Entity detected: "order status", "account balance", "appointment" → MCP
- Comparison/analysis: "what's the difference", "compare" → RAG
- Both possible: "what did the agent say about X and what's my status" → Both

### 5.2 MCP Request/Response Handling

```
┌─────────────────────────────────────────────────────────────────┐
│                     MCP SERVICE                                  │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │ MCP Config  │───▶│  Tool       │───▶│  Response   │        │
│  │ (Per-Org)   │    │  Executor   │    │  Parser     │        │
│  │             │    │             │    │             │        │
│  │ - URL       │    │ - Auth      │    │ - Schema    │        │
│  │ - Auth      │    │ - Timeout   │    │   Validation│        │
│  │ - Schema    │    │ - Retry     │    │ - Error     │        │
│  │ - Rate Limit│    │ - Sandbox   │    │   Handling  │        │
│  └─────────────┘    └─────────────┘    └─────────────┘        │
└─────────────────────────────────────────────────────────────────┘
```

### 5.3 Timeout & Retry Configuration

| Scenario | Timeout | Retry Strategy |
|----------|---------|----------------|
| **MCP Request** | 3 seconds | 2 retries with exponential backoff |
| **MCP Response > 1MB** | Reject | Log warning |
| **MCP Unavailable** | Fail fast | Return error to agent; do not retry infinitely |
| **MCP Partial Failure** | N/A | If partial data received, use available data |

### 5.4 Credential Management

| Aspect | Implementation |
|--------|----------------|
| **Storage** | Encrypted (AES-256) in secrets manager per org_id |
| **Refresh** | OAuth tokens refreshed automatically; service account keys rotated |
| **Access** | MCP config only accessible to org admin; platform never reads raw credentials |
| **Audit** | All MCP calls logged with org_id, timestamp, tool called |

### 5.5 Security Sandbox (SSRF Prevention)

| Control | Implementation |
|---------|----------------|
| **IP Allowlist** | Only configured URLs callable; no arbitrary destinations |
| **DNS Rebinding Protection** | Resolve DNS once; verify IP against allowlist |
| **Request Size Limits** | Max request/response size enforced |
| **Network Isolation** | MCP calls from isolated network segment; no access to internal services |
| **Method Restrictions** | Only GET/POST allowed; no DELETE/PUT unless explicitly configured |

---

## 6. Escalation Architecture

### 6.1 Escalation Decision Logic

```
Query Received
      │
      ▼
┌─────────────────┐
│ Intent          │
│ Classification  │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐  ┌────────┐
│ RAG    │  │ MCP    │
│ Needed │  │ Needed │
└───┬────┘  └───┬────┘
    │           │
    ▼           ▼
┌───────────────┐     ┌─────────────────┐
│ RAG Retrieval │     │ MCP Call        │
└───────┬───────┘     └────────┬────────┘
        │                      │
        ▼                      ▼
┌───────────────────────┐  ┌───────────────────────┐
│ Relevance Check       │  │ Response Validation  │
│ (All chunks > 0.5?)   │  │ (Valid data returned?)│
└───────────┬───────────┘  └───────────┬───────────┘
            │                          │
            ▼                          ▼
     ┌──────────────┐            ┌──────────────┐
     │ Yes: Proceed │            │ No: Escalate │
     └──────────────┘            └──────────────┘
            │
            ▼
┌───────────────────────┐
│ Confidence Check      │
│ (LLM confidence > 0.8?)│
└───────────┬───────────┘
            │
     ┌──────┴──────┐
     ▼             ▼
┌────────┐    ┌────────────┐
│  Yes   │    │ No:       │
│ Respond│    │ Escalate  │
└────────┘    └────────────┘
```

**Escalation Triggers (Agent-Centric):**
1. Agent **decides** it cannot answer — agent has agency to trigger escalation tool (not hardcoded binary)
2. MCP call fails or returns invalid data
3. Customer explicitly requests human ("talk to agent")

**Note:** This is NOT a hardcoded binary check. The agent uses its reasoning to determine when to escalate, making it more flexible and intelligent.

### 6.2 Handoff Mechanism

```
Escalation Triggered
        │
        ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ESCALATION HANDLER                           │
│                                                                 │
│  1. Preserve conversation state (full transcript + context)     │
│  2. Determine escalation target:                                │
│     - If org has human agent system → webhook/ticket            │
│     - If org has live chat → route to chat queue                │
│     - If org has phone support → generate callback request      │
│  3. Notify admin (in-app + optional email/Slack)                │
│  4. Show customer escalation UI                                 │
│  5. Store escalation record for analytics                       │
└─────────────────────────────────────────────────────────────────┘
```

### 6.3 Escalation UI (Customer Experience)

| Phase | Customer Sees |
|-------|---------------|
| **Escalation Initiated** | "I'm connecting you with a support specialist. Please hold." |
| **Waiting** | Animated "connecting..." indicator with estimated wait |
| **Agent Found** | "A specialist is on the way. Your conversation context has been shared." |
| **Transfer Complete** | Human agent joins (via chat or phone based on org config) |

### 6.4 Conversation Handoff Data

```json
{
  "escalation_id": "uuid",
  "org_id": "uuid",
  "customer_id": "string",
  "started_at": "ISO8601",
  "escalated_at": "ISO8601",
  "escalation_reason": "no_relevant_context|mcp_failure|low_confidence|explicit_request",
  "transcript": [
    {"role": "customer", "content": "...", "timestamp": "..."},
    {"role": "agent", "content": "...", "timestamp": "..."}
  ],
  "retrieved_chunks": ["doc_ids"],
  "mcp_tool_calls": [{"tool": "...", "response": "..."}],
  "customer_context": {
    "account_id": "...",
    "previous_interactions": 0
  }
}
```

---

## 7. Analytics Pipeline Architecture

### 7.1 Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      ANALYTICS PIPELINE                          │
│                                                                 │
│  Conversation ──▶ Event Stream ──▶ Real-time ──▶ Dashboard     │
│  Complete        (Kafka)        Processing    (Aggregated)    │
│       │                           │                             │
│       ▼                           ▼                             │
│  ┌─────────────┐           ┌─────────────┐                      │
│  │ Transcript  │           │  Scoring    │                      │
│  │ Storage     │           │  Engine     │                      │
│  └─────────────┘           └─────────────┘                      │
│                                    │                            │
│                                    ▼                            │
│                            ┌─────────────┐                      │
│                            │  Question   │                      │
│                            │  Tagger     │                      │
│                            └─────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

### 7.2 Satisfaction Scoring Model

**Two-Part System:**
1. **Sentiment Analysis** (Automatic) — Real-time voice tone analysis during the call
2. **Rating Agent** (Review-based) — Separate agent that reviews conversations and provides ratings

| Component | Type | Description |
|-----------|------|-------------|
| **Sentiment Analysis** | Automatic | Real-time voice tone analysis during the call |
| **Rating Agent** | Review-based | Separate agent reviews transcripts and rates conversations |

**Score Calculation:**
```
Final Score = sentiment_analysis_score + rating_agent_score
```

**Target Automation Rate:** 80-90% (feasible goal, not 100%)

**Threshold Configuration:**
- 85-100: Excellent
- 70-84: Good
- 50-69: Needs Attention
- 0-49: Poor

### 7.3 Question Tagging Taxonomy

| Category | Examples |
|----------|----------|
| **Account** | password reset, profile update, account access |
| **Billing** | payment, invoice, subscription, refund |
| **Product** | features, pricing, how-to, compatibility |
| **Order** | status, tracking, cancellation, return |
| **Technical** | error, bug, integration, API |
| **General** | greeting, feedback, complaint, compliment |
| **Unanswered** | questions with no response (flagged) |

**Tagging Mechanism:**
- Automatic: LLM classifies question into categories
- Manual: Admin can add/remove tags
- New category creation: Admin can create custom tags

### 7.4 Dashboard Metrics

| Metric | Calculation | Refresh |
|--------|-------------|---------|
| **Automation Rate** | (Total - Escalated) / Total * 100 | Real-time |
| **Avg Satisfaction** | Mean of scores for period | Daily |
| **Escalation Count** | Count where escalated = true | Real-time |
| **Top Unanswered Tags** | Group by tag, count unanswered | Daily |
| **Resolution Time** | Time from escalation to admin action | Daily |

---

## 8. Error Handling Architecture

### 8.1 Error Taxonomy

| Error Category | Examples | User-Facing Behavior |
|----------------|----------|----------------------|
| **Audio** | Mic permission denied, WebRTC failed | "Microphone access needed. Please check your browser settings." |
| **Network** | Connection lost, timeout | "Connection interrupted. Reconnecting..." (auto-retry) |
| **RAG** | No relevant context, vector DB unavailable | "I don't have information about that. Let me connect you with someone who can help." |
| **MCP** | Timeout, invalid response, server error | Proceed with RAG-only response; flag for admin |
| **LLM** | Gemini unavailable, rate limit | "I'm experiencing technical difficulties. Please try again in a moment." |
| **Auth** | Token expired, invalid org | "Session expired. Please refresh the page." |

### 8.2 Fallback Strategies

| Failure | Fallback Behavior |
|---------|------------------|
| **Gemini Live API unavailable** | Queue calls for retry; show "high demand" message |
| **Vector DB unavailable** | Return "service temporarily unavailable" |
| **MCP server unavailable** | Continue with RAG-only; log degraded mode |
| **All services unavailable** | "We're experiencing technical difficulties. Please try again later." |

### 8.3 Error Recovery UX

| Scenario | Customer UX |
|----------|-------------|
| **Audio input failed** | Visual prompt: "We didn't catch that. Please try again." |
| **Processing took too long** | "I'm looking into that for you..." (extended thinking indicator) |
| **Connection dropped** | "Your connection was interrupted. Reconnecting..." + auto-retry |
| **Services degraded** | "Some features may be slower than usual." |

---

## 9. API Contracts

### 9.1 Voice Session API

```
POST /api/v1/sessions
Create new voice session

Request:
{
  "org_id": "uuid",
  "customer_id": "string",
  "mode": "voice" | "text",
  "client_metadata": {...}
}

Response:
{
  "session_id": "uuid",
  "websocket_url": "wss://...",
  "token": "jwt"
}

---

GET /api/v1/sessions/{session_id}/status
Get session state

Response:
{
  "session_id": "uuid",
  "status": "active" | "escalated" | "completed" | "failed",
  "escalation_id": "uuid" | null,
  "started_at": "ISO8601",
  "last_activity": "ISO8601"
}
```

### 9.2 MCP Configuration API

```
POST /api/v1/orgs/{org_id}/mcp-servers
Add MCP server

Request:
{
  "name": "string",
  "url": "https://...",
  "auth_type": "api_key" | "oauth" | "none",
  "schema": {...},
  "rate_limit": 100
}

Response:
{
  "server_id": "uuid",
  "status": "pending" | "connected" | "error",
  "test_result": {...}
}

---

GET /api/v1/orgs/{org_id}/mcp-servers/{server_id}/tools
List available MCP tools
```

### 9.3 Analytics API

```
GET /api/v1/orgs/{org_id}/analytics/overview
Dashboard overview

Query params:
  - start_date: ISO8601
  - end_date: ISO8601

Response:
{
  "automation_rate": 0.85,
  "avg_satisfaction": 82.3,
  "total_calls": 1234,
  "escalated_count": 185,
  "top_tags": [...],
  "knowledge_gaps": [...]
}
```

---

## 10. Infrastructure Topology

### 10.1 Recommended Infrastructure

```
┌─────────────────────────────────────────────────────────────────┐
│                        AWS / GCP / Azure                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐        ┌─────────────┐        ┌─────────────┐ │
│  │   API       │        │   Voice     │        │   Worker    │ │
│  │   Cluster   │        │   Service   │        │   Cluster   │ │
│  │   (3+ nodes) │        │   (2+ nodes)│        │   (Auto-    │ │
│  │             │        │             │        │    scale)   │ │
│  └─────────────┘        └─────────────┘        └─────────────┘ │
│         │                     │                     │         │
│         └─────────────────────┼─────────────────────┘         │
│                               │                                 │
│                    ┌──────────┴──────────┐                     │
│                    │    Load Balancer     │                     │
│                    └──────────┬──────────┘                     │
└───────────────────────────────┼─────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                         DATA LAYER                               │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────┐  │
│  │  Redis      │  │  TimescaleDB│  │  Vector DB  │  │   S3    │  │
│  │  (Sessions)│  │  (Analytics)│  │  (Pinecone  │  │ (Docs)  │  │
│  │             │  │             │  │   / Weaviate│  │         │  │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 10.2 Scaling Targets

| Metric | Target |
|--------|--------|
| **Concurrent Voice Calls** | 100 per org, 1000 platform-wide |
| **Document Processing** | 1000 docs/day per org |
| **API Requests** | 10,000/min per org |
| **Vector DB Queries** | < 50ms P95 |

---

## 11. Open Architecture Decisions

| ID | Decision | Options | Recommendation |
|----|----------|---------|----------------|
| **AD-001** | Vector DB Provider | Pinecone / Weaviate / Chroma | Weaviate (open-source, hybrid search) |
| **AD-002** | Embedding Model | OpenAI text-embedding-3 / Vertex AI | text-embedding-3-small (cost-effective) |
| **AD-003** | WebRTC Provider | Twilio / Daily.co / Self-hosted | Daily.co (simpler integration) |
| **AD-004** | Analytics DB | TimescaleDB / ClickHouse | TimescaleDB (simpler, good for metrics) |
| **AD-005** | Orchestration | Kafka / Redis Streams / SQS | Redis Streams (simpler for our scale) |

---

## 12. Implementation Phases

### Phase 1: Core Voice + RAG (POC)

| Component | Implementation |
|-----------|---------------|
| Voice Pipeline | WebRTC + Gemini Live API + audio buffering |
| RAG Pipeline | Document upload → chunk → embed → store → retrieve |
| Session Management | Redis session store with context preservation |
| Escalation | Basic handoff (webhook) to org's external system |
| Analytics | Basic scoring + dashboard |
| Isolation | Tenant middleware + per-org namespace |

### Phase 2: MCP Integration

| Component | Implementation |
|-----------|---------------|
| MCP Config UI | Admin interface to add/config MCP servers |
| Tool Executor | Sandboxed MCP calls with timeout/retry |
| Data Merging | RAG + MCP data combined in prompt |

### Phase 3: Platform Enhancement

| Component | Implementation |
|-----------|---------------|
| Learning Loop | Capture admin resolutions → training data |
| Advanced Analytics | Trend analysis, predictive escalation |
| Multi-Channel | SMS, WhatsApp, chat integrations |

---

*Document Status: Draft — For technical implementation guidance*
*Next Steps: Review with team, finalize architecture decisions (AD-001 through AD-005), then proceed to implementation*