# Voice AI Customer Support Platform - PRD

## 1. Overview

**What:** AI-powered voice and chat customer support platform for businesses.

**Core Loop:**
1. Business uploads knowledge base (docs, FAQs, integrations)
2. Customers interact via voice or chat on business's subdomain
3. AI responds using RAG + MCP + Gemini Live API
4. Escalation to human when needed (configurable toggle)
5. LLM Judge evaluates conversations and feeds back knowledge gaps

**Target:** 80-90% automation rate

---

## 2. Architecture

### 2.1 Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| Primary DB | PostgreSQL |
| Analytics DB | DuckDB + S3 Parquet |
| Vector DB | Per-org namespace |
| Document Store | S3 (per-org prefix) |
| Session Store | Redis |
| AI | Gemini Live API |
| Integrations | MCP (Model Context Protocol) |

### 2.2 Three Interfaces

| Interface | Who | Access |
|-----------|-----|--------|
| Platform Dashboard | Super Admin | All orgs, system settings, alerts |
| Business Dashboard | Org Admins | Own docs, team, AI config, analytics |
| Customer Web App | End customers | Chat/Voice interaction |

### 2.3 Customer Access

- Subdomain per org: `{org-slug}.platform.com`
- Embeddable widget for external websites
- Voice + Chat available from launch

---

## 3. Roles & Permissions

### 3.1 Platform Roles

| Role | Who |
|------|-----|
| Super Admin | Platform team |
| Admin | Business owners |
| Customer | End consumers |

### 3.2 Org-Level Permissions

Admins can assign granular permissions to team members:

| Category | Permissions |
|----------|-------------|
| Documents | Upload, Delete, View, Edit Metadata |
| Knowledge Base | Configure RAG, Manage Chunks, View Analytics |
| AI / Voice | Configure Behavior, Manage MCP, Escalation Rules |
| Team | Invite, Assign Permissions, Remove, View Activity |
| Conversations | View Transcripts, Monitor Calls, Take Over |
| Analytics | View Dashboard, Export Reports |
| Settings | Edit Org Profile, Manage Webhooks |

---

## 4. Features

### 4.1 Knowledge Management

**Sources (MVP):**
- File upload (PDF, docs, FAQs)
- Manual entry
- API integration (Notion, Confluence, Zendesk)

**RAG:** 512 tokens chunk, 64 overlap, top-20 retrieval

**Future:** URL scraping

### 4.2 System Prompt Configuration

| Mode | Description |
|------|-------------|
| Auto-generate | AI analyzes docs → generates prompt |
| Auto + Edit | Auto-generated, user can customize |
| Custom | User writes prompt from scratch |

### 4.3 MCP Tools

| Provider | Tools |
|----------|-------|
| Platform | RAG, Send Notification, Create Ticket, Knowledge Gap Flag |
| Business | Own MCP servers (future) |

**Future connectors:** Notion, Google Calendar, Email, Generic API

### 4.4 Escalation

**Toggle-based:**

| Mode | Action |
|------|--------|
| Live | Real-time handoff to human agent |
| Async | Notification via SMS/WhatsApp/Email |

### 4.5 Feedback System (LLM Judge)

- Evaluates conversation quality
- Identifies knowledge gaps
- Sends actionable feedback to business with context
- Business updates KB → better future responses

### 4.6 Analytics Dashboard

| Category | Metrics |
|----------|----------|
| Volume | Total calls/chats, daily/weekly/monthly |
| Escalation | % escalated, reasons |
| AI Performance | Answer success rate, struggle topics |
| Knowledge Gaps | Unanswered questions flagged |
| Satisfaction | Feedback, ratings |
| Additional | Peak hours, failed queries, topics breakdown |

---

## 5. Multi-Tenant Isolation

**Requirement:** Org A cannot view Org B's data.

**Isolation:** All rows have `org_id`, separate namespaces, S3 prefixes, API validates org context.

---

## 6. Error Handling & Alerting

| Pattern | Details |
|---------|---------|
| Retry + Backoff | 3 retries, exponential (1s, 2s, 4s) |
| Circuit Breaker | Trip after 5 failures, half-open after 30s |
| Timeout | RAG: 3s, MCP: 5s, configurable |
| Fallback Messages | Configurable per org |
| Dead Letter Queue | Failed escalations queued for retry |

**Platform Alerting:** Email to core team on service failures.

---

## 7. Success Metrics

| Metric | Target |
|--------|--------|
| Automation rate | 80-90% |
| Escalation rate | <20% |
| AI accuracy | Judge scores >80% |
| Response latency | <2 seconds |

---

## 8. MVP Scope

| Component | MVP |
|-----------|-----|
| User Auth + Org Onboarding | ✅ |
| Knowledge Upload (File + Manual + API) | ✅ |
| RAG-based AI responses | ✅ |
| System Prompt (Auto + Edit + Custom) | ✅ |
| Chat + Voice channels | ✅ |
| MCP tools (Platform) | ✅ |
| Escalation (Live + Async) | ✅ |
| LLM Judge + Feedback | ✅ |
| Analytics Dashboard | ✅ |
| Error Handling + Alerting | ✅ |
| Multi-tenant Isolation | ✅ |

## Out of Scope (v0.2+)

URL scraping, Business-provided MCP, Pre-built connectors, SSO, Custom domains

---

## 9. Key Decisions

| Decision | Choice |
|----------|--------|
| Language | Go |
| Database | PostgreSQL + DuckDB/S3 |
| AI | Gemini Live API |
| Roles | Super Admin + Admin (granular) + Customer |
| Channels | Voice + Chat |
| Knowledge | File + Manual + API |
| MCP | Hybrid (Platform + Future business) |
| Escalation | Toggle: Live / Async |
| System Prompt | Auto / Auto+Edit / Custom |
| Feedback | LLM Judge |
| Isolation | Strict org-level |
