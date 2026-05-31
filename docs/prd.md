# Voice AI Customer Support Platform - PRD

## 1. Overview

**What:** AI-powered voice and chat customer support platform for businesses.

**Why:** Businesses need 24/7 customer support that handles 80-90% of queries autonomously, reducing human workload while improving response times.

**Core Loop:**
1. Business uploads knowledge base (documents, FAQs, integrations)
2. Customers interact via voice or chat on business's subdomain or embedded widget
3. AI responds using RAG + MCP tools + Gemini Live API
4. Escalation to human agent when needed (configurable toggle)
5. LLM Judge evaluates conversations and feeds back knowledge gaps

**Target:** 80-90% automation rate

---

## 2. Three Interfaces

| Interface | Who | Access |
|-----------|-----|--------|
| **Platform Dashboard** | Super Admin (platform team) | All orgs, system settings, internal alerts |
| **Business Dashboard** | Org Admins + Teams | Own docs, team management, AI config, analytics |
| **Customer Web App** | End customers | Chat/Voice interaction |

### Customer Access Points
- Subdomain per org: `<Live-api>{org-slug}.platform.com`
- Embeddable widget for company's website
- API exposed for integration with existing chatbots
- Voice + Chat available from launch

---

## 3. Roles & Permissions

### 3.0 Role Hierarchy (Tree Structure)

```
Platform (Super Admin)
├── Super Admin (Full Access)
│   ├── Read-write Admin
│   │   ├── All write permissions
│   │   └── Platform configuration
│   └── Read-only Admin
│       ├── View all orgs
│       └── No write access
│
└── Organizations
    └── <Organization>
        ├── Org Admin (Full Org Access)
        │   ├── Agent Team Lead
        │   │   ├── Manage agents
        │   │   ├── View escalations
        │   │   └── Priority assignment
        │   ├── Human Agent
        │   │   ├── Take escalated calls
        │   │   ├── View assigned queue
        │   │   └── Update ticket status
        │   ├── Content Manager
        │   │   ├── Upload documents
        │   │   ├── Manage knowledge base
        │   │   └── View search analytics
        │   └── Analyst
        │       ├── View dashboards
        │       ├── Export reports
        │       └── Question tagging
        │
        └── Customer (End User)
            ├── Chat interaction
            └── Voice interaction
```

### 3.1 Platform Roles

| Role | Who | Access Level |
|------|-----|--------------|
| **Super Admin** | Platform team | Full platform access |
| Read-only Admin | Team members | View all, no write access |
| Read-write Admin | Key decision makers | Platform changes |

### 3.2 Business/Organization Roles

| Role | Who | Access |
|------|-----|--------|
| **Org Admin** | Business owners | Full org access - manage all settings |
| **Agent Team Lead** | Team manager | Manage agent team, view escalations |
| **Human Agent** | Support staff | Take escalated calls/chats |
| **Content Manager** | Document owner | Upload/manage documents, analytics |
| **Analyst** | Analytics team | View-only access to reports |

### 3.3 Org-Level Permissions

Admins can create custom roles and assign granular permissions:

| Category | Permissions |
|----------|-------------|
| Documents | Upload, Delete, View, Edit Metadata |
| Knowledge Base | Configure RAG, Manage Chunks, View Search Analytics |
| AI / Voice | Configure Behavior, Manage MCP, Escalation Rules |
| Team | Create Roles, Invite Members, Assign Permissions, Manage Agents |
| Conversations | View Transcripts, Monitor Calls, Take Over Escalations |
| Analytics | View Dashboard, Export Reports, Question Tagging |
| Settings | Edit Org Profile, Manage Webhooks, Greeting Message |

---

## 4. Features

### 4.1 Knowledge Management

**Sources (MVP):**
- File upload (PDF, docs, FAQs)
- Manual content entry
- API integration (Notion, Confluence, Zendesk)

**RAG Pipeline:**
- Chunk size: 512 tokens, 64 overlap
- Top-20 retrieval, relevance threshold 0.5
- Per-org vector namespace isolation

