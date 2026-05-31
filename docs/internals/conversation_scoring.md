# Real-Time Conversation Scoring — Specification

## Overview

Every turn during a live customer conversation, a lightweight **Classifier Agent** runs in parallel to the main Gemini Live session. It evaluates the conversation state and outputs a health score. This score drives live dashboard visibility and human takeover decisions.

**Model:** Gemini 2.5 Flash or Gemini 3 Flash at temperature 0
**Trigger:** Every conversation turn (customer message + AI response)
**Output:** Score 0.0–1.0 + per-signal breakdown + suggested action

---

## Signal Taxonomy

### 1. Confidence & Grounding (Weight: 40%)

Measures whether the AI is drawing from actual knowledge or hallucinating.

| Condition | Score |
|-----------|-------|
| RAG retrieval, chunk score > 0.75 | 1.0 |
| RAG retrieval, chunk score 0.50–0.75 | 0.7 |
| RAG retrieval, chunk score < 0.50 | 0.3 |
| No RAG context available | 0.2 |
| AI explicitly says "I don't know" / "I can't help" | 0.4 |
| MCP tool call succeeds | +0.15 bonus |
| MCP tool call fails | −0.20 penalty |

**Calculation:** Average across the last 5 turns. Cap at 1.0.

---

### 2. Customer Sentiment & Frustration (Weight: 35%)

Measures whether the customer is satisfied or escalating.

| Condition | Score |
|-----------|-------|
| Explicit escalation request ("talk to a human," "frustrated") | 0.0 |
| High frustration markers ("doesn't work," "repeating," "why") | 0.1–0.3 |
| Mild skepticism ("okay but," "I'm not sure," "hmm") | 0.5 |
| Neutral / cooperative ("okay," "got it") | 0.7 |
| Satisfied / positive ("thanks," "perfect," "that helps") | 0.95 |

**Calculation:** Score the last customer message + the one before it. Average them.
**Override:** If explicit escalation request detected, sentiment score = 0.0 immediately, regardless of average.

---

### 3. Progress Toward Resolution (Weight: 15%)

Measures whether the conversation is making headway.

| Turn Count | Base Score |
|------------|------------|
| Turn 1–3 | 1.0 |
| Turn 4–5 | 0.8 |
| Turn 6–8 | 0.6 |
| Turn 9–12 | 0.4 |
| Turn 13+ | 0.2 |

**Penalties:**

| Condition | Adjustment |
|-----------|------------|
| Customer asks the same question twice | −0.30 |
| AI contradicts itself | −0.20 |
| Customer says "I already told you" | −0.40 |

**Bonuses:**

| Condition | Adjustment |
|-----------|------------|
| Customer asks a follow-up building on prior answer | +0.10 |
| MCP tool call returns actionable result | +0.20 |

---

### 4. Knowledge Gap Risk (Weight: 10%)

Measures likelihood the LLM Judge will flag this conversation post-session.

| Condition | Score |
|-----------|-------|
| All RAG chunks scored < 0.50 | 0.1 |
| MCP tool called but returned error | 0.2 |
| AI responds despite low confidence | 0.3 |
| RAG context found and used | 0.9 |
| Question matches known FAQ / common issue | 0.95 |

**Calculation:** Average of last 3 retrieved chunk relevance scores, capped at 1.0.

---

## Overall Score Calculation

```
score = (0.40 × confidence) + (0.35 × sentiment) + (0.15 × progress) + (0.10 × kg_risk)
```

Result range: **0.0 – 1.0**

---

## Threshold Tiers

Default thresholds (org-configurable):

| Range | Status | Action |
|-------|--------|--------|
| 0.70 – 1.00 | 🟢 Green | Normal. Continue. |
| 0.50 – 0.69 | 🟡 Yellow | Monitor. Flag on dashboard. Optional admin alert. |
| 0.30 – 0.49 | 🟠 Orange | At risk. Dashboard notification. Suggest takeover. |
| 0.00 – 0.29 | 🔴 Red | Critical. Auto-notify admin. Immediate takeover recommended. |

**Org-configurable overrides:**
- Red threshold: 0.2 (permissive) to 0.5 (aggressive)
- Auto-escalate on Red: toggle on/off per org
- Yellow/Orange thresholds: adjustable within platform bounds

---

## Worked Example

**Context:** Turn 7, customer says "this still isn't working."

| Signal | Value | Weighted |
|--------|-------|---------|
| Confidence (RAG hit 0.8, no MCP) | 0.80 | 0.40 × 0.80 = 0.32 |
| Sentiment ("still isn't working") | 0.20 | 0.35 × 0.20 = 0.07 |
| Progress (Turn 7, no resolution) | 0.50 | 0.15 × 0.50 = 0.075 |
| KG Risk (chunks scored 0.6–0.7) | 0.65 | 0.10 × 0.65 = 0.065 |
| **Total** | | **0.53 → 🟡 Yellow** |

Dashboard flags the conversation. Admin can monitor or initiate takeover.

---

## Human Takeover Flow

When a conversation is flagged (Orange or Red), or admin manually decides to intervene:

```
1. Admin sees flagged conversation on dashboard
   - Current score + which signal tanked
   - Last 3 turns preview
   - Suggested reason ("Customer frustrated," "AI out of knowledge," etc.)

2. Admin clicks "Take Over"

3. System:
   a. Pauses Gemini Live session gracefully
   b. Persists full conversation to PostgreSQL immediately
   c. Bundles classifier signals + context for the agent

4. Customer hears:
   "Thanks for your patience. A support specialist has reviewed our
    conversation and would like to help. One moment please..."

5. Human agent receives:
   - Full conversation transcript (read-only)
   - Per-signal score breakdown
   - Classifier's flagged reason
   - Session duration + turn count
   - Any knowledge gaps or failed tool calls

6. Agent resolves issue
   - Resolved → session marked resolved_human
   - Further escalation needed → marked escalated
```

---

## Dashboard Real-Time View

Live conversations panel shows per-conversation:

| Field | Description |
|-------|-------------|
| Session ID | Unique identifier |
| Customer | Identifier if provided by org |
| Mode | Voice / Chat |
| Turn Count | Current turn number |
| Current Score | 0.0–1.0 with color indicator |
| Score Trend | Arrow: improving / stable / declining |
| Top Signal | Which signal is lowest right now |
| Status | Active / Flagged / Takeover Pending |

Admins can click any row to see the full live transcript + per-signal breakdown in real time.

---

*This section to be incorporated into TRD under a new section: **Section 19: Real-Time Conversation Scoring & Live Intervention***
