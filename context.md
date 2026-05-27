# Session Context — live-api-poc

## Session Start
- **Date:** 2026-05-24
- **User:** Sreyash Reddy

## Project Info
- **Location:** `/mnt/c/Users/Sreyash/Desktop/Codebases/poc/live-api-poc`
- **Status:** Initial session — no prior context loaded

## What's Been Discussed
1. **Workflow rules:**
   - Maintain `context.md` — captures all session discussions and decisions
   - Maintain `learnings.md` — captures knowledge from online searches and discoveries
2. **PRD/TRD generation:**
   - Use `bmad-party-mode` skill
   - Include personas: Researcher, Backend Dev Friend, Developer, Architect, Project Manager (+ others as needed)

## Project Vision — Gemini Live API Call Platform (Customer Support Automation)

### Core Idea
A B2B SaaS platform for **customer support automation** — replacing human agents entirely with AI voice agents that can:
1. Answer queries using organization-specific knowledge (RAG)
2. Take actions via connected systems (MCP) — book appointments, update tickets, query databases, etc.
3. Provide a natural voice conversation experience via Gemini Live API

### Key Decisions (2026-05-24)
- **Target persona**: Open to all industries — industry-agnostic platform
- **Primary use case**: Customer support automation — replace human agents with AI voice
- **Human in loop**: No — fully autonomous agent handles calls independently
- **MCP scope**: Data retrieval + basic write operations (no human supervision needed)
- **Fallback UI**: "Call Agent" button in UI for manual help — workaround, not primary flow
- **Core value proposition**: Voice-first interaction — real-time, low-latency, interactive (Gemini Live API)
- **Text alternative**: Will have text chat as well, but voice is primary feature
- **Agent persona**: Friendly, trust-building, makes customers feel their problems will be resolved
- **Voice configuration**: Via Gemini Live API settings
- **MVP approach (POC phase)**: RAG first, then MCP integration
- **Final product**: Both RAG + MCP from day one

### Role-Based Access Control (RBAC)
- **Super Admin** = Platform provider (us)
- **Organization Admin** = Per organization — can add docs, manage data sources, view analytics
- **Other roles** = TBD per organization needs

### Feedback Loop Architecture (NEW)

#### User → Organization (Customer Satisfaction Tracking)
- Conversation transcripts visible to organization admins
- Each conversation scored for customer satisfaction
- Dissatisfaction analysis: identify unanswered questions, missing data
- Question tagging: frequently asked topics surfaced to admin
- Admins can add documents/data sources to fill knowledge gaps

#### Escalation Loop (NEW)
- When agent can't answer → immediately forwarded to human agent
- Question flagged as "unanswered" and displayed in dashboard
- Admin notified immediately of unanswered question
- Admin turnaround time: depends on admin (not under platform control)

#### Admin Win Condition (NEW)
- **Primary metric**: Maximize queries resolved without human intervention
- **Success =**: High automation rate = low human intervention = AI doing its job well

#### Organization → Super Admin (Platform Feedback)
- Organizations provide feedback directly to platform provider
- **No direct customer → super admin channel** — customers only interact with org, not with us
- Feedback flow: Customer → Org Admin → Super Admin (us)

### Open Questions (RESOLVED)
- **Escalation trigger**: Agent-centric — agent decides and triggers escalation tool itself
- **Satisfaction scoring**: Two-part system (sentiment analysis + rating agent)
- **Target automation rate**: 80-90% (feasible goal)
- **Voice greeting**: Basic configuration, decide later

### Multi-Tenant Architecture
- Each organization gets isolated dashboard
- Document storage, MCP connections, embeddings — all per-org with strict isolation
- No cross-org data leakage possible (architectural requirement, not optional)

### Current State
- PoC stage — product vision fully defined
- Team has provided multi-perspective feedback (Mary, Winston, John, Amelia, Victor, Sally)
- **All documents created and organized in `docs/` folder**

### Document Structure
```
docs/
├── _architecture/     # Project-related docs (PRD, TRD, architecture, epics)
│   ├── prd.md
│   ├── voice-ai-platform-architecture.md
│   ├── ux-design-specification.md
│   └── epics-and-stories.md
├── _learnings/        # LLM context-based learnings for agent consistency
│   └── (learnings accumulated during development)
└── learnings.md      # Session learnings
```

### What's Next
- Review documents with user
- Implementation details deferred to build phase