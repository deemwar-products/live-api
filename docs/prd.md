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

### 4.3 Greeting Message (Separate from System Prompt)

Configurable per organization:
- Toggle to enable/disable custom greeting
- User can set opening message (e.g., "Welcome to Ford. How can I help?")
- Separate from AI behavior/system prompt configuration

### 4.4 MCP Tools

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

### 4.5 Escalation

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

### 4.6 Feedback System (LLM Judge)

**Purpose:** Continuous improvement through AI evaluation.

**Function:**
- Judge evaluates whether AI answered correctly
- Identifies knowledge gaps in real-time
- Sends actionable feedback to business with context
- Business updates knowledge base → better future responses
- Question tagging by topic (e.g., Ford → Blue Cruise, Login Issues)

### 4.7 Analytics Dashboard

**Metrics tracked per org:**

| Category | Metrics |
|----------|---------|
| **Volume** | Total calls/chats, daily/weekly/monthly |
| **Escalation** | % escalated, reasons, response times |
| **AI Performance** | Answer success rate, struggle topics |
| **Knowledge Gaps** | Unanswered questions flagged |
| **User Satisfaction** | Feedback, ratings |
| **Peak Hours** | At what times most customers are calling |
| **Topic Tags** | Frequently asked question categories |

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

## 6. Alerting

### 6.1 Platform Alerting (Internal)

**Alert triggers:**
| Trigger | Action |
|---------|--------|
| Service failures | Email to platform team |
| High error rates | Email to platform team |
| Escalation channel failure | Email to platform team |
| **Low API credits** | Email to platform team (via Credit Alert MCP) |

### 6.2 Business Alerting (External)

- Escalation notifications to org agents
- Knowledge gap alerts to content managers
- Usage/cost alerts (future)

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

## 8. Success Metrics

| Metric | Target |
|--------|--------|
| Automation rate | 80-90% |
| Escalation rate | <20% |
| AI accuracy | Judge scores >80% |
| Response latency | <2 seconds |
| Customer satisfaction | TBD post-launch |

**Additional KPIs:**
- Time to onboard (speed to go live)
- Knowledge base coverage (% of FAQ topics covered)
- Escalation feedback loop speed

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

## 10. Key Decisions Summary

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

## 11. Pricing Reference

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