**Multilingual Embedding Support:**
- Embedding model must support 100+ languages (e.g., BGE-M3, Jina Embeddings V3)
- Cross-lingual retrieval: customers can query in any language and retrieve relevant chunks regardless of document language
- Language detection applied at ingestion and query time; stored as metadata on every vector chunk
- Language metadata enables filtered retrieval and per-language analytics

**Future:** URL scraping (React-based pages require special handling)

### 4.2 System Prompt Configuration

**Note:** Platform's core system prompt CANNOT be overwritten by businesses. Businesses can only APPEND their custom instructions.

| Mode | Description |
|------|-------------|
| Auto-generate | AI analyzes existing docs → generates initial prompt |
| Auto + Edit | Auto-generated, user can customize/add to platform prompt |
| Custom Append | User adds instructions that append to platform prompt |

**Auto-Regeneration:** The system prompt automatically refreshes (as a background job) whenever the knowledge base changes — new document uploaded, document updated, or document deleted. Active sessions are not affected; the new prompt applies from the next session. Admins can also trigger regeneration manually at any time.

### 4.3 Greeting Message (Separate from System Prompt)

Configurable per organization:
- Toggle to enable/disable custom greeting
- User can set opening message (e.g., "Welcome to Ford. How can I help?")
- Separate from AI behavior/system prompt configuration

### 4.4 Voice Experience Commitments

These are platform-level guarantees that define the quality of every customer voice interaction.

| Commitment | Behaviour |
|------------|-----------|
| **Natural interruption** | Customer can interrupt the AI mid-response at any time; AI stops, listens, and continues from the new query |
| **Session resilience** | Connection drops under 5 seconds auto-recover with full context intact; reconnect attempts made up to 3× for drops up to 30 seconds |
| **Page refresh tolerance** | Refreshing the browser does not end the session; conversation resumes within a 30-minute window |
| **No partial responses** | If interrupted, the AI never delivers an incomplete or incorrect answer — it waits for the customer's full query |
| **Graceful degradation** | If any backend service is degraded, customers receive clear messaging and are offered escalation rather than a silent failure |

**Customer-Facing Session States:**

The customer interface reflects the real-time state of the conversation at all times.

| State | What It Means |
|-------|--------------|
| `Connecting` | Session is initializing |
| `Listening` | AI is capturing the customer's speech |
| `Thinking` | AI is processing, searching knowledge base, or calling tools |
| `Speaking` | AI is delivering its response |
| `Interrupted` | Customer spoke over the AI; AI has stopped and is listening |
| `Escalating` | Transferring to a human agent |
| `Ended` | Session is complete |

**Optional Audio Recording:**

Orgs can choose to enable session audio recording. When enabled:
- Audio is uploaded to secure storage after the session ends (not during, to avoid latency impact)
- Recordings are auto-deleted after 30 days (configurable per org)
- Orgs can disable audio recording entirely if not required

### 4.5 MCP Tools

**Platform-Provided (Mandatory):**
| Tool | Purpose |
|------|---------|
| RAG Retrieval | Query org's knowledge base |
| Send Notification | SMS/WhatsApp for escalation |
| Create Ticket | Log unresolved issues |
| Knowledge Gap Flag | Notify business about content gaps |
| **Credit Alert** | Alert internal team when API credits are low |

**Business-Provided (Future):**
- Companies can connect their own MCP servers
- Validation and sandboxing required
- Team-based MCP grouping

**Future connectors:** Notion, Google Calendar, Email, Generic API

### 4.6 Escalation

**Toggle-based with two modes:**

| Mode | Trigger | Action |
|------|---------|--------|
| **Live** | Human agent available | Route to available agent (Teams/Microsoft integration) |
| **Async** | No active agents | Send SMS/WhatsApp/Email notification to org |

**Agent Management:**
- Queue-based routing with priority system
- Check agent availability before routing
- Multiple agents can be assigned to same org
- Priority order configurable by admin

**Customer Escalation Journey:**

The customer is never left in the dark during a handoff. The platform delivers a consistent 4-phase experience:

| Phase | What the Customer Sees |
|-------|------------------------|
| **Initiated** | "I'm connecting you with a support specialist. Please hold." |
| **Waiting** | Animated connecting indicator with estimated wait time |
| **Agent Found** | "A specialist is on the way. Your conversation context has been shared." |
| **Transfer Complete** | Human agent joins via chat or phone based on org configuration |

