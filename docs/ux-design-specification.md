---
title: Voice AI Customer Support Platform — UX Design Specification
status: draft
created: 2026-05-24
updated: 2026-05-24
author: Sreyash Reddy
source_prd: docs/prd.md
source_architecture: docs/voice-ai-platform-architecture.md
---

# Voice AI Customer Support Platform — UX Design Specification

## Executive Summary

This document provides comprehensive UX design specifications for the Voice AI Customer Support Platform PoC. It addresses the UX gaps identified during PRD review and provides concrete design decisions for the customer voice interface, admin dashboard, and escalation flows.

---

## 1. Design Principles

### 1.1 Core Design Philosophy

| Principle | Application |
|-----------|-------------|
| **Trust through transparency** | Users always know what's happening — no silent states, no confusing transitions |
| **Voice-first, not voice-only** | Voice is primary, but text is always available as fallback |
| **Human warmth** | AI should feel helpful and capable, not robotic or cold |
| **Clarity over cleverness** | When something goes wrong, users should know exactly what to do |
| **Progressive disclosure** | Show what matters now; reveal complexity on demand |

### 1.2 Design Language

| Element | Specification |
|---------|---------------|
| **Primary Color** | #4F46E5 (Indigo) — trustworthy, professional |
| **Secondary Color** | #10B981 (Emerald) — success, resolution |
| **Accent Color** | #F59E0B (Amber) — attention, escalation |
| **Error Color** | #EF4444 (Red) — problems, failures |
| **Background** | #F9FAFB (Light gray) — clean, minimal |
| **Text Primary** | #111827 (Near black) — high readability |
| **Text Secondary** | #6B7280 (Gray) — supporting information |
| **Border Radius** | 12px — friendly, modern |
| **Shadows** | Subtle (0 2px 4px rgba(0,0,0,0.1)) — depth without heaviness |
| **Typography** | Inter (headings), System font (body) — clean, accessible |

### 1.3 Accessibility Requirements

| Requirement | Standard |
|-------------|----------|
| **Color contrast** | WCAG 2.1 AA (4.5:1 minimum) |
| **Touch targets** | 44px minimum |
| **Focus indicators** | Visible on all interactive elements |
| **Screen reader** | Full ARIA labeling |
| **Motion** | Respects reduced-motion preferences |

---

## 2. User Personas

### 2.1 Customer (End User)

**Profile:** Contacting support via voice or text chat
**Goals:** Get quick, accurate answers; feel heard and respected
**Pain Points:** Long wait times, repeating themselves, unclear next steps
**Emotional State:** Usually frustrated or seeking help

### 2.2 Organization Admin

**Profile:** Managing support operations for their organization
**Goals:** Maintain high automation rate, identify knowledge gaps, improve customer satisfaction
**Pain Points:** Unclear analytics, time-consuming manual work, unknown AI failures
**Emotional State:** Seeking control and visibility

### 2.3 Platform Super Admin

**Profile:** Managing multiple organizations on the platform
**Goals:** Onboard new clients, monitor platform health, collect feedback
**Pain States:** Juggling multiple priorities, need quick overview of issues
**Emotional State:** Strategic, Birds-eye view

---

## 3. Customer Voice Interface

### 3.1 Screen Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  [Org Logo]                              [?] [End Call]         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│                     ┌─────────────────┐                         │
│                     │                 │                         │
│                     │   AI AGENT      │                         │
│                     │   AVATAR        │                         │
│                     │   "Listening"   │                         │
│                     │                 │                         │
│                     └─────────────────┘                         │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  🔊 ▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░  [Processing...]   │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Transcript Area:                                               │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Customer: I need help with my order                      │  │
│  │ AI: I can help you with that. Let me check your order    │  │
│  │ status... Your order #12345 is currently being shipped.  │  │
│  │ Expected delivery: tomorrow by 5 PM.                     │  │
│  └───────────────────────────────────────────────────────────┘  │
│                                                                  │
├─────────────────────────────────────────────────────────────────┤
│  [🎤 Voice]  [💬 Text]              [📞 Call Agent]           │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Conversation States

