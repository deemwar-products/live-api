---
title: Voice AI Customer Support Platform — Epics and User Stories
status: draft
created: 2026-05-24
updated: 2026-05-24
author: Sreyash Reddy
source_prd: docs/prd.md
source_architecture: docs/voice-ai-platform-architecture.md
source_ux: docs/ux-design-specification.md
---

# Voice AI Customer Support Platform — Epics and User Stories

## Document Overview

This document breaks down the Voice AI Customer Support Platform into implementable epics and user stories. Stories are organized by phase and feature area, with acceptance criteria for each.

---

## EPIC STRUCTURE

| Epic | Phase | Priority | Description |
|------|-------|----------|-------------|
| **E-001** | 1 | P0 | Voice Call Interface — Customer-facing voice support |
| **E-002** | 1 | P0 | RAG Pipeline — Document ingestion and retrieval |
| **E-003** | 1 | P0 | Organization Admin Dashboard — Analytics and management |
| **E-004** | 1 | P0 | Organization Onboarding — Setup and configuration |
| **E-005** | 1 | P1 | Escalation System — Human handoff when AI fails |
| **E-006** | 1 | P1 | Satisfaction Scoring — Post-call feedback |
| **E-007** | 1 | P1 | Text Chat Fallback — Alternative to voice |
| **E-008** | 2 | P1 | MCP Integration — External data source connectivity |
| **E-009** | 2 | P2 | Question Tagging — Topic classification |
| **E-010** | 3 | P2 | Platform Super Admin — Multi-org management |
| **E-011** | 3 | P3 | Platform Enhancement — Cross-org insights |

---

## EPIC E-001: Voice Call Interface

**Phase:** 1 (POC Core)
**Priority:** P0
**Description:** Customer-facing voice support interface using Gemini Live API

### User Stories

#### US-E001-001: Initiate Voice Call
**As a** customer, **I want to** initiate a voice support call with the AI agent, **so that I** can speak naturally without navigating menus.

**Acceptance Criteria:**
- [ ] Customer can open support widget/page
- [ ] Microphone permission is requested on first use
- [ ] Voice call starts with AI greeting
- [ ] AI greeting is configurable per organization
- [ ] Call session is created and stored in Redis
- [ ] WebRTC connection is established with Gemini Live API

#### US-E001-002: Voice Conversation with RAG
**As a** customer, **I want to** speak my query and receive an AI response with context from my organization's knowledge base, **so that I** get accurate, relevant answers.

**Acceptance Criteria:**
- [ ] Customer speech is captured in 100ms chunks
- [ ] Speech is transcribed in real-time
- [ ] Query is embedded and searched against vector DB
- [ ] Relevant context (chunks with >0.5 relevance) is retrieved
- [ ] AI responds with context-augmented answer via TTS
- [ ] Response latency is <2s P95
- [ ] Conversation history is maintained for context

#### US-E001-003: Voice Processing States
**As a** customer, **I want to** see visual feedback during AI processing, **so that I** know the system is working and don't think the call dropped.

**Acceptance Criteria:**
- [ ] "Listening" state shows animated waveform
- [ ] "Thinking" state shows animated dots
- [ ] "Speaking" state shows animated avatar
- [ ] State transitions are smooth with no jarring changes
- [ ] No silent state lasts longer than 1 second

#### US-E001-004: Voice Interruption Handling
**As a** customer, **I want to** interrupt the AI mid-response if I need to clarify or redirect, **so that I** have control over the conversation flow.

**Acceptance Criteria:**
- [ ] Customer speaking stops AI response immediately
- [ ] "Interrupted" indicator briefly shown (1s)
- [ ] AI responds to new input with interruption acknowledged
- [ ] Customer can interrupt multiple times
- [ ] Conversation context is preserved through interruptions

#### US-E001-005: Voice Connection Recovery
**As a** customer, **I want to** not lose my conversation if my connection drops briefly, **so that I** don't have to repeat everything.

**Acceptance Criteria:**
- [ ] Connection drop < 5s triggers auto-reconnect
- [ ] Connection drop 5-30s shows "reconnecting" UI
- [ ] Up to 3 reconnect attempts are made
- [ ] Session state is preserved in Redis
- [ ] If reconnection fails, customer is notified with option to retry
- [ ] Browser refresh preserves session via session_id

