---
title: Voice AI Customer Support Platform (PoC)
status: draft
created: 2026-05-24
updated: 2026-05-24
author: Sreyash Reddy
---

# Voice AI Customer Support Platform — Product Requirements Document

## 1. Overview

### 1.1 Product Name
Voice AI Customer Support Platform (PoC)

### 1.2 Product Type
B2B SaaS — Multi-tenant voice-first AI customer support platform

### 1.3 Core Functionality Summary
A voice-first customer support platform that replaces human agents with an autonomous AI agent powered by Gemini Live API. The agent listens to customer queries, retrieves relevant context from organization-specific knowledge bases (via RAG), and responds in real-time with a friendly, trust-building tone. Text chat is available as a fallback, but voice is the primary modality.

### 1.4 Target Market
Industry-agnostic — designed for organizations of all sizes and verticals that need to automate customer support while maintaining high satisfaction scores.

---

## 2. Goals

### 2.1 Primary Goal
Automate customer support interactions with an AI voice agent that resolves queries without human intervention, reducing operational costs and improving response times.

### 2.2 Supporting Goals

| Goal | Description |
|------|-------------|
| **High Automation Rate** | Maximize the percentage of customer queries resolved autonomously by the AI agent |
| **Customer Satisfaction** | Ensure customers feel heard, valued, and confident their problems will be resolved |
| **Organization Control** | Provide org admins with full visibility into conversation quality, gaps, and knowledge improvements |
| **Platform Intelligence** | Enable continuous improvement through feedback loops that surface unanswered questions and knowledge gaps |

### 2.3 Success Definition
An organization using this platform should be able to:
- Handle 80%+ of customer queries without human intervention
- Provide instant, 24/7 support to customers
- Continuously improve AI performance by adding documents to fill knowledge gaps
- Maintain clear visibility into AI performance via dashboards and satisfaction scores

---

## 3. Non-Goals (Out of Scope for PoC)

The following are explicitly NOT in scope for the PoC phase:

| Item | Reason |
|------|--------|
| **Human-in-the-loop supervision** | Agent is fully autonomous; escalation forwards to human agents but no ongoing AI monitoring by humans |
| **Action execution via MCP** | MCP is scoped to data retrieval only, not for performing transactions (booking, canceling, etc.) |
| **Learning loop automation** | Admin resolutions do not automatically retrain the agent; improvement is manual via document additions |
| **Cross-organization intelligence** | Each org's data and feedback remains siloed; no aggregated platform-wide insights in PoC |
| **Complex RBAC customization** | Only Super Admin and Organization Admin roles defined; additional roles deferred |
| **Escalation threshold calibration** | Binary escalation (can/cannot answer); confidence-based thresholds deferred to implementation |
| **Target automation rate specifics** | 80% target noted as directional; exact calibration deferred |

---

## 4. User Stories

### 4.1 End User (Customer calling for support)

**US-001: Voice Support Request**
> As a customer, I want to speak naturally to an AI agent so that I can get my question answered quickly without navigating complex menus.

**US-002: Text Fallback**
> As a customer, if voice is not convenient, I want to switch to text chat so that I can still get support via typing.

**US-003: Seamless Escalation**
> As a customer, if my question cannot be answered, I want to be seamlessly forwarded to a human agent without repeating myself.

**US-004: Trust-building Interaction**
> As a customer, I want the AI agent to speak in a friendly, empathetic tone so that I feel confident my problem will be resolved.

### 4.2 Organization Admin

**US-005: Document Management**
> As an org admin, I want to upload documents and manage data sources so that the AI agent can answer questions specific to my organization.

**US-006: Conversation Oversight**
> As an org admin, I want to see conversation transcripts and scores so that I can understand how well the AI is performing.

**US-007: Gap Identification**
> As an org admin, I want to see which questions the AI could not answer so that I can add documents to fill those gaps.

**US-008: Question Analytics**
> As an org admin, I want to see tagged question topics so that I can identify frequently asked questions and prioritize knowledge additions.

**US-009: Escalation Notification**
> As an org admin, I want to be immediately notified when the AI escalates a conversation so that I can track AI quality in real-time.

### 4.3 Super Admin (Platform Provider)

**US-010: Organization Management**
> As the platform provider, I want to manage organizations (onboard, configure, monitor) so that I can deliver the service to multiple clients.

**US-011: Platform Feedback**
> As the platform provider, I want to receive feedback from organizations so that I can improve the platform for all clients.

---

## 5. Feature Requirements

### 5.1 Core Platform Features