| State | Visual Indicator | Audio Indicator | Duration |
|-------|-----------------|------------------|----------|
| **Idle** | Avatar static, muted | None | Until customer initiates |
| **Listening** | Avatar with animated pulse, waveform active | "I'm listening..." indicator | 0 → 30s |
| **Thinking** | Avatar thinking (dots animation) | Soft chime | 0.5s → 5s |
| **Speaking** | Avatar speaking (animated), response text appearing | TTS audio | Variable |
| **Interrupting** | Avatar paused, "interrupted" indicator | None | Until customer speaks |
| **Escalating** | Avatar with loading indicator | "Let me connect you..." | 2s → 10s |
| **Transferred** | "Human Agent" badge, chat interface shown | None | Until resolution |

### 3.3 Voice Conversation UX Details

#### 3.3.1 Processing States (Critical)

```
STATE: "I'm looking into that for you..."
  - Show animated thinking dots
  - Display "Searching knowledge base..." if RAG active
  - Display "Checking external systems..." if MCP active
  - Never show blank screen for > 1 second

STATE: "Let me connect you with a specialist..."
  - Show amber "connecting" badge
  - Display estimated wait time if available
  - Show human agent photo/name when assigned
```

#### 3.3.2 Error States

| Error | Customer UI | Recovery Action |
|-------|-------------|-----------------|
| **Microphone denied** | "Microphone access needed" modal with instructions | "Open Settings" button |
| **Connection lost** | "Reconnecting..." banner with spinner | Auto-retry 3x |
| **AI failure** | "Experiencing technical difficulties" with retry | "Try Again" button |
| **Long wait (>30s)** | "Your support specialist is on the way" | Shows position in queue |

### 3.4 Voice Greeting Script

```
[AI]: "Hi there! I'm [Org Name]'s support assistant. 
       I can help you with questions about your orders, 
       account, and more. What can I help you with today?"
```

- Greeting is configurable per organization
- If customer is returning (detected via session), personalize:
  ```
  [AI]: "Welcome back! How can I help you today?"
  ```

### 3.5 Interruption Handling

```
Customer speaks while AI is responding:
  → AI stops speaking immediately
  → Display "Customer interrupted" briefly (1s)
  → Show customer input in transcript
  → AI responds to new input
  → If interruption is meaningful, acknowledge: 
    "Got it — let me help with that instead."
```

### 3.6 Conversation End Flow

```
TRIGGER: Customer says "thanks", "bye", or stays silent for 10s

AI Response: 
  "Thanks for calling! Is there anything else I can help with?"
  
  If no response for 5s:
  "Take care! Feel free to call back if you need anything else. 
   Goodbye!"

CONVERSATION ENDS
      │
      ▼
┌─────────────────────────────────────────┐
│         SATISFACTION COLLECTION          │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │  How was your experience today?   │  │
│  │                                   │  │
│  │   😞    😐    🙂    😊    🤩    │  │
│  │   1      2     3      4     5    │  │
│  │                                   │  │
│  │  [Skip this question]            │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

**Scoring UX Decisions:**
- Emoji scale is more friendly than stars for voice context
- "Skip" option increases completion rate
- Post-call survey is optional (not blocking)

---

## 4. Escalation UX Design

### 4.1 Escalation Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     ESCALATION FLOW                             │
│                                                                 │
│  Customer: "I need to speak to a real person"                  │
│            OR                                                    │
│  AI cannot answer (confidence < threshold)                     │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  STEP 1: Acknowledgment                                         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  [AI]: "I understand — let me connect you with a       │   │
│  │        support specialist who can help."               │   │
│  │                                                     │   │
│  │  ⏳ Connecting to human agent...                      │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  STEP 2: Context Handoff                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  [AI]: "I've shared your conversation details with    │   │
│  │        the agent. You won't need to repeat anything." │   │
│  │                                                     │   │
│  │  ✓ Your context has been forwarded                  │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  STEP 3: Wait / Assign                                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  [If waiting]:                                        │   │
│  │  "Please hold — a specialist will be with you shortly"│   │
│  │  Estimated wait: ~2 minutes                           │   │
│  │                                                     │   │
│  │  [If assigned]:                                      │   │
│  │  "Sarah from our support team is joining now."      │   │
│  │  [Photo] Sarah M. — Support Specialist              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Escalation UI Components

| Component | Description |
|-----------|-------------|
| **Connecting Badge** | Amber pulsing circle with "Connecting..." text |
| **Context Confirmation** | Green checkmark with "Context shared" |
| **Wait Estimate** | "Estimated wait: ~X minutes" (if available) |
| **Agent Card** | Photo, name, title when assigned |
| **Transfer Complete** | "You're now connected with [Agent Name]" |

### 4.3 Escalation States in Dashboard

```
┌─────────────────────────────────────────────────────────────────┐
│  ESCALATION BADGE (appears on affected conversation)            │
│                                                                 │
│  🔴 Escalated: No relevant context                              │
│     Reason: AI couldn't find information in knowledge base     │
│     Time: 2:34 PM                                               │
│     [View Transcript] [Chat with Agent]                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. Text Chat Interface

