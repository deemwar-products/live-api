/**
 * Common labels — used across interfaces. Per-interface labels live in
 * `customer.ts`, `business.ts`, `platform.ts`, `landing.ts`. Components must
 * never hard-code visible strings; always reference these.
 */
export const COMMON_LABELS = {
  app: {
    name: "Live API",
  },
  actions: {
    learnMore: "Learn more",
    getStarted: "Get started",
    signIn: "Sign in",
    signOut: "Sign out",
    seeHowItWorks: "See how it works",
    startFree: "Start free",
    bookDemo: "Book a demo",
    contactSales: "Contact sales",
    viewPricing: "View pricing",
    readDocs: "Read docs",
  },
} as const;