#### US-E001-006: End Voice Call
**As a** customer, **I want to** end the call gracefully and provide feedback, **so that I** feel the interaction was complete.

**Acceptance Criteria:**
- [ ] Customer can end call via button or voice ("thanks, goodbye")
- [ ] AI provides closing response before disconnect
- [ ] Conversation transcript is saved
- [ ] Post-call survey is displayed (optional)
- [ ] Call metrics are recorded (duration, turns)

#### US-E001-007: Voice Greeting Configuration
**As an** organization admin, **I want to** configure the AI greeting message, **so that I** can personalize the customer experience.

**Acceptance Criteria:**
- [ ] Admin can set custom greeting text
- [ ] Admin can choose AI voice from available options
- [ ] Admin can preview greeting before saving
- [ ] Greeting applies to all customer calls for the org

---

## EPIC E-002: RAG Pipeline

**Phase:** 1 (POC Core)
**Priority:** P0
**Description:** Document ingestion, embedding, and retrieval for AI responses

### User Stories

#### US-E002-001: Document Upload
**As an** organization admin, **I want to** upload documents to the knowledge base, **so that** the AI can answer questions about my organization.

**Acceptance Criteria:**
- [ ] Admin can drag-and-drop or browse for files
- [ ] Supported formats: PDF, DOCX, TXT, MD, HTML
- [ ] Max file size: 50MB per file
- [ ] File type and size validation occurs on upload
- [ ] Upload progress is displayed
- [ ] Document is stored in S3 with org-scoped prefix

#### US-E002-002: Document Metadata
**As an** organization admin, **I want to** provide metadata for uploaded documents, **so that** I can organize and search my knowledge base.

**Acceptance Criteria:**
- [ ] Admin can set document name
- [ ] Admin can select category (Policies, FAQs, How-to, Product, Custom)
- [ ] Admin can add custom tags
- [ ] Metadata is stored with document in S3
- [ ] Metadata is indexed for filtering and search

#### US-E002-003: Document Chunking
**As a** system, **I want to** split uploaded documents into chunks for embedding, **so that** I can retrieve relevant sections without loading entire documents.

**Acceptance Criteria:**
- [ ] Documents are split into ~512 token chunks
- [ ] Chunks overlap by 64 tokens to preserve context
- [ ] Chunks split at semantic boundaries (sentence-aware)
- [ ] Chunks below 100 tokens are discarded
- [ ] Chunks above 1024 tokens are further split

#### US-E002-004: Document Embedding
**As a** system, **I want to** convert document chunks into vector embeddings, **so that** I can perform semantic search.

**Acceptance Criteria:**
- [ ] Chunks are batched (100 per call) for embedding
- [ ] Embeddings are generated asynchronously
- [ ] Documents are searchable within 5 minutes of upload
- [ ] Embeddings are stored in vector DB with org namespace
- [ ] Processing status is visible in admin UI

#### US-E002-005: Document Management
**As an** organization admin, **I want to** view, edit, and delete my documents, **so that** I can keep the knowledge base accurate.

**Acceptance Criteria:**
- [ ] Admin sees list of all documents with status
- [ ] Admin can view document details and usage stats
- [ ] Admin can delete documents (cascades to embeddings)
- [ ] Admin can update documents (old embeddings deleted, new created)
- [ ] Documents show "Last updated" and "Chunks" count

#### US-E002-006: RAG Retrieval
**As a** system, **I want to** retrieve relevant document context for customer queries, **so that** the AI can provide accurate answers.

**Acceptance Criteria:**
- [ ] Query is embedded using same model as documents
- [ ] Top-20 most similar chunks are retrieved
- [ ] Chunks below 0.5 relevance are discarded
- [ ] Maximum 10 chunks are used in context
- [ ] Retrieved chunks include document source reference

#### US-E002-007: RAG Update on Document Change
**As a** system, **I want to** automatically update embeddings when documents change, **so that** AI responses reflect the latest information.

**Acceptance Criteria:**
- [ ] When document is updated, old embeddings are deleted
- [ ] New embeddings are generated for updated document
- [ ] When document is deleted, all embeddings are removed
- [ ] Updates do not affect unrelated documents

---

## EPIC E-003: Organization Admin Dashboard