**Context Handoff Guarantee:** The full conversation transcript, retrieved knowledge, and customer context are automatically passed to the human agent. The customer never has to repeat themselves.

**Escalation Triggers:**

Escalation can be triggered in four ways:

| Trigger | Description |
|---------|-------------|
| **Customer request** | Customer explicitly says "talk to a human" or similar |
| **AI confidence** | AI cannot answer after retries and falls back to escalation |
| **MCP failure** | A required tool call fails repeatedly |
| **Live score threshold** | Real-time conversation health score drops below the org-configured threshold (see Live Monitoring in Section 4.8) |

The last trigger means admins can configure how aggressively the platform auto-escalates — tighter thresholds escalate sooner, looser thresholds give the AI more attempts.

### 4.7 Feedback System (LLM Judge)

**Purpose:** Continuous improvement through AI evaluation.

**Function:**
- Judge evaluates whether AI answered correctly
- Identifies knowledge gaps in real-time
- Sends actionable feedback to business with context
- Business updates knowledge base → better future responses
- Question tagging by topic (e.g., Ford → Blue Cruise, Login Issues)

### 4.8 Analytics Dashboard

**Metrics tracked per org:**

| Category | Metrics | Refresh |
|----------|---------|---------|
| **Volume** | Total calls/chats, daily/weekly/monthly | Real-time |
| **Escalation** | % escalated, reasons, response times | Real-time |
| **AI Performance** | Answer success rate, struggle topics | Daily |
| **Knowledge Gaps** | Unanswered questions flagged | Daily |
| **User Satisfaction** | Composite score (see below) | Daily |
| **Peak Hours** | At what times most customers are calling | Daily |
| **Topic Tags** | Frequently asked question categories | Daily |

**Satisfaction Scoring Model:**

Satisfaction is measured automatically using a two-part system — no manual input required.

| Component | How It Works |
|-----------|-------------|
| **Voice Sentiment Analysis** | Real-time tone analysis during the call |
| **AI Rating Agent** | Separate AI agent reviews the full transcript post-call and assigns a quality score |

Combined score bands:

| Score | Rating |
|-------|--------|
| 85 – 100 | Excellent |
| 70 – 84 | Good |
| 50 – 69 | Needs Attention |
| 0 – 49 | Poor |

**Out-of-Box Question Tagging Categories:**

Every conversation is automatically tagged by topic. Businesses get these categories on day one, with the ability to add custom ones.

| Category | Examples |
|----------|----------|
| **Account** | Password reset, profile update, access issues |
| **Billing** | Payment, invoice, subscription, refund |
| **Product** | Features, pricing, how-to, compatibility |
| **Order** | Status, tracking, cancellation, return |
| **Technical** | Errors, bugs, integration, API |
| **General** | Greeting, feedback, complaint, compliment |
| **Unanswered** | Questions the AI could not answer (flagged for knowledge gap review) |

Admins can create additional custom categories and manually re-tag conversations as needed.

**Conversation Resolution Tracking:**

Every session is assigned a resolution outcome, used to calculate true automation effectiveness.

| Status | Meaning |
|--------|---------|
| `Resolved by AI` | AI handled the query; customer satisfied |
| `Resolved by Human` | Escalated and resolved by a human agent |
| `Unresolved` | Escalated but issue not resolved |
| `Unknown` | Session ended without a clear outcome |

**Implicit Resolution Rule:** If the same customer contacts support about the same issue within 24 hours of a previous session, that prior session is automatically marked as `Unresolved` — even if it was originally logged as resolved.

**Live Conversation Monitoring:**

Admins can monitor every active conversation in real time from the dashboard. Each conversation is scored continuously and colour-coded:

| Colour | Score Range | Action |
|--------|-------------|--------|
| 🟢 Green | 0.70 – 1.00 | Normal — no action needed |
| 🟡 Yellow | 0.50 – 0.69 | Monitor — flagged on dashboard |
| 🟠 Orange | 0.30 – 0.49 | At risk — takeover suggested |
| 🔴 Red | 0.00 – 0.29 | Critical — admin notified, immediate takeover recommended |

