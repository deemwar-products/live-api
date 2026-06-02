---
name: feedback-judge-system
description: LLM-powered judge for conversation quality assessment and feedback
metadata:
  type: project
---

# LLM as Judge - Feedback System

## Purpose
An LLM-based agent that evaluates conversation quality and generates actionable feedback for businesses.

## Functionality

### Conversation Evaluation
- Judge whether the AI agent answered the customer's query correctly
- Identify what went wrong when answer was incorrect/incomplete
- Assess overall conversation flow and customer satisfaction signals

### Feedback Generation
- Generate structured feedback with context (what was asked, what was answered, why it failed)
- Flag knowledge gaps in the organization's knowledge base
- Help businesses identify what content needs to be added/updated

## Use Case
End of day or in real-time: Business owner receives feedback
→ "Customer asked about X, your AI couldn't answer"
→ Business updates their knowledge base
→ Better customer experience next time

## Technical Note
This is an AI-powered quality assurance layer that sits alongside the conversation and generates insights for continuous improvement.