**Phase:** 1 (POC Core)
**Priority:** P0
**Description:** Admin interface for managing knowledge, viewing analytics, and identifying gaps

### User Stories

#### US-E003-001: Dashboard Overview
**As an** organization admin, **I want to** see key metrics at a glance, **so that I** can quickly understand AI performance.

**Acceptance Criteria:**
- [ ] Dashboard shows: Automation Rate, Satisfaction Score, Escalations Today
- [ ] Each metric shows current value and trend (up/down %)
- [ ] Metrics update in real-time
- [ ] Admin can click on metric to see detailed view
- [ ] Trend graphs show last 7 days

#### US-E003-002: Conversation List
**As an** organization admin, **I want to** see recent conversations, **so that I** can monitor AI quality.

**Acceptance Criteria:**
- [ ] List shows last 50 conversations
- [ ] Each item shows: ID, time, status (resolved/escalated), score
- [ ] Status badges: Green (resolved), Red (escalated), Yellow (low score)
- [ ] Admin can click to view full transcript
- [ ] List can be filtered and searched

#### US-E003-003: Conversation Detail
**As an** organization admin, **I want to** view full transcript and context of a conversation, **so that I** can understand what happened.

**Acceptance Criteria:**
- [ ] Full transcript is displayed with timestamps
- [ ] Escalation reason is shown if applicable
- [ ] Retrieved context chunks are visible
- [ ] Question tags are displayed
- [ ] Admin can mark conversation as reviewed

#### US-E003-004: Knowledge Gap Identification
**As an** organization admin, **I want to** see which questions the AI couldn't answer, **so that I** can add documents to fill those gaps.

**Acceptance Criteria:**
- [ ] "Knowledge Gaps" section shows top unanswered topics
- [ ] Each gap shows: topic, unanswered count, trend
- [ ] Admin can click gap to see example questions
- [ ] "Add Document" action is available per gap
- [ ] Gaps update daily based on escalation analysis

#### US-E003-005: Satisfaction Score Analysis
**As an** organization admin, **I want to** understand why some conversations had low satisfaction scores, **so that I** can improve the AI.

**Acceptance Criteria:**
- [ ] Admin can view satisfaction score breakdown
- [ ] Score components (rating, duration, sentiment) are visible
- [ ] Low-scoring conversations are highlighted
- [ ] Admin can identify patterns in low scores

#### US-E003-006: Question Analytics
**As an** organization admin, **I want to** see which topics are most frequently asked, **so that I** can prioritize knowledge additions.

**Acceptance Criteria:**
- [ ] Analytics shows top question tags by volume
- [ ] Tag trends show increasing/decreasing patterns
- [ ] "Question Coverage" metric shows % with documents
- [ ] Admin can drill down by tag to see example questions

#### US-E003-007: Escalation Notification
**As an** organization admin, **I want to** be immediately notified when an escalation occurs, **so that I** can monitor AI quality in real-time.

**Acceptance Criteria:**
- [ ] Admin receives in-app notification on escalation
- [ ] Optional email/Slack notification is configurable
- [ ] Notification includes conversation ID and reason
- [ ] Notification links to conversation detail

---

## EPIC E-004: Organization Onboarding

**Phase:** 1 (POC Core)
**Priority:** P0
**Description:** Setup and configuration for new organizations

### User Stories

#### US-E004-001: Create Organization
**As a** platform super admin, **I want to** create a new organization in the platform, **so that** I can onboard new clients.

**Acceptance Criteria:**
- [ ] Super admin can create org with name and details
- [ ] Organization ID is generated
- [ ] Initial admin user is assigned
- [ ] Org-scoped resources (S3 prefix, vector namespace) are created
- [ ] Organization appears in platform dashboard

#### US-E004-002: Organization Admin Setup
**As a** platform super admin, **I want to** invite an organization admin, **so that** they can manage their organization's settings.

**Acceptance Criteria:**
- [ ] Invitation email is sent to new admin
- [ ] Admin sets up password on first login
- [ ] Admin has org-scoped access (no cross-org visibility)
- [ ] Role is set to Organization Admin

#### US-E004-003: Organization Configuration
**As an** organization admin, **I want to** configure my organization's settings, **so that** the AI behaves according to my brand.