Admins can click any live conversation to view the full real-time transcript and per-signal score breakdown. The Red threshold and auto-escalation behaviour are configurable per org.

**Human Takeover:** When an admin takes over a flagged conversation, the customer hears: *"A support specialist has reviewed our conversation and would like to help."* The agent receives the full transcript, flagged reason, and a summary of what the AI tried — no context is lost.

**Platform-Level Analytics:**
- Cross-org performance overview
- Credit/usage monitoring
- System health metrics

---

## 5. Multi-Tenant Isolation

**Requirement:** Organization A must never be able to view, access, or infer Organization B's data.

**Isolation Points:**
- All database rows have `org_id`
- Vector DB namespaces per org
- S3 prefixes per org (`s3://bucket/org-{id}/...`)
- API gateway validates org context on every request
- MCP tools scoped to respective orgs
- Agent team assignments are org-specific

---

## 6. Notifications & Alerting

### 6.1 Notification Events

Every significant platform event triggers a notification to the right person through the right channel.

| Event | Who Gets Notified | Channels |
|-------|------------------|---------|
| API credit usage > 50% | Super Admin | In-app, Email |
| API credit usage > 80% | Super Admin | In-app, Email (urgent) |
| API credit usage > 95% | Super Admin | In-app, Email, SMS |
| API key pool failure | Super Admin | In-app, Email |
| Service failure / high error rates | Super Admin | In-app, Email |
| MCP server down | Org Admin | In-app, Email |
| Escalation triggered | Org Agents | In-app, configurable channel |
| Escalation threshold hit | Org Admin | In-app, Email |
| Document processed successfully | Org Admin | In-app |
| Knowledge gap identified | Content Manager | In-app, Email |
| System degraded | Super Admin | In-app, Email |

### 6.2 Notification Channels

| Channel | Configurable | Notes |
|---------|-------------|-------|
| In-app | Always on | Unread badge counter visible in dashboard |
| Email | On/off, digest frequency configurable | |
| SMS | Critical events only | |
| Slack / Teams | Yes — webhook URL per org | |

---

## 7. Error Handling

| Pattern | Details |
|---------|---------|
| Retry + Backoff | 3 retries, exponential (1s, 2s, 4s) |
| Circuit Breaker | Trip after 5 failures, half-open after 30s |
| Timeout | RAG: 3s, MCP: 5s, configurable per org |
| Fallback Messages | Configurable per org |
| Dead Letter Queue | Failed escalations queued for retry |

---

## 8. Success Metrics & Platform Capacity

### 8.1 Success Metrics

| Metric | Target |
|--------|--------|
| Automation rate | 80–90% |
| Escalation rate | <20% |
| AI accuracy | Judge scores >80% |
| AI response latency | <2 seconds end-to-end |
| Voice audio latency | <500ms (microphone to speaker) |
| Session initialization | <2 seconds |
| Session resumption | <5 seconds |
| Customer satisfaction | Composite score >70 (Good band) |

**Additional KPIs:**
- Time to onboard (speed to go live)
- Knowledge base coverage (% of FAQ topics covered)
- Escalation feedback loop speed

### 8.2 Platform Capacity & SLA

**Availability**

| Level | Target |
|-------|--------|
| Platform uptime | 99.9% |
| Max degraded time before escalation | <1 hour |

**Capacity Commitments (MVP → Scale)**

| Metric | MVP | Scale |
|--------|-----|-------|
| Concurrent voice calls | 50 | 500 |
| Organizations supported | 10 | 100+ |
| Documents per org | 100 | 10,000 |
| Sessions per day (platform-wide) | 1,000 | 50,000 |
| Knowledge base searchable after upload | Within 5 min | Within 5 min |

**Security Requirements**

| Requirement | Policy |
|-------------|--------|
| Multi-factor authentication | Required for Super Admin; recommended for Org Admin |
| Encryption at rest | All data — database, file storage, cache |
| Encryption in transit | TLS 1.3 |
| Audit logging | All admin actions logged with 1-year retention |

---

## 9. MVP Scope