### 5.1 Screen Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  [Org Logo]  Chat Support           [End Chat] [📞 Voice]        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  AI: Hi! I'm here to help. What can I assist you with    │  │
│  │     today?                                               │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│                          ▼                                      │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Customer: I need to track my order                     │  │
│  │  Sent: 2:31 PM                                          │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│                          ▼                                      │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  AI: I can help you track your order. Could you provide │  │
│  │     your order number or the email used for the order?  │  │
│  │  Sent: 2:31 PM                                          │  │
│  └───────────────────────────────────────────────────────────┘  │
│  │                                           │                  │
│  │  [Typing indicator: Agent is typing...]  │                  │
│  └───────────────────────────────────────────┘                  │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────┐ [Send]        │
│  │ Type your message...                        │               │
│  └─────────────────────────────────────────────┘               │
│                                                                 │
│  [📎 Attach] [📞 Call Agent]                    [💬 Voice Mode]  │
└─────────────────────────────────────────────────────────────────┘
```

### 5.2 Chat Features

| Feature | Description |
|---------|-------------|
| **Typing Indicator** | Shows when AI is processing/typing |
| **Read Receipts** | "Seen" status for customer messages |
| **File Attachments** | Support for screenshots (for reference) |
| **Rich Responses** | AI can include links, formatted text |
| **Quick Replies** | Suggested responses for common questions |

### 5.3 Mode Switching (Voice ↔ Text)

```
During Voice Call → Customer clicks [💬 Text]:
  → Audio pauses (not ended)
  → Chat interface opens
  → Previous transcript carries over
  → Customer can type additional context
  → AI responds in text
  → Customer can return to voice anytime

During Text Chat → Customer clicks [🎤 Voice]:
  → Text chat pauses (not ended)
  → Voice interface opens
  → Previous context is preserved
  → Full voice interaction begins
```

---

## 6. Organization Admin Dashboard

### 6.1 Dashboard Overview (Default View)

```
┌─────────────────────────────────────────────────────────────────────────┐
│  [Logo] Acme Corp Support          [🔔 3] [Settings] [Profile]           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐        │
│  │   AUTOMATION    │  │  SATISFACTION   │  │   ESCALATIONS   │        │
│  │     RATE        │  │     SCORE       │  │     TODAY       │        │
│  │                 │  │                 │  │                 │        │
│  │    87.3%        │  │    82.5         │  │      12         │        │
│  │   ▲ 2.1%        │  │   ▼ 1.2%        │  │   ▼ 5 from avg  │        │
│  │                 │  │                 │  │                 │        │
│  │  [Trend Graph]  │  │  [Trend Graph]  │  │  [View All]     │        │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘        │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  KNOWLEDGE GAPS                           [+ Add Document]     │       │
│  │                                                               │       │
│  │  🔴 Return Policy Questions         23 unanswered  ▲ 12%    │       │
│  │  🟡 Technical Troubleshooting       18 unanswered  ▲ 5%     │       │
│  │  🟡 Account Merging Process         11 unanswered  NEW      │       │
│  │  🟢 Shipping International          4 unanswered  ▼        │       │
│  │                                                               │       │
│  │  [View All Gaps]                                              │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  RECENT CONVERSATIONS                                         │       │
│  │                                                               │       │
│  │  🔴 Conv #4521 — Escalated — "Can't cancel my order"  2:30PM │       │
│  │  🟡 Conv #4520 — Score: 62 — Low satisfaction        2:15PM  │       │
│  │  🟢 Conv #4519 — Score: 95 — Resolved perfectly    1:50PM  │       │
│  │  🟢 Conv #4518 — Score: 88 — Resolved                1:30PM  │       │
│  │                                                               │       │
│  │  [View All Conversations]                                    │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.2 Conversation Detail View

