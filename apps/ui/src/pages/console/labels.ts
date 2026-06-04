/**
 * Page-specific labels for the org console overview. Shared strings live
 * in `src/labels/business.ts`; this file is only for copy that's unique
 * to the overview page composition.
 */
export const CONSOLE_PAGE_LABELS = {
  meta: {
    title: "Overview · Live API",
    description: "Real-time view of your support conversations and AI performance.",
  },

  greetingName: "Priya",
  orgName: "Acme Corp",

  volumeChart: {
    label: "Conversation volume",
    caption: "Last 7 days",
    todayMarker: "Today",
  },
} as const;
