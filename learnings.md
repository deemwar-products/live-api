# Learnings — live-api-poc

## Session Start: 2026-05-24

### Gemini Live API
- Real-time voice interaction API from Google AI (Gemini)
- Supports streaming audio in/out for low-latency conversations
- Can be integrated with RAG pipelines for context-aware responses
- Configurable voice persona for tone/style

### MCP (Model Context Protocol)
- Anthropic's protocol for connecting AI models to external tools/servers
- Enables AI to call external APIs, databases, and tools in real-time
- Standardized way to extend AI capabilities beyond training data

### Key Architecture Insights (from party mode)
- **RAG latency for voice is critical** — sub-second required for good UX
- Pre-fetch and hybrid retrieval (BM25 + vector) helps
- MCP isolation must be architectural, not just logical
- Session state management is essential for voice continuity
- Escalation decision logic is the hardest problem

### Product Insights (from party mode)
- "Replace customer care completely" vs "agent assist" changes everything
- MCP scope clarification: data retrieval only, NOT action execution
- Human escalation is safety valve, not primary flow
- Admin win condition: maximize automation rate
- Feedback loop: Customer → Org Admin → Super Admin (no direct channel)

### Team Insights
- **Mary (BA)**: Needs outcome statements, not features. Trust/isolation is #1 enterprise concern.
- **Winston (Architect)**: RAG latency for voice is the kill zone. Pre-fetch + hybrid retrieval needed.
- **John (PM)**: Clarify who the user is. MCP without "closing the loop" is just a search engine.
- **Amelia (Dev)**: Pre-computed embeddings, query caching, rolling token windows.
- **Victor (Innovation)**: Blue ocean question — who do you put out of business?
- **Sally (UX)**: Voice is relational, not transactional. Agent persona matters.

### Document Structure (as per user request)
```
docs/
├── _architecture/     # Project-related docs (PRD, TRD, architecture, epics)
├── _learnings/        # LLM context-based learnings for agent consistency
└── learnings.md        # Session learnings
```

### Open Questions (RESOLVED from user clarification)
- **Escalation trigger**: Agent-centric — agent decides when it can't answer and triggers escalation tool itself. Not hardcoded binary, agent has agency.
- **Satisfaction scoring**: Two-part system:
  1. Sentiment analysis (automatic)
  2. Separate rating agent that reviews and rates conversations