| ID | Feature | Description | Priority |
|----|---------|-------------|----------|
| FR-001 | **Organization Onboarding** | Create organization, assign admin, configure dashboard | Required |
| FR-002 | **Multi-Tenant Isolation** | Ensure documents, MCP connections, and embeddings are isolated per organization | Required |
| FR-003 | **Document Upload & Management** | Allow org admins to upload, organize, and manage knowledge base documents | Required |
| FR-004 | **Vector Embedding Pipeline** | Convert uploaded documents to vector embeddings for RAG retrieval | Required |
| FR-005 | **MCP Server Configuration** | Allow org admins to connect MCP servers for data retrieval | Required |
| FR-006 | **Voice Call Interface** | WebRTC-based voice call UI where customers speak and AI responds | Required |
| FR-007 | **Text Chat Fallback** | Text-based chat interface as alternative to voice | Required |
| FR-008 | **RAG Retrieval** | Real-time retrieval of relevant document context to augment AI responses | Required |
| FR-009 | **MCP Data Integration** | Fetch data from connected MCP servers to augment AI responses | Required |
| FR-010 | **Agent Persona Configuration** | Configure AI voice tone (friendly, trust-building) via Gemini Live API settings | Required |

### 5.2 Escalation Features

| ID | Feature | Description | Priority |
|----|---------|-------------|----------|
| FR-011 | **Automatic Escalation** | When AI cannot answer, immediately forward conversation to human agent | Required |
| FR-012 | **Escalation Dashboard** | Display escalated conversations with "unanswered" flag in admin dashboard | Required |
| FR-013 | **Admin Notification** | Immediately notify org admin when escalation occurs | Required |
| FR-014 | **Conversation Handoff** | Preserve full conversation context for human agent to avoid customer repeating information | Required |

### 5.3 Analytics & Feedback Features

| ID | Feature | Description | Priority |
|----|---------|-------------|----------|
| FR-015 | **Conversation Transcripts** | Store and display full conversation transcripts to org admin | Required |
| FR-016 | **Satisfaction Scoring** | Score each conversation for customer satisfaction | Required |
| FR-017 | **Dissatisfaction Analysis** | Identify unanswered questions and knowledge gaps | Required |
| FR-018 | **Question Tagging** | Tag questions by topic to identify frequently asked categories | Required |
| FR-019 | **Dashboard Overview** | Admin dashboard showing automation rate, satisfaction scores, escalation counts | Required |

### 5.4 RBAC Features

| ID | Feature | Description | Priority |
|----|---------|-------------|----------|
| FR-020 | **Super Admin Role** | Platform provider with full access to all organizations and platform settings | Required |
| FR-021 | **Organization Admin Role** | Per-organization admin with access to their org's documents, analytics, and settings | Required |
| FR-022 | **Role-Per-Organization Matrix** | Support users holding different roles across different organizations | Required |

### 5.5 Feedback Channel Features

| ID | Feature | Description | Priority |
|----|---------|-------------|----------|
| FR-023 | **Organization → Super Admin Feedback** | Allow org admins to provide feedback to platform provider | Required |
| FR-024 | **No Direct Customer → Platform Channel** | Ensure customer feedback only flows through organization admin | Required |

---

## 6. Technical Requirements

### 6.1 Core Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Voice API** | Gemini Live API | Low-latency, real-time voice interaction with configurable voice persona |
| **WebRTC** | Client-side audio streaming | Enable real-time voice communication |
| **Vector Database** | TBD (e.g., Pinecone, Weaviate, Chroma) | Store document embeddings for RAG retrieval |
| **Document Storage** | TBD (e.g., S3 with per-org prefixes) | Store original documents with organization isolation |
| **Embedding Model** | TBD (e.g., OpenAI Embeddings, Vertex AI) | Convert documents to vector representations |
| **MCP Protocol** | Model Context Protocol | Connect to organization's external data sources |

### 6.2 Multi-Tenant Isolation Requirements

| Requirement | Implementation |
|-------------|----------------|
| **Document Isolation** | Per-org S3 prefixes with KMS encryption |
| **Vector DB Isolation** | Namespace by organization_id |
| **MCP Isolation** | org_id on every MCP config row; tenant-context middleware on all tool calls |
| **No Cross-Org Access** | Architectural impossibility, not just logical filtering |

### 6.3 Performance Requirements

| Metric | Target |
|--------|--------|
| **Voice Response Latency** | Sub-second (P95 target to be calibrated during implementation) |
| **RAG Retrieval Latency** | Minimize impact on voice response time via pre-fetch and caching strategies |
| **Escalation Forward Latency** | Immediate decision, near-instant handoff to human agent |

### 6.4 Security Requirements

| Requirement | Description |
|-------------|-------------|
| **SSRF Prevention** | Sandbox MCP outbound calls to prevent server-side request forgery |
| **Content Validation** | Validate uploaded document types, virus scan if feasible |
| **MCP Credential Isolation** | Store per-org API keys/OAuth tokens securely, scoped to org only |
| **Audit Logging** | Log all MCP invocations with organization attribution |

---

## 7. User Journeys