**Acceptance Criteria:**
- [ ] Admin can set organization name and logo
- [ ] Admin can configure AI greeting message
- [ ] Admin can select AI voice persona
- [ ] Settings are saved and apply to all customer calls

#### US-E004-004: MCP Server Configuration
**As an** organization admin, **I want to** connect MCP servers for data retrieval, **so that** the AI can access real-time information.

**Acceptance Criteria:**
- [ ] Admin can add MCP server with URL and auth
- [ ] Supported auth types: API key, OAuth, None
- [ ] Connection is tested before saving
- [ ] MCP server appears in admin settings
- [ ] Multiple MCP servers can be connected

#### US-E004-005: Organization Deactivation
**As a** platform super admin, **I want to** deactivate an organization, **so that** I can pause service for non-payment or misuse.

**Acceptance Criteria:**
- [ ] Super admin can deactivate org
- [ ] Deactivated org's customer calls return "unavailable"
- [ ] Deactivated org's data is preserved
- [ ] Super admin can reactivate org

---

## EPIC E-005: Escalation System

**Phase:** 1 (POC Core)
**Priority:** P1
**Description:** Human handoff when AI cannot answer

### User Stories

#### US-E005-001: Escalation Trigger (Agent-Centric)
**As a** system, **I want to** enable the agent to decide when it cannot answer, **so that** I can forward the customer to a human agent.

**Acceptance Criteria:**
- [ ] Agent has agency to decide when it cannot answer satisfactorily
- [ ] Agent triggers "escalate" tool based on its own reasoning (not hardcoded thresholds)
- [ ] Escalation triggers when: agent determines it can't help, MCP failure, explicit request
- [ ] Customer can say "talk to agent" to trigger escalation
- [ ] Escalation decision is made within 1 second
- [ ] Escalation type is logged for analytics

#### US-E005-002: Escalation Acknowledgment
**As a** customer, **I want to** hear an acknowledgment when I'm being transferred, **so that I** know the system is working.

**Acceptance Criteria:**
- [ ] AI says: "I understand — let me connect you with a support specialist"
- [ ] Amber "connecting" indicator is shown
- [ ] Estimated wait time is displayed if available
- [ ] Transition is smooth with no dead air

#### US-E005-003: Context Handoff
**As a** system, **I want to** preserve conversation context for the human agent, **so that** the customer doesn't have to repeat information.

**Acceptance Criteria:**
- [ ] Full transcript is preserved
- [ ] Retrieved context is included
- [ ] MCP data (if any) is included
- [ ] Escalation reason is attached
- [ ] Handoff package is sent to human agent system

#### US-E005-004: Escalation Dashboard Display
**As an** organization admin, **I want to** see escalated conversations in the dashboard, **so that I** can track AI quality and human workload.

**Acceptance Criteria:**
- [ ] Escalated conversations show red "Escalated" badge
- [ ] Escalation reason is displayed
- [ ] Handoff status (waiting, assigned, resolved) is visible
- [ ] Admin can click to view full details

#### US-E005-005: Escalation Resolution
**As a** human agent, **I want to** resolve escalated conversations and provide feedback, **so that** the loop can be closed.

**Acceptance Criteria:**
- [ ] Human agent receives handoff with full context
- [ ] Agent can chat with customer (external system)
- [ ] Agent can mark conversation as resolved
- [ ] Resolution is logged for analytics
- [ ] Resolved escalations do not affect automation rate

---

## EPIC E-006: Satisfaction Scoring

**Phase:** 1 (POC Core)
**Priority:** P1
**Description:** Two-part satisfaction scoring: automatic sentiment analysis + separate rating agent

### User Stories

#### US-E006-001: Sentiment Analysis (Automatic)
**As a** system, **I want to** analyze customer sentiment during the call, **so that** I can automatically score satisfaction.

**Acceptance Criteria:**
- [ ] Real-time voice tone analysis during the call
- [ ] Sentiment signals are captured throughout conversation
- [ ] Sentiment score is calculated automatically
- [ ] Score contributes to final satisfaction rating

#### US-E006-002: Rating Agent (Review-Based)
**As a** system, **I want to** have a separate agent review conversations and provide ratings, **so that** we have an intelligent review mechanism.