```
┌─────────────────────────────────────────────────────────────────────────┐
│  [← Back to Dashboard]           Conversation #4521            [⋮ Menu] │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CONVERSATION TRANSCRIPT                                        │   │
│  │                                                                 │   │
│  │  Customer: Hi, I need help canceling my order                  │   │
│  │  AI: I can help you with that! What's your order number?      │   │
│  │  Customer: It's order #78956                                    │   │
│  │  AI: I found your order. However, I don't have access to      │   │
│  │       cancellation permissions in my current capabilities.    │   │
│  │  AI: Let me connect you with a support specialist who can      │   │
│  │       help with this request.                                  │   │
│  │                                                                 │   │
│  │  ───────────────────────────────────────────────────────────   │   │
│  │  ESCALATION REASON: No relevant context for order cancellation │   │
│  │  TIME: 2:30 PM | DURATION: 1:45 | SCORE: 45 (Escalated)        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  QUESTION ANALYSIS                                              │   │
│  │                                                                 │   │
│  │  Tags: [Order] [Cancellation] [Account Access]                │   │
│  │  Knowledge Gap: "Order cancellation process" — NO DOCUMENT     │   │
│  │                                                                 │   │
│  │  Suggested Action:                                             │   │
│  │  [+ Add Document] Create/Upload document about cancellation    │   │
│  │                   policy and agent permissions                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  [Add Document] [Mark as Reviewed] [Contact Customer]                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.3 Document Management View

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Documents                              [+ Upload] [Filter ▼] [Search] │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  📄 order-cancellation-policy.pdf                              │       │
│  │     Uploaded: May 20, 2026 | Chunks: 45 | Status: Active      │       │
│  │     Used in: 234 responses this month                         │       │
│  │     [View] [Edit] [Delete]                                    │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  📄 return-policy-2026.pdf                                     │       │
│  │     Uploaded: May 15, 2026 | Chunks: 38 | Status: Active      │       │
│  │     Used in: 189 responses this month                         │       │
│  │     [View] [Edit] [Delete]                                    │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  📄 shipping-information.pdf              ⚠️ Low usage         │       │
│  │     Uploaded: Apr 10, 2026 | Chunks: 22 | Status: Active      │       │
│  │     Used in: 12 responses this month (down from 89)          │       │
│  │     [View] [Edit] [Update] [Delete]                           │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  📄 account-faqs.pdf                     ⚠️ Needs attention   │       │
│  │     Uploaded: Mar 5, 2026 | Chunks: 15 | Status: Outdated    │       │
│  │     23 unanswered questions this month                        │       │
│  │     [Update] [Delete] [View Gaps]                            │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
│  [+ Upload New Document]                                                │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.4 Document Upload Flow

```
STEP 1: Select File
┌─────────────────────────────────────────────────────────────────┐
│  [Drag and drop files here, or click to browse]                 │
│                                                                 │
│  Supported: PDF, DOCX, TXT, MD, HTML                          │
│  Max size: 50MB per file                                        │
└─────────────────────────────────────────────────────────────────┘

STEP 2: Document Metadata
┌─────────────────────────────────────────────────────────────────┐
│  Document Name: [________________________]                       │
│                                                                 │
│  Category: [Select ▼]                                           │
│  - Policies                                                    │
│  - FAQs                                                        │
│  - How-to Guides                                               │
│  - Product Info                                                │
│  - Custom                                                      │
│                                                                 │
│  Tags: [Add tags separated by commas]                          │
│                                                                 │
│  ☐ Make this document the primary knowledge source             │
│  ☐ Auto-update embeddings when document changes               │
└─────────────────────────────────────────────────────────────────┘

STEP 3: Processing
┌─────────────────────────────────────────────────────────────────┐
│  Processing document...                                         │
│                                                                 │
│  ✓ File validated                                              │
│  ████████░░░░ Chunking (45 chunks)                             │
│  ░░░░░░░░░░░ Embedding generation                              │
│  ░░░░░░░░░░░ Indexing in knowledge base                       │
│                                                                 │
│  This usually takes 1-3 minutes.                                │
└─────────────────────────────────────────────────────────────────┘