| Component | MVP |
|-----------|-----|
| User Auth + Org Onboarding | ✅ |
| Knowledge Upload (File + Manual + API) | ✅ |
| RAG-based AI responses | ✅ |
| Multilingual Embedding Support (100+ languages, cross-lingual retrieval) | ✅ |
| System Prompt (Auto + Edit, platform prompt protected) | ✅ |
| Greeting Message Configuration | ✅ |
| Chat + Voice channels | ✅ |
| Customer Access (Subdomain + Widget + API) | ✅ |
| MCP tools (Platform-provided mandatory) | ✅ |
| Escalation (Live queue + Async notification) | ✅ |
| Human Agent team management | ✅ |
| LLM Judge + Feedback | ✅ |
| Question Tagging | ✅ |
| Analytics Dashboard | ✅ |
| Error Handling + Alerting | ✅ |
| Multi-tenant Isolation | ✅ |

## Out of Scope (v0.2+)

- URL scraping (React pages require special handling)
- Business-provided MCP servers
- Pre-built connectors (Notion, Calendar, etc.)
- SSO authentication
- Custom domain routing
- Onboarding wizard detail

---

## 10. Delivery Roadmap

| Phase | Focus | Key Deliverables |
|-------|-------|-----------------|
| **Phase 1 — Core (POC)** | Voice + RAG foundation | Voice pipeline, document upload, RAG-based responses, session management, basic escalation, analytics dashboard, multi-tenant isolation |
| **Phase 2 — Integrations** | MCP + tool connectivity | Admin UI for MCP server configuration, tool execution engine, RAG + live data combined in responses |
| **Phase 3 — Expansion** | Multi-channel + learning loop | SMS, WhatsApp, chat integrations; advanced trend analytics; predictive escalation; admin resolution feedback fed back into training |

Items in the Out of Scope list above are candidates for Phase 2 and Phase 3, prioritised based on business demand.

---

## 11. Key Decisions Summary

| Decision | Choice |
|----------|--------|
| AI | Gemini Live API |
| Roles | Super Admin + Org Roles (Admin, Agent, Content, Analyst) |
| Channels | Voice + Chat (both at launch) |
| Customer Access | Subdomain + Widget + API |
| Knowledge | File upload + Manual + API integration |
| Embedding | Multilingual model (BGE-M3 / Jina V3, 100+ languages, cross-lingual) |
| MCP | Platform tools (mandatory) + Business tools (future) |
| Escalation | Queue-based Live / Async Toggle |
| System Prompt | Auto-generate + Edit + Append only (no override) |
| Greeting | Separate configurable message |
| Feedback | LLM Judge with question tagging |
| Isolation | Strict org-level (no cross-org data access) |
| Platform Alerting | Email on failures + Credit alerts |

---

## 12. Pricing Reference

### 11.1 Gemini Multimodal Live API (AI & Voice)

Google's real-time, low-latency API for voice, audio, and video interactions — the core AI engine powering this platform. Billed by tokens processed.

> **Important:** Live interactions stream constant data. Audio is charged at **25 tokens/second**, meaning a 2-minute voice call consumes ~3,000 audio input tokens before any output is counted.

**Free Tier (Google AI Studio)**

| Limit | Value |
|-------|-------|
| Requests per day | Up to 1,500 (model-dependent, e.g. Gemini 1.5 Flash) |
| Cost | $0 |

**Paid Tier (Pay-As-You-Go)**

| Token Type | Rate |
|------------|------|
| Text Input | $0.075 – $0.35 per 1M tokens |
| Audio / Video Input | $0.70 – $2.10 per 1M tokens |
| Output | $0.30 – $1.50 per 1M tokens |

**Cost Estimation Example (per voice conversation):**

| Scenario | Approx. Tokens | Approx. Cost |
|----------|---------------|--------------|
| 1-min voice call (audio input only) | ~1,500 audio tokens | ~$0.001 – $0.003 |
| 3-min voice call + RAG context output | ~4,500 audio + ~500 output tokens | ~$0.004 – $0.010 |

> These are illustrative estimates. Actual costs vary by model tier and conversation complexity.

**Reference:** [Gemini API Pricing — Gemini 2.5 Flash Live Preview](https://ai.google.dev/gemini-api/docs/pricing#gemini-3.1-flash-live-preview)

---

*Note: Technical implementation details (tech stack, database schema, API contracts) are covered in TRD.*
