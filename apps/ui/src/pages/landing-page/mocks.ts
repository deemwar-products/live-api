/**
 * Mock data for the landing page. One file per page, as the AGENTS.md
 * convention requires. Replace with real API calls when ready — page
 * components should not need to change shape.
 */

export type TranscriptTurn = {
  speaker: "Customer" | "Live API";
  text: string;
  at: string;
};

export const DEMO_TRANSCRIPT: TranscriptTurn[] = [
  { speaker: "Customer", at: "0:04", text: "Hi — I was charged twice this month, can you take a look?" },
  {
    speaker: "Live API",
    at: "0:11",
    text: "Of course. I can see two charges on the 3rd and the 7th. The 7th looks like a duplicate from a failed retry. I'll start a refund for that one — it usually clears in three to five business days. Want me to also turn off auto-retry on your card so this doesn't happen again?",
  },
  { speaker: "Customer", at: "0:32", text: "Yes please. And can you email me a receipt for the original charge?" },
  {
    speaker: "Live API",
    at: "0:38",
    text: "Done on both — auto-retry is off, and the receipt is on its way to your inbox. Anything else I can help with?",
  },
  { speaker: "Customer", at: "0:48", text: "That's it, thanks." },
  { speaker: "Live API", at: "0:50", text: "Anytime. Have a good one." },
];
