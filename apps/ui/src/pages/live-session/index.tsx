/**
 * /console/live — admin-initiated voice session.
 *
 * The page is intentionally simple: an entry animation, then either
 * the initiate card or the session view, with the same ConsoleShell
 * chrome the rest of /console uses. The lifecycle hook (use-live-
 * session) owns the WS + media handles.
 */

import { useEffect } from "react";
import { motion } from "framer-motion";
import { Container } from "@/components/ui/container";
import { ConsoleShell } from "@/components/layout/console-shell";
import { InitiateCard } from "@/pages/live-session/parts/initiate-card";
import { SessionView } from "@/pages/live-session/parts/session-view";
import { useLiveSession } from "@/pages/live-session/use-live-session";
import { liveStore } from "@/lib/live/store";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";

const ease = [0.22, 1, 0.36, 1] as const;

function friendlyErrorMessage(code: string, fallback: string): string {
 switch (code) {
 case "mic_denied":
 return LIVE_SESSION_PAGE_LABELS.error.micDenied;
 case "gemini_unavailable":
 return LIVE_SESSION_PAGE_LABELS.error.geminiUnavailable;
 default:
 return fallback;
 }
}

export function LiveSessionPage() {
 useEffect(() => {
 document.title = LIVE_SESSION_PAGE_LABELS.meta.title;
 }, []);

 const { status, error, start, end, toggleMute } = useLiveSession();
 const muted = liveStore.getState().micMuted;

 return (
 <ConsoleShell>
 <Container size="default" className="py-8 sm:py-10">
 <motion.div
 initial={{ opacity: 0, y: 12 }}
 animate={{ opacity: 1, y: 0 }}
 transition={{ duration: 0.6, ease }}
 className="mb-6"
 >
 <h1 className="text-2xl font-semibold tracking-tight text-fg sm:text-3xl">
 {LIVE_SESSION_PAGE_LABELS.meta.title}
 </h1>
 <p className="mt-1 text-sm text-fg-muted">
 {LIVE_SESSION_PAGE_LABELS.meta.description}
 </p>
 </motion.div>

 {error && (
 <motion.div
 initial={{ opacity: 0 }}
 animate={{ opacity: 1 }}
 className="mx-auto mb-4 max-w-3xl rounded-2xl border border-border bg-danger-soft px-4 py-3 text-sm text-danger"
 >
 {friendlyErrorMessage(error.code, error.message || LIVE_SESSION_PAGE_LABELS.error.generic)}
 </motion.div>
 )}

 {(status === "idle" || status === "connecting") && (
 <InitiateCard status={status} onStart={start} />
 )}

 {(status === "live" || status === "interrupted" || status === "ended") && (
 <SessionView muted={muted} onToggleMute={toggleMute} onEnd={end} />
 )}
 </Container>
 </ConsoleShell>
 );
}
