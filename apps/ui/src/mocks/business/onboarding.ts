/**
 * Mock state for the org onboarding flow. Per AGENTS.md: one file per
 * flow under `src/mocks/{interface}/`. The real implementation will
 * read this from a per-org record (or a `live-api.onboarding.completed`
 * flag in user metadata). For now, the page reads/writes a localStorage
 * key to drive the `/console` redirect.
 */

export type Source = "notion" | "confluence" | "zendesk" | "url";

export type DocumentRow = {
  id: string;
  name: string;
  sizeKb: number;
  status: "indexed" | "processing" | "queued";
  lang: string;
  chunks: number;
};

export type ToolKey = "rag" | "notify" | "ticket" | "gap" | "credit";

export type RoleKey = "admin" | "lead" | "agent" | "content" | "analyst";

export type Invite = {
  id: string;
  email: string;
  role: RoleKey;
  sentAt: string;
};

export type OnboardingState = {
  /** Documents already in the org's knowledge base. */
  documents: DocumentRow[];
  /** Languages the org wants supported. ISO codes, lowercase. */
  languages: string[];
  /** System-prompt generation mode. */
  mode: "auto" | "autoEdit" | "custom";
  tone: "calm" | "friendly" | "formal" | "technical";
  systemPrompt: string;
  greetingEnabled: boolean;
  greeting: string;
  /** Which of the 5 platform tools are enabled. */
  tools: Record<ToolKey, boolean>;
  /** Pending team invites. */
  invites: Invite[];
  /** Steps the user has finished. 1-indexed. */
  completedSteps: number[];
};

const STORAGE_KEY = "live-api.onboarding.completed";

export const isOnboardingComplete = (): boolean => {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(STORAGE_KEY) === "1";
};

export const markOnboardingComplete = (): void => {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(STORAGE_KEY, "1");
};

export const resetOnboarding = (): void => {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(STORAGE_KEY);
};

/**
 * Initial state — empty by design. A new org lands on step 1 with
 * nothing filled in, so the form is honest about what's unconfigured.
 */
export const getOnboardingState = (): OnboardingState => ({
  documents: [],
  languages: ["en"],
  mode: "auto",
  tone: "calm",
  systemPrompt: "",
  greetingEnabled: true,
  greeting: "Hi! How can I help you today?",
  tools: {
    rag: true,
    notify: true,
    ticket: true,
    gap: true,
    credit: true,
  },
  invites: [],
  completedSteps: [],
});