- **Target automation rate**: 80-90% as feasible goal (not 100% — that's unrealistic)
- **Voice greeting**: Basic configuration, decide later

---

## Session: 2026-05-28

### Technical Research — Gemini Live API

**Architecture: Backend in the middle (not direct browser → Gemini)**
- Browser → Go Backend (WebSocket) → Gemini Live API
- Go backend orchestrates: system prompt, RAG context, MCP tools, session state
- Browser only handles audio in/out — all intelligence goes through backend
- Why: Safety, logging, PII filtering, compliance, rate limiting, enterprise requirements

**Go SDK**: `google.golang.org/genai`
- Official Google SDK, has full Live API support
- Model: `gemini-live-2.5-flash-preview`
- API Version: v1alpha (required for Live API)

**Audio Specs**:
- Input (Mic → Backend): PCM 16kHz, 16-bit, mono, little-endian
- Output (Backend → Speaker): PCM 24kHz, 16-bit, mono, little-endian
- Chunk size: ~32ms (512 samples recommended)
- Browser does resampling: 16kHz input, 24kHz output

**WebSocket Protocol**:
- All messages are JSON (text frames) — no binary control messages
- Client sends: setup, realtimeInput (audio), toolResponse
- Server sends: setupComplete, serverContent (text/audio), toolCall, turnComplete
- Gemini Live API endpoint: `wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateContentConstrained`

### Stack Decisions (Locked)
| Component | Decision |
|-----------|----------|
| Backend | Go |
| Primary DB | PostgreSQL |
| Session/Cache | Redis |
| Document Storage | S3-compatible |
| Voice AI | Gemini Live API (Go SDK) |
| Tool Calling | MCP |
| Knowledge Retrieval | RAG |
| Analytics | DuckDB (Phase 2) — PostgreSQL for MVP |
| Multi-tenancy | org_id based (simple) |
| Deployment | Kamal (discuss later) |

---

## Session: 2026-05-28 (Afternoon)

### Additional Clarifications

**Token/Credit Management:**
- Primary API key for platform usage
- Secondary API key as fallback ("dead case" — 1 day survival)
- Tracking: per platform level, but cost calculated per org internally
- Credit usage shown in dashboard per org
- Notification: credits < 20% → Super Admin (warning), < 5% → urgent alert

**Notification System:**
- MCP down → Org Admin (in-app + email)
- Escalation threshold → Org Admin (in-app + email)
- System health → Super Admin (in-app + email)
- In-app notification center in the app
- Teams integration
- Webhook for external handling (user handles)

**Role Hierarchy:**
```
PLATFORM LEVEL (Super Admin)
├── Write All (fixed role)
└── Read All (fixed role)

ORGANIZATION LEVEL
├── Admin (fixed, creates custom roles)
├── Custom Role (Admin creates with granular permissions)
├── Agent (human support, takes escalated calls)
└── Customer (end user, no dashboard access)
```

**Permissions:**
- Granular toggle system (like GitHub)
- Admin assigns specific permissions per role/member
- No preset groups — each permission is independently toggled

**System Prompt + Greeting:**
- Base system prompt: provided by platform (org cannot edit)
- Greeting message: org can configure, gets appended to base
- Org cannot overwrite base system prompt

**Credit Tracking:**
- Per org internally calculated
- Shown in dashboard for each org
- Super Admin sees platform-wide usage + per-org breakdown

**Custom Role Creation:**
- Admin can create custom roles
- Admin can set granular permissions directly on the role
- Not based on template — free-form toggle per permission

**Dashboard Requirements:**
- Credit usage visible per org
- Notification center included
- Analytics, conversations, documents all visible based on role permissions

---

## Session: 2026-05-28 (Late Afternoon)

### Gemini Live API Rate Limits Research

**Quotas are 3-dimensional (simultaneously enforced):**
- RPM (Requests Per Minute)
- TPM (Tokens Per Minute)
- RPD (Requests Per Day)

**Important: Exact limits NOT publicly documented** — must check AI Studio dashboard for your specific project. Google reserves right to adjust.

**Concurrent session limits:** Not publicly documented. Community reports suggest 1-5 for standard tier. Enterprise (Vertex AI) can request increases.

**Tiered quota system (based on spend):**
| Tier | Spend | Priority |
|------|-------|----------|
| Free | $0 | Lowest |
| Tier 1 | Up to $250 | Higher than free |
| Tier 2 | Up to $2,000 | Higher |
| Tier 3 | $20K-$100K+ | Highest |

**Live API Connection:** WebSocket, ~55 second timeout. Must handle reconnection.

**Context windows:**
- Gemini 2.0 Flash: 128K (DEPRECATED — shutdown June 1, 2026)
- Gemini 2.5/3.5 Flash: 1M tokens
- Gemini 3.1 Pro: 2M tokens

**Cost (Live API):**
- Audio input: $3.00/1M tokens OR $0.005/min
- Audio output: $12.00/1M tokens OR $0.018/min

### Rate Limiting Strategy for Multi-Tenant Platform

**Per-org rate limiting** (to protect shared API key):
- Token bucket or sliding window per org
- Redis-based for speed
- When org hits limit → queue or fallback

**Handling 429 (rate limit) responses:**
1. Parse `Retry-After` header
2. If present → wait that duration
3. If absent → exponential backoff (starting 5s)
4. Log with org_id, timestamp, retry count
5. If persists > 60s → alert ops team

**Fallback chain:**
```
Primary API key
    ↓ (fails/rate limited)
Secondary API key (1 day survival)
    ↓ (also fails)
Queue for retry
    ↓ (fails)
Escalation to human
```

### Failure Handling Strategy

**Degradation levels:**
| Level | Condition | User Experience |
|-------|-----------|------------------|
| Full | All services available | Normal |
| Degraded | RAG down | "I can help with general questions..." |
| Degraded | MCP down | Continue with RAG only |
| Critical | Gemini down | "Technical difficulties" |
| Outage | All down | Escalate immediately |

**Session recovery:**
- Drop < 5s: Auto-reconnect, preserve context
- Drop 5-30s: Show UI, retry 3x
- Drop > 30s: End gracefully, flag incomplete
- Drop > 2min: Mark for review, notify admin

**Circuit breaker per service:**
- Failure threshold: 5 failures in 30s
- Open: Return fallback immediately
- Half-open: Allow 1 test request
- Close if test succeeds

### Credit Management

- Primary API key: day-to-day
- Secondary API key: dead case fallback (1 day survival)
- Track per org internally
- Show in dashboard
- Alert at 20% remaining, urgent at 5%

### Gaps to Address in TRD

1. **Redis session recovery** — how to resume Gemini Live session after disconnect
2. **Audio chunk retry** — buffer and retry vs discard on failure
3. **Dead letter queue** — retry interval and max retries
4. **Health check endpoints** — intervals and "unhealthy" definition per service
5. **SLA for degradation** — how long can platform stay degraded

---

## Session: 2026-05-28 (Evening)

### TRD Created

- File: `docs/trd.md`
- 18 sections drafted
- Needs review and refinement with user

### TRD Sections Drafted
1. Overview
2. System Architecture (with topology diagram)
3. API Design (REST + WebSocket, message formats)
4. Data Models (PostgreSQL schemas, Redis structures, S3 paths)
5. Voice Pipeline (audio specs, context management, session resumption)
6. RAG Pipeline (chunking, retrieval, relevance gate)
7. MCP Integration (tool calling flow, timeout/retry)
8. Session Management (lifecycle, TTL)
9. Multi-Tenant Isolation (enforcement layers)
10. Authentication & Authorization (JWT structure)
11. Role Management (hierarchy: platform + org level)
12. Permission Matrix (granular toggles)
13. Error Handling & Resilience (retry, circuit breaker, degradation)
14. Credit & Billing Management (API key pool, monitoring, failover)
15. Notification System (types, channels, preferences)
16. Analytics & Logging
17. Infrastructure & Deployment (placeholder)
18. Non-Functional Requirements (targets)

### TRD Section Proposal (18 sections)
1. Overview
2. System Architecture
3. API Design (REST + WebSocket)
4. Data Models (DB schemas, Redis, S3)
5. Voice Pipeline (Gemini integration)
6. RAG Pipeline (chunking, embeddings, retrieval)
7. MCP Integration (tool calling, timeout, retry)
8. Session Management (Redis state)
9. Multi-Tenant Isolation (org_id enforcement)
10. Authentication & Authorization
11. Role Management (platform + org hierarchy)
12. Permission Matrix (granular access control)
13. Error Handling & Resilience
14. Credit & Billing Management (Gemini usage tracking)
15. Notification System (alerts, who gets what)
16. Analytics & Logging
17. Infrastructure & Deployment (Kamal - later)
18. Non-Functional Requirements