/**
 * Mock data for the org console overview. Per AGENTS.md: one file per page
 * under `src/mocks/{interface}/`, exported as typed factories so the page
 * can swap in real API data later without changing component shape.
 *
 * All values are deterministic — reloads stay stable for visual review.
 */

export type HealthBand = "normal" | "monitor" | "atRisk" | "critical";

export type Channel = "voice" | "chat";

export type TopicCategory =
  | "Account"
  | "Billing"
  | "Product"
  | "Order"
  | "Technical"
  | "General";

export type OrgSnapshot = {
  name: string;
  plan: "Free" | "Starter" | "Growth" | "Scale";
  region: string;
};

export type KpiTile = {
  id: string;
  value: string;
  label: string;
  caption: string;
  delta: { value: string; direction: "up" | "down" | "flat" };
  /** Last 7 days, oldest → newest, normalized 0–1. */
  spark: number[];
};

export type LiveConversation = {
  id: string;
  topic: TopicCategory;
  channel: Channel;
  /** Seconds since session start. */
  durationSec: number;
  /** Health score, 0–1. Maps to band per PRD §4.8. */
  healthScore: number;
  healthBand: HealthBand;
  assignee: "AI" | string;
  lastUtterance: string;
  customer: string;
};

export type ActivityEvent = {
  id: string;
  /** Short, terminal-friendly label. */
  kind: "tool" | "escalation" | "document" | "resolved" | "gap";
  message: string;
  atMinAgo: number;
};

export type TopicDatum = {
  category: TopicCategory;
  count: number;
  /** 0–1 share of total. */
  share: number;
};

export type KnowledgeGap = {
  id: string;
  question: string;
  askedTimes: number;
  lastAskedAt: string;
};

export type Escalation = {
  id: string;
  customer: string;
  topic: TopicCategory;
  waitingSec: number;
  priority: "high" | "medium" | "low";
  reason: string;
};

export const getOrgSnapshot = (): OrgSnapshot => ({
  name: "Acme Corp",
  plan: "Growth",
  region: "us-east-1",
});

export const getKpiSnapshot = (): KpiTile[] => [
  {
    id: "conversations",
    value: "1,284",
    label: "Conversations today",
    caption: "vs. yesterday",
    delta: { value: "+12%", direction: "up" },
    spark: [0.4, 0.55, 0.6, 0.5, 0.7, 0.85, 1.0],
  },
  {
    id: "automation",
    value: "86%",
    label: "Automation rate",
    caption: "resolved by AI",
    delta: { value: "+3.1pp", direction: "up" },
    spark: [0.6, 0.65, 0.7, 0.72, 0.78, 0.82, 0.86],
  },
  {
    id: "responseTime",
    value: "0:42",
    label: "Avg response time",
    caption: "across voice & chat",
    delta: { value: "−8s", direction: "down" },
    spark: [0.7, 0.6, 0.55, 0.5, 0.4, 0.35, 0.3],
  },
  {
    id: "satisfaction",
    value: "78",
    label: "Customer satisfaction",
    caption: "composite score",
    delta: { value: "+4", direction: "up" },
    spark: [0.6, 0.62, 0.65, 0.7, 0.72, 0.74, 0.78],
  },
];

export const getLiveConversations = (): LiveConversation[] => [
  {
    id: "#4821",
    topic: "Billing",
    channel: "voice",
    durationSec: 42,
    healthScore: 0.92,
    healthBand: "normal",
    assignee: "AI",
    customer: "J. Patel",
    lastUtterance: "Thanks, that clears it up.",
  },
  {
    id: "#4820",
    topic: "Account",
    channel: "chat",
    durationSec: 78,
    healthScore: 0.61,
    healthBand: "monitor",
    assignee: "AI",
    customer: "M. Okafor",
    lastUtterance: "I tried that already — let me explain.",
  },
  {
    id: "#4819",
    topic: "Technical",
    channel: "voice",
    durationSec: 184,
    healthScore: 0.42,
    healthBand: "atRisk",
    assignee: "AI",
    customer: "S. Lindgren",
    lastUtterance: "It's still throwing the same error after the restart.",
  },
  {
    id: "#4818",
    topic: "Order",
    channel: "chat",
    durationSec: 22,
    healthScore: 0.88,
    healthBand: "normal",
    assignee: "AI",
    customer: "D. Cho",
    lastUtterance: "Got it, thanks!",
  },
  {
    id: "#4817",
    topic: "Billing",
    channel: "voice",
    durationSec: 311,
    healthScore: 0.21,
    healthBand: "critical",
    assignee: "AI",
    customer: "R. Almeida",
    lastUtterance: "This is the third time I've had to call about this.",
  },
];

