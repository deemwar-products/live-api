---
name: error-handling-alerting
description: Error handling patterns and platform team alerting
metadata:
  type: project
---

# Error Handling & Alerting

## Error Patterns for MVP

| Pattern | Where | Details |
|---------|-------|---------|
| Retry + Backoff | All external calls | 3 retries, exponential backoff (1s, 2s, 4s) |
| Circuit Breaker | RAG, MCP tools | Trip after 5 failures, half-open after 30s |
| Fallback Messages | AI can't answer | Configurable per org |
| Timeout | Per component | RAG: 3s, MCP: 5s, configurable |
| Dead Letter Queue | Failed escalations | Queue for retry |

## Platform Alerting
**New Requirement:** Alert email to core team when critical services are not functioning.

| Trigger | Alert Action |
|---------|--------------|
| Circuit breaker tripped | Email: "Service X is failing - investigating..." |
| All services down | Email: "Critical - all services down" |
| Escalation channel failure | Email: "Escalation notifications failing" |
| High error rate | Email: "Error rate spiked for [service]" |

## Fallback Messages (Configurable per Org)
- What AI says when it can't answer
- Fallback escalation trigger

## Logging
- All errors logged for debugging
- Structured logging for observability