**Acceptance Criteria:**
- [ ] Rating agent reviews conversation transcripts
- [ ] Agent provides ratings based on conversation quality
- [ ] Rating is combined with sentiment analysis for final score
- [ ] Rating agent runs asynchronously after call ends

#### US-E006-003: Final Satisfaction Score
**As an** organization admin, **I want to** see comprehensive satisfaction scores, **so that** I can understand AI quality.

**Acceptance Criteria:**
- [ ] Final Score = sentiment_analysis_score + rating_agent_score
- [ ] Dashboard shows average satisfaction score
- [ ] Score trend graph shows last 7/30 days
- [ ] Low-scoring conversations are flagged
- [ ] Score threshold configuration: 85+ excellent, 70-84 good, 50-69 attention, <50 poor

---

## EPIC E-007: Text Chat Fallback

**Phase:** 1 (POC Core)
**Priority:** P1
**Description:** Text-based chat alternative to voice

### User Stories

#### US-E007-001: Text Chat Interface
**As a** customer, **I want to** type my questions instead of speaking, **so that I** can get support even without audio.

**Acceptance Criteria:**
- [ ] Text chat button is available on support page
- [ ] Chat window opens with AI greeting
- [ ] Customer can type messages
- [ ] AI responds with text (uses same RAG pipeline)
- [ ] Typing indicator shows while AI is processing

#### US-E007-002: Mode Switching
**As a** customer, **I want to** switch between voice and text during a session, **so that I** can use whichever is more convenient.

**Acceptance Criteria:**
- [ ] "Switch to Voice" button available in text chat
- [ ] "Switch to Text" button available in voice call
- [ ] Context is preserved when switching modes
- [ ] Session continues seamlessly

#### US-E007-003: Text Chat Features
**As a** customer, **I want to** have a rich chat experience, **so that I** can communicate effectively.

**Acceptance Criteria:**
- [ ] Message timestamps are shown
- [ ] Read receipts are displayed
- [ ] Quick reply suggestions are available
- [ ] File attachments can be sent
- [ ] Rich responses (links, formatting) are supported

---

## EPIC E-008: MCP Integration

**Phase:** 2
**Priority:** P1
**Description:** External data source connectivity via MCP

### User Stories

#### US-E008-001: MCP Server Management
**As an** organization admin, **I want to** add, edit, and remove MCP servers, **so that** I can connect external data sources.

**Acceptance Criteria:**
- [ ] Admin can add MCP server (URL, auth, name)
- [ ] Admin can test connection before saving
- [ ] Admin can edit MCP server configuration
- [ ] Admin can delete MCP server
- [ ] MCP server status (connected/error) is visible

#### US-E008-002: MCP Tool Discovery
**As a** system, **I want to** discover available tools from connected MCP servers, **so that** the AI knows what data it can access.

**Acceptance Criteria:**
- [ ] MCP tools are enumerated on connection
- [ ] Tools are mapped to org's available capabilities
- [ ] Tool schemas are parsed for AI consumption
- [ ] Invalid/malformed tools are flagged

#### US-E008-003: MCP Data Retrieval
**As a** system, **I want to** call MCP tools to fetch data for customer queries, **so that** AI responses include real-time information.

**Acceptance Criteria:**
- [ ] Intent classification determines when MCP is needed
- [ ] MCP call is made with appropriate parameters
- [ ] Response is parsed and validated
- [ ] Data is merged with RAG context
- [ ] MCP failure triggers escalation, not silent failure

#### US-E008-004: MCP Timeout and Retry
**As a** system, **I want to** handle MCP timeouts gracefully, **so that** customer experience is not degraded.

**Acceptance Criteria:**
- [ ] MCP requests timeout after 3 seconds
- [ ] Retries are attempted 2 times with exponential backoff
- [ ] Failed MCP calls fall back to RAG-only response
- [ ] MCP failures are logged for admin review
- [ ] Admin is notified of repeated MCP failures

#### US-E008-005: MCP Security Sandboxing
**As a** system, **I want to** prevent MCP calls from accessing unauthorized resources, **so that** tenant isolation is maintained.

**Acceptance Criteria:**
- [ ] Only configured URLs are callable
- [ ] Request size limits are enforced
- [ ] DNS rebinding protection is active
- [ ] MCP calls originate from isolated network segment
- [ ] Cross-org MCP access is impossible

