/**
 * Session view — shown when status is live, interrupted, or ended.
 * Header with model + session id + status pill; controls bar; and a
 * scrolling transcript panel. The transcript card is allowed to
 * scroll inside its own bounds so the page chrome doesn't move.
 */

import { useEffect, useState } from "react";
import { Card, CardBody, CardHeader } from "@/components/ui/card";
import { liveStore } from "@/lib/live/store";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";
import { StatusIndicator } from "@/pages/live-session/parts/status-indicator";
import { TranscriptList } from "@/pages/live-session/parts/transcript-list";
import { ControlsBar } from "@/pages/live-session/parts/controls-bar";
import type { StatusState } from "@/lib/live/protocol";

function useStatus(): StatusState | "idle" {
 const [s, setS] = useState(() => liveStore.getState().status);
 useEffect(() => liveStore.subscribe((next) => setS(next.status)), []);
 return s;
}

function formatElapsed(ms: number) {
 const s = Math.max(0, Math.floor(ms / 1000));
 const m = Math.floor(s / 60);
 const r = s % 60;
 return `${m}:${r.toString().padStart(2, "0")}`;
}

const STATUS_LABEL: Record<StatusState, string> = {
 connecting: LIVE_SESSION_PAGE_LABELS.session.statusConnecting,
 live: LIVE_SESSION_PAGE_LABELS.session.statusLive,
 interrupted: LIVE_SESSION_PAGE_LABELS.session.statusInterrupted,
 ended: LIVE_SESSION_PAGE_LABELS.session.statusEnded,
};

export function SessionView({
 muted,
 onToggleMute,
 onEnd,
}: {
 muted: boolean;
 onToggleMute: () => void;
 onEnd: () => void;
}) {
 const status = useStatus();
 const ready = liveStore.getState().ready;
 const session = liveStore.getState().session;
 const [elapsed, setElapsed] = useState(0);

 useEffect(() => {
 if (!session.startedAt) return;
 const tick = () => {
 const end = session.endedAt ?? Date.now();
 setElapsed(end - session.startedAt!);
 };
 tick();
 const id = window.setInterval(tick, 1000);
 return () => window.clearInterval(id);
 }, [session.startedAt, session.endedAt]);

 const statusKey = status === "idle" ? "connecting" : status;

 return (
 <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
 <Card>
 <CardHeader
 title={
 <div className="flex items-center gap-3">
 <span>{ready?.model ?? "—"}</span>
 <StatusIndicator
 state={statusKey as StatusState}
 label={STATUS_LABEL[statusKey as StatusState] ?? "—"}
 />
 {status === "live" && (
 <span className="text-xs tabular-nums text-fg-muted">{formatElapsed(elapsed)}</span>
 )}
 </div>
 }
 caption={
 session.sessionId
 ? `${LIVE_SESSION_PAGE_LABELS.session.sessionIdLabel}: ${session.sessionId}`
 : undefined
 }
 />
 <CardBody className="space-y-4">
 <ControlsBar muted={muted} onToggleMute={onToggleMute} onEnd={onEnd} />
 </CardBody>
 </Card>

 <Card className="flex-1">
 <CardHeader
 title={LIVE_SESSION_PAGE_LABELS.transcript.title}
 />
 <CardBody className="max-h-[60vh] overflow-y-auto">
 <TranscriptList />
 </CardBody>
 </Card>
 </div>
 );
}
