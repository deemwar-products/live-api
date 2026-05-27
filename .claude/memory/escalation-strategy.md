---
name: escalation-strategy
description: Toggle-based escalation - live vs async messaging
metadata:
  type: project
---

# Escalation Strategy

## Decision
Per-organization escalation toggle with two modes:

### Live Mode (Toggle ON)
- AI escalates to available human agent in real-time
- Voice: WebRTC transfer / Call handoff
- Chat: Chat handoff to agent dashboard
- Agent must be online/available

### Async Mode (Toggle OFF)
- AI captures the unresolved query
- Sends notification to org via messaging channel (SMS, WhatsApp, etc.)
- Org responds when available
- No live agent required

## Rationale
Serves both small companies (no active staff needed) and large companies (live support desk).

## Technical Implications
- AI needs to read org's escalation toggle setting
- AI needs to check agent availability (presence system)
- Async mode requires: messaging channel config, message queue/webhook
- MCP tool for sending messages
- SLA expectation settings for async mode (optional)

## Open Questions
- Which messaging channels to support first? (SMS, WhatsApp, Email)
- Should async notifications include AI-summarized context or just raw transcript?