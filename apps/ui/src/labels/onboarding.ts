/**
 * Onboarding flow labels. Shared across all 4 steps — page-specific
 * copy lives in `pages/console/onboarding/labels.ts`.
 */
export const ONBOARDING_LABELS = {
  shell: {
    title: "Get Live API set up",
    saveAndExit: "Save & finish later",
    brand: "Live API",
  },

  stepper: {
    step1: "Knowledge",
    step2: "AI persona",
    step3: "Tools",
    step4: "Team",
  },

  actions: {
    back: "Back",
    next: "Next",
    skip: "Skip for now",
    finish: "Finish setup",
    goToDashboard: "Go to dashboard",
    customizeFurther: "Customize further",
  },

  knowledge: {
    title: "Bring your knowledge",
    description:
      "Upload the docs, FAQs, and policies your AI should learn from. We'll chunk, embed, and index them — usually in under a minute.",
    dropzone: {
      title: "Drop files here",
      subtitle: "or click to browse · PDF, DOCX, TXT, MD up to 25 MB each",
    },
    sources: {
      title: "Connect a source",
      notion: "Notion",
      confluence: "Confluence",
      zendesk: "Zendesk",
      url: "Public URL",
    },
    uploaded: {
      title: "Indexed documents",
      empty: "No documents yet — upload at least one to get useful answers.",
    },
    tips: {
      title: "What we do with your docs",
      items: [
        "Chunked at 512 tokens with 64-token overlap for context continuity.",
        "Embedded in 100+ languages — ask in any of them.",
        "Stored in a per-org namespace. No cross-org leakage, ever.",
        "Re-indexed automatically when you drop a new version.",
      ],
    },
  },

  persona: {
    title: "Shape your AI's voice",
    description:
      "The system prompt is append-only — the platform prompt underneath stays untouched. The greeting is its own thing and is what customers see first.",
    mode: {
      label: "Generation mode",
      auto: "Auto-generate",
      autoEdit: "Auto + edit",
      custom: "Custom append",
      hint: "Auto-generate writes a starter prompt from your knowledge base. Custom append lets you add your own block on top.",
    },
    tone: {
      label: "Tone",
      options: { calm: "Calm", friendly: "Friendly", formal: "Formal", technical: "Technical" },
    },
    languages: {
      label: "Languages",
      addPlaceholder: "Add a language (e.g. es, fr, de)",
      hint: "100+ languages supported. Customers will be auto-routed to the right one.",
    },
    systemPrompt: {
      label: "System prompt",
      hint: "Append-only. The platform prompt under this is fixed.",
    },
    greeting: {
      label: "Greeting message",
      enabled: "Greeting enabled",
      hint: "What customers see when the session starts. Short, calm, on-brand.",
    },
    preview: {
      title: "Live preview",
      customerSays: "Hi! I have a question about my last invoice.",
      aiResponds:
        "Hi there — happy to help. Could you share the invoice number? It starts with INV- and is in the email we sent on the 1st.",
    },
  },

  tools: {
    title: "Choose what your AI can do",
    description:
      "Five platform tools are enabled by default. You can opt out of any that don't apply — for example, if you don't have a ticketing system, turn off Create Ticket.",
    whatAre: {
      title: "What are MCP tools?",
      body: "MCP (Model Context Protocol) tools let the AI take real actions — pull an order status, file a ticket, page your team — instead of just answering questions. The five below are first-party platform tools, scoped to your org.",
    },
    items: {
      rag: {
        name: "RAG Retrieval",
        description: "Search your knowledge base for the most relevant chunks before answering.",
      },
      notify: {
        name: "Send Notification",
        description: "Ping the right person on your team when something needs human eyes.",
      },
      ticket: {
        name: "Create Ticket",
        description: "Open a support ticket in your connected helpdesk when escalation is needed.",
      },
      gap: {
        name: "Knowledge Gap Flag",
        description: "Surface questions the AI couldn't answer so your content team can fill them in.",
      },
      credit: {
        name: "Credit Alert",
        description: "Warn an admin when usage approaches your plan's monthly limit.",
      },
    },
  },

  team: {
    title: "Invite your team",
    description:
      "Bring the people who'll be writing, monitoring, and stepping in. Each role gets scoped permissions — you can fine-tune them later.",
    inviteRow: {
      emailPlaceholder: "teammate@company.com",
      rolePlaceholder: "Role",
      invite: "Send invite",
    },
    pending: {
      title: "Pending invites",
      empty: "No invites yet — add at least one teammate to keep your inbox calm.",
    },
    roles: {
      title: "Roles at a glance",
      admin: {
        name: "Org Admin",
        perms: "Full org access — settings, billing, team, all data.",
      },
      lead: {
        name: "Agent Team Lead",
        perms: "Manage agents, view escalations, assign priority.",
      },
      agent: {
        name: "Human Agent",
        perms: "Take escalated calls, work the assigned queue.",
      },
      content: {
        name: "Content Manager",
        perms: "Upload docs, manage knowledge base, view search analytics.",
      },
      analyst: {
        name: "Analyst",
        perms: "View dashboards, export reports, tag questions.",
      },
    },
  },

  summary: {
    title: "You're set up.",
    subtitle: "Live API is ready to handle real customer conversations. You can fine-tune anything later.",
    checklist: {
      knowledge: "Knowledge base connected",
      persona: "AI voice and greeting configured",
      tools: "Tools selected",
      team: "Team invited",
    },
  },
} as const;
