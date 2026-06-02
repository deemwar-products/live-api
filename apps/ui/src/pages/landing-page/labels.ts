/**
 * Landing page labels. All visible strings on the marketing site live here.
 * When real i18n lands, swap the import to a translation bundle.
 */
export const LANDING_LABELS = {
  nav: {
    product: "Product",
    howItWorks: "How it works",
    useCases: "Use cases",
    pricing: "Pricing",
  },

  hero: {
    title: "Resolve 80% of support without a human.",
    subtitle:
      "Live API gives your customers instant, accurate answers over voice and chat — trained on your knowledge base, available 24/7, escalating to a human only when it should.",
    primaryCta: "Start free",
    secondaryCta: "See how it works",
    trustNote: "No credit card. Live in under 10 minutes.",
  },

  problem: {
    title: "Your support queue is a tax on growth.",
    items: [
      {
        title: "Customers wait, then leave.",
        body: "71% of customers expect a response within five minutes. A backlog of tickets is a backlog of churn.",
      },
      {
        title: "Good agents do repetitive work.",
        body: "Your best people spend their days answering the same ten questions. Burnout follows, and so does turnover.",
      },
      {
        title: "Coverage never quite scales.",
        body: "Hiring more agents is linear, expensive, and slow. Demand is not.",
      },
    ],
  },

  product: {
    title: "Three steps. One calm experience.",
    steps: [
      {
        title: "Bring your knowledge",
        body: "Upload docs, FAQs, and help-center articles. Live API indexes them in minutes and supports 100+ languages with cross-lingual retrieval.",
      },
      {
        title: "Customers ask, naturally",
        body: "Voice or chat on your subdomain, in your widget, or via API. The AI responds in your tone, in the customer's language, with the right tool calls when it needs them.",
      },
      {
        title: "Escalate with full context",
        body: "When a question needs a human, the conversation — transcript, retrieved knowledge, the AI's confidence signals — is handed off. The customer never repeats themselves.",
      },
    ],
  },

  capabilities: {
    title: "Everything support needs. Nothing it doesn't.",
    items: [
      {
        title: "Voice that feels like voice",
        body: "Sub-500ms audio round trip. Customers can interrupt mid-sentence. Page refresh doesn't end the call.",
      },
      {
        title: "Knowledge that stays fresh",
        body: "Drop a new doc and the system prompt refreshes automatically. No manual retraining, no stale answers.",
      },
      {
        title: "Tool calls, scoped to you",
        body: "Connect MCP servers for orders, billing, and tickets. Credentials are encrypted and isolated per organization.",
      },
      {
        title: "Live monitoring",
        body: "Watch every conversation in real time. A health score per turn surfaces the ones that need a human — before customers ask.",
      },
      {
        title: "Topic tagging, automated",
        body: "Every conversation is tagged to the right topic — Billing, Account, Product — so you can see what to build next.",
      },
      {
        title: "Built for compliance",
        body: "Encryption in transit and at rest. Org-level data isolation. Audit logs for every admin action. MFA for super admins.",
      },
    ],
  },

  ticker: ["Voice", "Chat", "RAG", "MCP", "Multilingual", "Live monitoring", "Topic tagging", "24/7"],

  liveDemo: {
    title: "Listen to a real conversation.",
    body: "Live API on a typical SaaS support queue — a billing question, a follow-up, a clean handoff. No script, no studio.",
    cta: "Play sample",
    duration: "1 min 42 sec",
  },

  useCases: {
    title: "Built for the conversations that matter.",
    items: [
      {
        title: "SaaS support",
        body: "Replace tier-1 queues. Free your team to ship, not to answer reset-password tickets.",
      },
      {
        title: "DTC and retail",
        body: "Order status, returns, exchanges — answered at midnight when customers actually shop.",
      },
      {
        title: "Fintech and banking",
        body: "KYC questions, card issues, statement queries. Encrypted, isolated, audit-logged.",
      },
      {
        title: "Healthcare intake",
        body: "Triage patient questions, route to the right clinic, and respect every privacy boundary.",
      },
    ],
  },

  outcomes: {
    title: "Outcomes, not features.",
    items: [
      { metric: "80–90%", label: "of conversations resolved by AI" },
      { metric: "< 2s", label: "average response time" },
      { metric: "< 500ms", label: "voice round-trip latency" },
      { metric: "24/7", label: "coverage across time zones" },
    ],
  },

  pricing: {
    title: "Pay for what you use. Nothing else.",
    body: "Live API is billed by tokens processed. Voice, chat, and tools share one transparent meter.",
    cta: "See full pricing",
    note: "Free tier available. No seat fees. No minimums.",
  },

  cta: {
    title: "Your customers shouldn't have to wait.",
    body: "Set up Live API on your subdomain in under ten minutes. Bring your docs, keep your tone.",
    primary: "Start free",
    secondary: "Talk to the team",
  },

  footer: {
    tagline: "Voice and chat AI for support teams that measure outcomes.",
    columns: {
      product: { title: "Product", links: ["Features", "How it works", "Pricing", "Changelog"] },
      company: { title: "Company", links: ["About", "Careers", "Press", "Contact"] },
      resources: { title: "Resources", links: ["Docs", "API reference", "Security", "Status"] },
      legal: { title: "Legal", links: ["Privacy", "Terms", "DPA", "Subprocessors"] },
    },
    copyright: "© 2026 Live API. All rights reserved.",
  },
} as const;