STEP 4: Complete
┌─────────────────────────────────────────────────────────────────┐
│  ✓ Document Ready!                                              │
│                                                                 │
│  "order-cancellation-policy.pdf" is now active in your        │
│  knowledge base. It will be used to answer customer questions.  │
│                                                                 │
│  [View Document] [Add Another] [Go to Dashboard]              │
└─────────────────────────────────────────────────────────────────┘
```

### 6.5 Analytics Deep Dive

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Analytics                          [This Week ▼] [Export ▼]            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CONVERSATION TRENDS                                            │   │
│  │                                                                 │   │
│  │  150 ─┐                                                        │   │
│  │      │      ┌───┐                                              │   │
│  │  100 ─┤  ┌──┘   └──┐     ┌───┐                               │   │
│  │      │  └──┐       └──┐──┘   └──┐      Current: 127/day      │   │
│  │   50 ─┤     └──┐           └──┐                                │   │
│  │      │        └──┐              └──┐                           │   │
│  │    0 ─┴─────────┴────────────────┴──                          │   │
│  │       Mon   Tue   Wed   Thu   Fri   Sat   Sun                  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────┐          │
│  │  TOP QUESTION TAGS                          [Last 30 days] │          │
│  │                                                               │          │
│  │  Account Access     ████████████████████  1,234 (23%)        │          │
│  │  Order Status      ████████████████      892 (17%)          │          │
│  │  Returns          ██████████████        678 (13%)          │          │
│  │  Technical         ████████████          567 (11%)          │          │
│  │  Billing          ██████████            456 (9%)           │          │
│  │  Other            ████████████████████  1,023 (27%)        │          │
│  └─────────────────────────────────────────────────────────────┘          │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────┐          │
│  │  AUTOMATION FUNNEL                                          │          │
│  │                                                               │          │
│  │  Total Calls: 2,456                                         │          │
│  │  └── Resolved by AI: 2,145 (87%)                           │          │
│  │       └── Resolved without hint: 1,876 (76%)               │          │
│  │  └── Escalated: 311 (13%)                                  │          │
│  │       └── Resolved by Human: 287                           │          │
│  │       └── Pending: 24                                       │          │
│  └─────────────────────────────────────────────────────────────┘          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Super Admin Dashboard

### 7.1 Platform Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Platform Dashboard              [Alerts: 2] [Settings] [Profile]         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐          │
│  │  TOTAL          │  │  ACTIVE         │  │  PLATFORM       │          │
│  │  ORGANIZATIONS │  │  CALLS (NOW)    │  │  HEALTH         │          │
│  │                 │  │                 │  │                 │          │
│  │     12          │  │     47          │  │   ✅ Healthy    │          │
│  │   +2 this month │  │   ▲ 15% vs avg  │  │                 │          │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘          │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  ORGANIZATIONS                                                │       │
│  │                                                               │       │
│  │  🟢 Acme Corp          Automation: 87%   Calls: 234/day      │       │
│  │  🟢 Beta Inc           Automation: 92%   Calls: 156/day      │       │
│  │  🟡 Gamma LLC         Automation: 71%   Calls: 89/day       │       │
│  │  🔴 Delta Co          Automation: 45%   Calls: 12/day       │       │
│  │     ⚠️ Low automation — knowledge gaps detected             │       │
│  │                                                               │       │
│  │  [Manage Organization] [Add Organization]                     │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐       │
│  │  RECENT ALERTS                                                │       │
│  │                                                               │       │
│  │  🔴 Delta Co — Automation dropped below 50%        1h ago      │       │
│  │  🟡 Acme Corp — MCP server connection unstable   3h ago      │       │
│  │  🟢 Gamma LLC — Documents indexed successfully   5h ago       │       │
│  └───────────────────────────────────────────────────────────────┘       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 7.2 Organization Management

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Acme Corp                              [Edit] [Settings] [Deactivate]  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DETAILS                                                               │
│  ──────────────────────────────────────────────────────────────────── │
│  Organization ID: org_8x7f2k                                           │
│  Plan: Professional                                                    │
│  Created: January 15, 2026                                             │
│  Admin: john@acme.com                                                  │
│                                                                          │
│  USAGE                                                                 │
│  ──────────────────────────────────────────────────────────────────── │
│  API Calls (This Month): 45,678 / 100,000                             │
│  Documents: 23 active                                                  │
│  MCP Connections: 2 configured                                         │
│                                                                          │
│  PERFORMANCE                                                            │
│  ──────────────────────────────────────────────────────────────────── │
│  Automation Rate: 87.3%                                                │
│  Avg Satisfaction: 82.5                                               │
│  Escalations (30d): 156                                                │
│                                                                          │
│  FEEDBACK                                                              │
│  ──────────────────────────────────────────────────────────────────── │
│  Latest Feedback: "Would be great to have Slack integration"          │
│  Received: May 20, 2026                                                │
│  Status: Reviewed — Planned for Q3                                      │
│                                                                          │
│  [View Dashboard] [Manage Documents] [Manage MCP] [View Feedback]       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Common Component Library

### 8.1 Buttons

| Variant | Use Case | Style |
|---------|----------|-------|
| **Primary** | Main actions (Submit, Save) | Solid indigo background |
| **Secondary** | Alternative actions (Cancel, Back) | White with border |
| **Danger** | Destructive actions (Delete) | Red background |
| **Ghost** | Tertiary actions (View, Edit) | No background |
| **Icon** | Compact actions (Settings, Menu) | Icon only, 44px touch target |

### 8.2 Input Fields

| Type | Use Case | States |
|------|----------|--------|
| **Text Input** | General text entry | Default, Focus, Error, Disabled |
| **Textarea** | Multi-line text | Default, Focus, Error, Disabled |
| **Select** | Dropdown selection | Default, Open, Selected, Disabled |
| **File Upload** | Document upload | Default, Dragging, Uploading, Complete |
| **Search** | Search with icon | Default, Focus, With Results |

### 8.3 Cards

| Type | Use Case | Content |
|------|----------|---------|
| **Metric Card** | Dashboard KPIs | Large number, label, trend indicator |
| **Conversation Card** | List items | Status, preview, timestamp, score |
| **Document Card** | Document list | Icon, name, metadata, actions |
| **Alert Card** | Notifications | Icon, message, timestamp, actions |

### 8.4 Badges & Status Indicators

| Badge | Color | Meaning |
|-------|-------|---------|
| **Resolved** | Green | Successfully handled |
| **Escalated** | Red | Requires human attention |
| **Pending** | Amber | Awaiting action |
| **Active** | Blue | Currently in progress |
| **Draft** | Gray | Not yet published |

### 8.5 Loading States

| State | Component | Description |
|-------|-----------|-------------|
| **Skeleton** | Cards, Lists | Gray animated placeholder |
| **Spinner** | Buttons, Actions | Circular loading indicator |
| **Progress Bar** | Upload, Processing | Linear progress with percentage |
| **Pulse** | Real-time data | Subtle animation for live updates |

### 8.6 Empty States

```
NO CONVERSATIONS YET
────────────────────
📞 Your first customer conversation will appear here.