---

## EPIC E-009: Question Tagging

**Phase:** 2
**Priority:** P2
**Description:** Automatic question classification and tagging

### User Stories

#### US-E009-001: Automatic Tag Classification
**As a** system, **I want to** automatically tag customer questions by topic, **so that** admins can see what customers are asking about.

**Acceptance Criteria:**
- [ ] Questions are classified into predefined tags
- [ ] Tags: Account, Billing, Product, Order, Technical, General, Unanswered
- [ ] Classification is done by LLM
- [ ] Tag is stored with conversation record

#### US-E009-002: Custom Tags
**As an** organization admin, **I want to** create custom tags specific to my business, **so that** I can categorize questions my way.

**Acceptance Criteria:**
- [ ] Admin can create new tags
- [ ] Admin can assign keywords to tags
- [ ] Admin can edit/remove custom tags
- [ ] Custom tags appear alongside predefined tags

#### US-E009-003: Tag Analytics
**As an** organization admin, **I want to** see which tags are most frequent, **so that I** can prioritize knowledge additions.

**Acceptance Criteria:**
- [ ] Analytics shows tag frequency distribution
- [ ] Tag trends show increasing/decreasing topics
- [ ] Unanswered questions are tagged separately
- [ ] Admin can drill down by tag to see examples

---

## EPIC E-010: Platform Super Admin

**Phase:** 3
**Priority:** P2
**Description:** Multi-organization management for platform provider

### User Stories

#### US-E010-001: Platform Dashboard
**As a** platform super admin, **I want to** see an overview of all organizations, **so that I** can monitor platform health.

**Acceptance Criteria:**
- [ ] Dashboard shows: Total orgs, Active calls, Platform health
- [ ] Organization list shows status and key metrics
- [ ] Alerts are displayed for issues
- [ ] Quick actions are available per org

#### US-E010-002: Organization Management
**As a** platform super admin, **I want to** manage organizations (create, edit, deactivate), **so that I** can onboard and support clients.

**Acceptance Criteria:**
- [ ] Super admin can view all orgs in list
- [ ] Super admin can create new org
- [ ] Super admin can edit org settings
- [ ] Super admin can deactivate/reactivate org
- [ ] Org detail view shows usage, performance, feedback

#### US-E010-003: Platform Feedback Collection
**As a** platform super admin, **I want to** receive and review feedback from organizations, **so that I** can improve the platform.

**Acceptance Criteria:**
- [ ] Org admins can submit feedback
- [ ] Feedback is visible in platform dashboard
- [ ] Super admin can mark feedback as reviewed
- [ ] Feedback status (new, reviewed, planned, implemented) is tracked

---

## EPIC E-011: Platform Enhancement

**Phase:** 3
**Priority:** P3
**Description:** Advanced features for future enhancement

### User Stories

#### US-E011-001: Cross-Org Intelligence
**As a** platform super admin, **I want to** see aggregated insights across organizations, **so that I** can identify platform-wide patterns.

**Acceptance Criteria:**
- [ ] Aggregated analytics show platform trends
- [ ] Common knowledge gaps are identified
- [ ] Best practices from top performers are surfaced
- [ ] Cross-org data is anonymized

#### US-E011-002: Learning Loop
**As a** system, **I want to** learn from admin resolutions, **so that** AI improves over time.

**Acceptance Criteria:**
- [ ] Admin resolutions are captured as training data
- [ ] Approved answers are indexed for future use
- [ ] Agent confidence improves based on resolution patterns
- [ ] Learning is controlled and measurable

---

## STORY METADATA TEMPLATE

```
#### US-XXX-XXX: [Story Title]
**As a** [persona], **I want to** [action], **so that** [outcome].

**Acceptance Criteria:**
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion N]

**Technical Notes:**
- [Backend/API/UI/Integration]
- [Priority: P0/P1/P2]
- [Estimated: S/M/L]
```

---

## ESTIMATION NOTES

| Size | Description |
|------|-------------|
| **S** | 1-2 days |
| **M** | 3-5 days |
| **L** | 5-10 days |
| **XL** | 10+ days (needs breakdown) |

---

*Document Status: Draft — For sprint planning and implementation*
*Next Steps: Review with team, break into sprints, prioritize stories for Phase 1*