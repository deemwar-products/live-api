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