/**
 * Business dashboard (org console) labels. All visible strings for the
 * `/console` route group live here. Pages may add their own `labels.ts`
 * for page-specific copy, but anything reused across console parts
 * belongs here.
 */
export const BUSINESS_LABELS = {
  shell: {
    brand: "Live API",
    orgSwitcherLabel: "Switch organization",
    searchPlaceholder: "Search conversations, topics, docs…",
    signOut: "Sign out",
    profile: "Profile",
  },

  nav: {
    overview: "Overview",
    conversations: "Conversations",
    knowledge: "Knowledge",
    tools: "Tools & MCP",
    monitoring: "Live monitoring",
    team: "Team",
    billing: "Billing",
    settings: "Settings",
  },

  greeting: {
    goodMorning: "Good morning",
    goodAfternoon: "Good afternoon",
    goodEvening: "Good evening",
    subtitle: "Here's how your support is performing right now.",
  },

  kpi: {
    conversations: {
      label: "Conversations today",
      caption: "vs. yesterday",
    },
    automation: {
      label: "Automation rate",
      caption: "resolved by AI",
    },
    responseTime: {
      label: "Avg response time",
      caption: "across voice & chat",
    },
    satisfaction: {
      label: "Customer satisfaction",
      caption: "composite score",
    },
  },

  live: {
    title: "Live conversations",
    caption: "5 active right now",
    viewAll: "View all",
    columns: {
      id: "ID",
      topic: "Topic",
      channel: "Channel",
      duration: "Duration",
      health: "Health",
      action: "Action",
    },
    actions: {
      watch: "Watch",
      takeover: "Take over",
    },
    channels: {
      voice: "Voice",
      chat: "Chat",
    },
    health: {
      label: "Health",
      normal: "Normal",
      monitor: "Monitor",
      atRisk: "At risk",
      critical: "Critical",
    },
  },

  activity: {
    title: "Recent activity",
    caption: "Last few minutes",
  },

  topics: {
    title: "Topic breakdown",
    caption: "Past 7 days",
    totalLabel: "conversations",
  },

  knowledge: {
    title: "Knowledge health",
    caption: "Gaps surfaced by the LLM judge",
    unansweredToday: "unanswered today",
    viewDocs: "Open knowledge base",
    samplesTitle: "Top unanswered questions",
  },

  agents: {
    title: "Escalation queue",
    caption: "Waiting for a human",
    columns: {
      id: "ID",
      customer: "Customer",
      topic: "Topic",
      waiting: "Waiting",
      priority: "Priority",
    },
    priorities: {
      high: "High",
      medium: "Medium",
      low: "Low",
    },
    claim: "Claim",
  },
} as const;