### 7.1 Customer Support Call Journey

```
1. Customer initiates voice call via web interface
2. AI agent greets customer in friendly tone
3. Customer speaks their query
4. AI agent processes query:
   a. Classifies intent (does this need RAG? MCP? both?)
   b. Retrieves relevant context from org knowledge base
   c. Fetches any required data from MCP servers
   d. Generates response
5. AI agent responds with context-aware answer
6. Customer asks follow-up or concludes call
7. Conversation scored for satisfaction
8. If escalation needed → forward to human with full context
```

### 7.2 Admin Knowledge Improvement Journey

```
1. Customer query goes unanswered by AI
2. Conversation flagged as "unanswered" in dashboard
3. Admin receives notification
4. Admin reviews transcript and identifies gap
5. Admin uploads relevant document(s) to fill gap
6. Document processed into embeddings
7. AI agent can now answer similar questions
8. Automation rate improves
```

### 7.3 Organization Onboarding Journey

```
1. Super Admin creates new organization
2. Organization Admin receives access
3. Admin uploads initial knowledge base documents
4. Admin configures MCP server connections (if applicable)
5. AI agent trained on org-specific data
6. Voice interface ready for customer calls
```

---

## 8. Success Metrics

### 8.1 Primary Metrics

| Metric | Target | Measurement |
|--------|--------|--------------|
| **Automation Rate** | 80%+ of queries resolved without human intervention | Escalations / Total queries |
| **Customer Satisfaction Score** | >80% positive (TBD calibration) | Per-conversation satisfaction score |
| **Response Time** | <2s average voice response | End-to-end latency measurement |

### 8.2 Secondary Metrics

| Metric | Description |
|--------|-------------|
| **Escalation Rate** | Percentage of conversations requiring human handoff |
| **Knowledge Gap Resolution Time** | Time between flagged question and document addition by admin |
| **Admin Engagement** | Frequency of admin logins and document additions |
| **Question Coverage** | Percentage of question tags with corresponding documents |

### 8.3 Counter-Metrics (Things We Want to Minimize)

| Metric | Description |
|--------|-------------|
| **Repeat Escalations** | Same questions escalated multiple times (indicates unresolved gaps) |
| **Handoff Friction** | Customer having to repeat information after escalation |
| **Response Time Degradation** | Latency increases under load |

---

## 9. Timeline / Phases

### 9.1 Phase 1: POC — RAG Core

**Duration:** [TBD based on resources]

**Scope:**
- Organization onboarding and dashboard
- Document upload and vector embedding pipeline
- Voice call interface (Gemini Live API)
- RAG retrieval and response generation
- Conversation scoring and dashboard analytics
- Escalation to human agent (basic)

**Deliverables:**
- Functional POC with at least one pilot organization
- Core voice interaction working
- Admin dashboard with conversation visibility

### 9.2 Phase 2: MCP Integration

**Duration:** [TBD]

**Scope:**
- MCP server configuration per organization
- Data retrieval integration via MCP
- Enhanced response generation with MCP data

**Deliverables:**
- MCP connections available to organizations
- AI agent can pull real-time data from connected systems

### 9.3 Phase 3: Platform Enhancement (Future)

**Scope (TBD):**
- Learning loop for agent improvement
- Advanced RBAC customization
- Cross-org intelligence (aggregated insights)
- Enhanced analytics and reporting

---

## 10. Open Questions

| ID | Question | Owner | Status |
|----|----------|-------|--------|
| OQ-006 | How do we measure satisfaction? | Sreyash | **RESOLVED** — Two-part: (1) Sentiment analysis (automatic), (2) Separate rating agent reviews and rates conversations |
| OQ-002 | What is the escalation trigger — agent decides and triggers escalation tool? | Sreyash | **RESOLVED** — Agent-centric: agent has agency to decide when it can't answer and triggers escalation tool |
| OQ-002b | What is the escalation mechanism? | Technical | Agent triggers "escalate" tool, not hardcoded binary |
| OQ-003 | What format should admin resolutions take for future agent training? | Sreyash | Open |
| OQ-004 | Which vector database and embedding model to use? | Technical | Open |
| OQ-005 | What is the P95 latency target for voice responses? | Technical | Open |

---

## 11. Assumptions

| ID | Assumption | Impact if Wrong |
|----|------------|-----------------|
| ASMP-001 | Gemini Live API provides sufficient latency for interactive voice | May need to optimize RAG pipeline or choose alternative |
| ASMP-002 | Organizations have documents that can be converted to knowledge bases | Without documents, RAG provides no value |
| ASMP-003 | MCP integration is primarily read-only data retrieval | Write operations may require additional safety measures |
| ASMP-004 | 80% automation rate is achievable with proper document coverage | May require iteration on document strategy and AI calibration |

---

*Document Status: Draft — Awaiting review and finalization*