[Invite Customers] [Learn More]

NO DOCUMENTS
───────────
📄 Upload your first document to train the AI agent.

[Upload Document] [View Tutorial]

NO ESCALATIONS
──────────────
✅ Great job! No escalations in the last 24 hours.

[Review Analytics]
```

---

## 9. User Flows

### 9.1 Customer Support Call Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CUSTOMER CALL FLOW                             │
│                                                                         │
│  START                                                                │
│  │                                                                    │
│  ▼                                                                    │
│  ┌─────────────────┐                                                 │
│  │ Customer opens  │                                                 │
│  │ support widget  │                                                 │
│  └────────┬────────┘                                                 │
│           │                                                           │
│           ▼                                                           │
│  ┌─────────────────┐                                                 │
│  │ Voice greeting  │                                                 │
│  │ "Hi, I'm here   │                                                 │
│  │ to help..."    │                                                 │
│  └────────┬────────┘                                                 │
│           │                                                           │
│           ▼                                                           │
│  ┌─────────────────┐                                                 │
│  │ Customer speaks  │─────────────────────────────────────────────┐   │
│  │ query           │                                             │   │
│  └────────┬────────┘                                             │   │
│           │                                                       │   │
│           ▼                                                       │   │
│  ┌─────────────────┐         ┌─────────────────┐                  │   │
│  │ AI processes    │──Yes──▶ │ AI responds     │                  │   │
│  │ Can answer?      │         │ with context   │                  │   │
│  └────────┬────────┘         └────────┬────────┘                  │   │
│           │No                         │                            │   │
│           ▼                            │                            │   │
│  ┌─────────────────┐                  │                            │   │
│  │ Check MCP       │──────────────▶   │                            │   │
│  │ for data?      │                  │                            │   │
│  └────────┬────────┘                  │                            │   │
│           │                            │                            │   │
│     ┌─────┴─────┐                     │                            │   │
│     ▼           ▼                     │                            │   │
│  ┌───────┐  ┌───────────┐             │                            │   │
│  │ MCP   │  │ Escalate  │◀────────────┘                            │   │
│  │ failed │  │ to human  │                                          │   │
│  └───┬───┘  └─────┬─────┘                                          │   │
│      │            │                                                │   │
│      ▼            ▼                                                │   │
│  ┌───────────┐  ┌───────────┐                                      │   │
│  │ RAG-only  │  │ Wait for  │                                      │   │
│  │ response │  │ human     │                                      │   │
│  └─────┬─────┘  └─────┬─────┘                                      │   │
│        │              │                                             │   │
│        └──────┬───────┘                                             │   │
│               ▼                                                      │   │
│  ┌─────────────────┐                                                │   │
│  │ Customer        │                                                │   │
│  │ satisfied?      │                                                │   │
│  └────────┬────────┘                                                │   │
│           │                                                         │   │
│     ┌─────┴─────┐                                                   │   │
│     ▼           ▼                                                   │   │
│  ┌───────┐  ┌───────────┐                                           │   │
│  │ Score │  │ Escalate │                                           │   │
│  │ call  │  │ resolved │                                           │   │
│  └───┬───┘  └─────┬─────┘                                           │   │
│      │            │                                                │   │
│      └──────┬──────┘                                                │   │
│             ▼                                                       │   │
│        END / FEEDBACK                                               │   │
└─────────────────────────────────────────────────────────────────────┘
```

