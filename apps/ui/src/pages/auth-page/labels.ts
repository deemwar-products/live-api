/**
 * Auth page labels.
 */
export const AUTH_LABELS = {
  meta: {
    title: "Sign in to Live API",
    description: "Use your work account to sign in. SSO for your team is coming soon.",
  },
  brand: {
    tagline: "Voice and chat AI for support teams that measure outcomes.",
  },
  providers: {
    personal: "Personal",
    team: "For your team",
    google: {
      label: "Continue with Google",
    },
    microsoft: {
      label: "Continue with Microsoft",
    },
    sso: {
      label: "Continue with SSO (SAML)",
      comingSoonNote: "Feature coming soon",
    },
  },
  legal: {
    termsPrefix: "By continuing, you agree to our",
    terms: "Terms",
    and: "and",
    privacy: "Privacy Policy",
  },
  footer: {
    backToHome: "Back to home",
  },
} as const;
