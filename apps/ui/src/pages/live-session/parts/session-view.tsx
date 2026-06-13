/**
 * Session view — shown when status is live, interrupted, or ended.
 * Header with model + session id + status pill; controls bar; and a
 * scrolling transcript panel. The transcript card is allowed to
 * scroll inside its own bounds so the page chrome doesn't move.
 */

import { useEffect, useState } from "react";
import { RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardHeader } from "@/components/ui/card";
import { Waveform } from "@/components/waveform";
import { cn } from "@/lib/cn";
import { liveStore } from "@/lib/live/store";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";
import { StatusIndicator } from "@/pages/live-session/parts/status-indicator";
import { TranscriptList } from "@/pages/live-session/parts/transcript-list";
import { ControlsBar } from "@/pages/live-session/parts/controls-bar";
import { LiveCaption } from "@/pages/live-session/parts/live-caption";
import type { StatusState } from "@/lib/live/protocol";

function useStatus(): StatusState | "idle" {
 const [s, setS] = useState(() => liveStore.getState().status);
 useEffect(() => liveStore.subscribe((next) => setS(next.status)), []);
 return s;
}

function useSubState(): "idle" | "listening" | "speaking" | "paused" | "dropped" {
 const [s, setS] = useState(() => liveStore.getState().subState);
 useEffect(() => liveStore.subscribe((next) => setS(next.subState)), []);
 return s;
}

function useAudioActivity(): { micActive: boolean; modelActive: boolean } {
 const [s, setS] = useState(() => ({
 micActive: liveStore.getState().micActive,
 modelActive: liveStore.getState().modelActive,
 }));
 useEffect(
 () =>
 liveStore.subscribe((next) =>
 setS({ micActive: next.micActive, modelActive: next.modelActive }),
 ),
 [],
 );
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

type SubState = "idle" | "listening" | "speaking" | "paused" | "dropped";

const SUB_LABEL: Record<SubState, string> = {
 idle: "—",
 listening: LIVE_SESSION_PAGE_LABELS.session.subListening,
 speaking: LIVE_SESSION_PAGE_LABELS.session.subSpeaking,
 paused: LIVE_SESSION_PAGE_LABELS.session.subPaused,
 dropped: LIVE_SESSION_PAGE_LABELS.session.subDropped,
};

const SUB_TONE: Record<SubState, string> = {
 idle: "bg-bg-muted text-fg-muted",
 listening: "bg-success-soft text-success",
 speaking: "bg-info-soft text-info",
 paused: "bg-warning-soft text-warning",
 dropped: "bg-watch-soft text-watch",
};

const SUB_DOT: Record<SubState, string> = {
 idle: "bg-fg-subtle",
 listening: "bg-success",
 speaking: "bg-info",
 paused: "bg-warning",
 dropped: "bg-watch",
};

export function SessionView({
 muted,
 onToggleMute,
 onEnd,
 onStart,
}: {
 muted: boolean;
 onToggleMute: () => void;
 onEnd: () => void;
 onStart: () => void;
}) {
 const status = useStatus();
 const subState = useSubState();
 const audioActivity = useAudioActivity();
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
 const showSub = status === "live" || status === "interrupted";

 return (
 <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
 <Card>
 <CardHeader
 title={
 <div className="flex flex-wrap items-center gap-3">
 <span>{ready?.model ?? "—"}</span>
 <StatusIndicator
 state={statusKey as StatusState}
 label={STATUS_LABEL[statusKey as StatusState] ?? "—"}
 />
 {showSub && (
 <span
 className={cn(
 "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium tracking-tight",
 SUB_TONE[subState],
 )}
 >
 <span
 className={cn(
 "size-1.5 rounded-full",
 SUB_DOT[subState],
 (subState === "speaking" || subState === "listening") && "animate-pulse",
 )}
 />
 {SUB_LABEL[subState]}
 </span>
 )}
 {(status === "live" || status === "ended") && session.startedAt && (
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
 {showSub && (
 <Waveform
 micActive={audioActivity.micActive}
 modelActive={audioActivity.modelActive}
 />
 )}
 {showSub && <LiveCaption />}
 {status === "ended" ? (
 <div className="flex flex-col items-center gap-3 py-2 text-center">
 <span className="inline-flex size-10 items-center justify-center rounded-full bg-bg-muted text-fg-muted">
 <RotateCcw className="size-4" strokeWidth={1.75} />
 </span>
 <div className="space-y-1">
 <p className="text-sm font-medium text-fg">
 {LIVE_SESSION_PAGE_LABELS.ended.title}
 </p>
 <p className="text-sm text-fg-muted">
 {LIVE_SESSION_PAGE_LABELS.ended.body}
 </p>
 </div>
 <Button size="md" onClick={onStart} leading={<RotateCcw className="size-4" />}>
 {LIVE_SESSION_PAGE_LABELS.ended.cta}
 </Button>
 </div>
 ) : (
 <ControlsBar muted={muted} onToggleMute={onToggleMute} onEnd={onEnd} />
 )}
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
