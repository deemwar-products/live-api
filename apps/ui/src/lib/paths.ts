/**
 * Route path constants. Every URL in the app is defined here so refactors
 * are one-file changes, not a global find-and-replace.
 *
 * Convention:
 *   - LANDING.*        marketing routes (no auth)
 *   - PUBLIC_ROUTES.* routes anyone can hit (no auth)
 *   - BUSINESS.*       org-admin routes (auth required)
 *   - PLATFORM.*       super-admin routes (auth required)
 *
 * Use these in <Link to={...}>, navigate(...), and href attributes.
 *
 * Note: PUBLIC is a reserved global in some JS environments, so we name the
 * group PUBLIC_ROUTES to avoid shadowing.
 */

export const LANDING = {
  home: "/",
} as const;

export const PUBLIC_ROUTES = {
  auth: "/auth",
} as const;

export const BUSINESS = {
  dashboard: "/b",
  conversations: "/b/conversations",
  knowledge: "/b/knowledge",
  tools: "/b/tools",
  monitoring: "/b/monitoring",
  team: "/b/team",
  billing: "/b/billing",
  settings: "/b/settings",
} as const;

export const CONSOLE = {
  home: "/console",
  onboarding: "/console/onboarding",
  conversations: "/console/conversations",
  knowledge: "/console/knowledge",
  tools: "/console/tools",
  monitoring: "/console/monitoring",
  team: "/console/team",
  billing: "/console/billing",
  settings: "/console/settings",
} as const;

export const PLATFORM = {
  dashboard: "/p",
  organizations: "/p/organizations",
  users: "/p/users",
  audit: "/p/audit",
  settings: "/p/settings",
} as const;
