/**
 * Page-specific labels for /console/live. Shared strings live in
 * `business.ts`; this file is only for copy that's unique to the
 * live session page composition.
 */
export const LIVE_SESSION_PAGE_LABELS = {
 meta: {
 title: "Live session · Live API",
 description:
 "Initiate a real-time voice session with Gemini Live. Speak naturally; transcripts stream here.",
 },

 initiate: {
 eyebrow: "Voice test",
 title: "Start a live session",
 body:
 "Click below to open a WebSocket to the backend, request microphone access, and begin a real-time voice conversation with Gemini Live.",
 ctaIdle: "Start live session",
 ctaConnecting: "Connecting…",
 permissionHint:
 "Your browser will ask for microphone permission the first time you start a session.",
 },

 session: {
 modelLabel: "Model",
 sessionIdLabel: "Session",
 statusConnecting: "Connecting",
 statusLive: "Live",
 statusInterrupted: "Interrupted",
 statusEnded: "Ended",
 },

 controls: {
 mute: "Mute",
 unmute: "Unmute",
 end: "End session",
 },

 transcript: {
 title: "Transcript",
 empty: "Say something — your words and Gemini's will appear here in real time.",
 userLabel: "You",
 modelLabel: "Gemini",
 },

 error: {
 generic: "Something went wrong. You can try starting a new session.",
 micDenied:
 "Microphone access was blocked. Allow it in your browser settings, then start a new session.",
 geminiUnavailable:
 "The backend couldn't reach Gemini Live. Check the API key in apps/api/.env.local and try again.",
 },

 ended: {
 title: "Session ended",
 body: "Wrap up your notes, or start a new session when you're ready.",
 cta: "Start a new session",
 },
} as const;