### 9.2 Admin Knowledge Gap Resolution Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      ADMIN KNOWLEDGE IMPROVEMENT FLOW                    │
│                                                                         │
│  ┌─────────────────┐                                                   │
│  │ Admin logs into  │                                                   │
│  │ dashboard        │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ View "Knowledge │                                                   │
│  │ Gaps" section   │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ See top gap:    │                                                   │
│  │ "Return Policy" │                                                   │
│  │ 23 unanswered   │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Click gap to    │                                                   │
│  │ view examples   │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ [Add Document]  │                                                   │
│  │ uploads return  │                                                   │
│  │ policy PDF      │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Document        │                                                   │
│  │ processes       │                                                   │
│  │ (chunk, embed, │                                                   │
│  │  index)         │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ AI can now      │                                                   │
│  │ answer return   │                                                   │
│  │ policy questions│                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Monitor: Return │                                                   │
│  │ questions ↓     │                                                   │
│  │ Escalations ↓   │                                                   │
│  │ Automation ↑    │                                                   │
│  └─────────────────┘                                                   │
│                                                                         │
│  SUCCESS: Knowledge gap closed, automation rate improved               │
└─────────────────────────────────────────────────────────────────────────┘
```

### 9.3 Organization Onboarding Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      ORGANIZATION ONBOARDING FLOW                       │
│                                                                         │
│  PLATFORM ADMIN                                                         │
│  │                                                                     │
│  ▼                                                                     │
│  ┌─────────────────┐                                                   │
│  │ Create new org  │                                                   │
│  │ in platform    │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Send invite to  │                                                   │
│  │ org admin       │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ORG ADMIN                                                             │
│  │                                                                     │
│  ▼                                                                     │
│  ┌─────────────────┐                                                   │
│  │ Set up account  │                                                   │
│  │ & org profile   │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Upload initial  │                                                   │
│  │ documents       │                                                   │
│  │ (knowledge base)│                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Configure MCP   │                                                   │
│  │ servers (opt.)  │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Customize AI    │                                                   │
│  │ persona &       │                                                   │
│  │ greeting        │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Test with       │                                                   │
│  │ sample calls    │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────┐                                                   │
│  │ Go Live!        │                                                   │
│  │ Customers can   │                                                   │
│  │ now call        │                                                   │
│  └─────────────────┘                                                   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 10. Responsive Design

### 10.1 Breakpoints

| Breakpoint | Width | Target Device |
|------------|-------|---------------|
| **Mobile** | < 640px | Smartphones |
| **Tablet** | 640px - 1024px | Tablets, mobile landscape |
| **Desktop** | > 1024px | Desktop browsers |

### 10.2 Mobile Adaptations

| Component | Desktop | Mobile |
|-----------|---------|--------|
| **Dashboard** | Multi-column grid | Single column, cards stack |
| **Voice Call** | Full screen | Compact with bottom controls |
| **Document List** | Table view | Card list |
| **Charts** | Full width with legend | Simplified, scrollable |

### 10.3 Mobile Voice Call Interface

```
┌───────────────────────────┐
│  [Logo]        [End Call] │
├───────────────────────────┤
│                           │
│      ┌───────────┐        │
│      │    🤖     │        │
│      │  Avatar   │        │
│      │ Listening │        │
│      └───────────┘        │
│                           │
│  ┌─────────────────────┐  │
│  │ "I'm looking into   │  │
│  │  that for you..."   │  │
│  └─────────────────────┘  │
│                           │
│  [Waveform Animation]     │
│  ▓▓▓▓▓▓▓▓░░░░░░░░░░░      │
│                           │
│  Transcript:               │
│  ┌─────────────────────┐  │
│  │ Customer: How do I   │  │
│  │ return my order?     │  │
│  │ AI: To return an    │  │
│  │ order, you can...   │  │
│  └─────────────────────┘  │
│                           │
├───────────────────────────┤
│  [Mute]  [💬 Text]  [📞]  │
└───────────────────────────┘
```

---

## 11. Error Handling UX

### 11.1 Error State Templates

| Error Type | UI Template | Recovery |
|------------|-------------|----------|
| **Network Offline** | Banner + retry button | Auto-retry on reconnect |
| **Session Expired** | Modal + redirect | Refresh page |
| **Upload Failed** | Inline error + retry | Retry button |
| **API Error** | Toast notification | Dismiss or retry |

### 11.2 Toast Notifications

```
┌─────────────────────────────────────────────────┐
│  ✓ Success: Document uploaded successfully     │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│  ⚠️ Warning: Large file may take longer        │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│  ❌ Error: Connection lost. [Retry]            │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│  ℹ️ Info: Your changes have been saved          │
└─────────────────────────────────────────────────┘
```

---

## 12. Accessibility Checklist

| Requirement | Status |
|-------------|--------|
| **Color contrast** | ✅ WCAG 2.1 AA compliant |
| **Keyboard navigation** | ✅ All interactions accessible via keyboard |
| **Focus indicators** | ✅ Visible focus rings on all interactive elements |
| **ARIA labels** | ✅ All buttons, inputs, and regions labeled |
| **Screen reader** | ✅ Tested with common screen readers |
| **Reduced motion** | ✅ Respects prefers-reduced-motion |
| **Text scaling** | ✅ UI scales up to 200% without breaking |
| **Touch targets** | ✅ Minimum 44px on all mobile interactions |

---

## 13. Prototyping Notes

### 13.1 Recommended Prototyping Tools

| Tool | Use Case |
|------|----------|
| **Figma** | Primary design tool for screens and components |
| **Principle** | Animation and interaction prototyping |
| **Maze** | User testing with interactive prototypes |

### 13.2 Prototype Scope for MVP

| Priority | Screens |
|----------|---------|
| **P0** | Customer voice call, Customer text chat, Escalation flow |
| **P1** | Admin dashboard overview, Document upload |
| **P2** | Admin analytics, Super admin org management |
| **P3** | Full responsive adaptation, Advanced settings |

---

*Document Status: Draft — For design and implementation guidance*
*Next Steps: Review with team, create interactive prototypes, user testing*