export const getRecentActivity = (): ActivityEvent[] => [
  {
    id: "a1",
    kind: "tool",
    message: "MCP tool returned order status for #4821",
    atMinAgo: 1,
  },
  {
    id: "a2",
    kind: "resolved",
    message: "AI resolved conversation #4818 in 22s",
    atMinAgo: 2,
  },
  {
    id: "a3",
    kind: "escalation",
    message: "Conversation #4817 escalated to human queue",
    atMinAgo: 3,
  },
  {
    id: "a4",
    kind: "gap",
    message: "Knowledge gap: 'how to change billing cycle' (asked 6×)",
    atMinAgo: 6,
  },
  {
    id: "a5",
    kind: "document",
    message: "Refund policy v3 indexed into knowledge base",
    atMinAgo: 11,
  },
  {
    id: "a6",
    kind: "tool",
    message: "MCP tool timeout on Stripe connector (auto-retry ok)",
    atMinAgo: 14,
  },
  {
    id: "a7",
    kind: "resolved",
    message: "AI resolved conversation #4811 with satisfaction 92",
    atMinAgo: 18,
  },
  {
    id: "a8",
    kind: "escalation",
    message: "Conversation #4809 claimed by agent K. Singh",
    atMinAgo: 22,
  },
];

export const getTopicBreakdown = (): TopicDatum[] => [
  { category: "Billing", count: 487, share: 0.38 },
  { category: "Account", count: 282, share: 0.22 },
  { category: "Product", count: 198, share: 0.155 },
  { category: "Order", count: 154, share: 0.12 },
  { category: "Technical", count: 96, share: 0.075 },
  { category: "General", count: 67, share: 0.05 },
];

export const getKnowledgeGaps = (): KnowledgeGap[] => [
  {
    id: "g1",
    question: "How do I change my billing cycle from annual to monthly?",
    askedTimes: 6,
    lastAskedAt: "12 min ago",
  },
  {
    id: "g2",
    question: "Does the Pro plan include SSO at no extra cost?",
    askedTimes: 4,
    lastAskedAt: "47 min ago",
  },
  {
    id: "g3",
    question: "Can I export call recordings for compliance?",
    askedTimes: 3,
    lastAskedAt: "1 h ago",
  },
  {
    id: "g4",
    question: "What happens to my data if I cancel mid-cycle?",
    askedTimes: 3,
    lastAskedAt: "2 h ago",
  },
  {
    id: "g5",
    question: "How do I set up per-region sub-accounts?",
    askedTimes: 2,
    lastAskedAt: "3 h ago",
  },
  {
    id: "g6",
    question: "Is there a way to pause my subscription for 60 days?",
    askedTimes: 2,
    lastAskedAt: "yesterday",
  },
];

export const getAgentQueue = (): Escalation[] => [
  {
    id: "#4817",
    customer: "R. Almeida",
    topic: "Billing",
    waitingSec: 311,
    priority: "high",
    reason: "Repeated escalation flag — same customer, third call this week.",
  },
  {
    id: "#4819",
    customer: "S. Lindgren",
    topic: "Technical",
    waitingSec: 184,
    priority: "high",
    reason: "AI flagged low confidence; customer frustrated after restart.",
  },
  {
    id: "#4809",
    customer: "A. Mensah",
    topic: "Account",
    waitingSec: 92,
    priority: "medium",
    reason: "Account recovery flow needs identity verification.",
  },
  {
    id: "#4803",
    customer: "Y. Tanaka",
    topic: "Product",
    waitingSec: 41,
    priority: "low",
    reason: "Customer asked to speak to a human about a feature request.",
  },
];
