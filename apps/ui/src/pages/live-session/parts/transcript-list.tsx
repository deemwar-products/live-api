/**
 * Append-only transcript list. Anchored to the bottom on new entries.
 * Each entry shows the speaker label and the current best text for
 * that turn. Finalised turns are dimmer; in-flight turns are bright.
 */

import { useEffect, useRef, useState } from "react";
import { liveStore, type TranscriptEntry } from "@/lib/live/store";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";
import { cn } from "@/lib/cn";

function useLiveTranscript(): TranscriptEntry[] {
 const [list, setList] = useState<TranscriptEntry[]>(() => liveStore.getState().transcript);
 useEffect(() => {
 return liveStore.subscribe((s) => setList(s.transcript));
 }, []);
 return list;
}

export function TranscriptList() {
 const list = useLiveTranscript();
 const bottomRef = useRef<HTMLDivElement>(null);

 useEffect(() => {
 bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
 }, [list.length]);

 if (list.length === 0) {
 return (
 <p className="text-sm text-fg-subtle">
 {LIVE_SESSION_PAGE_LABELS.transcript.empty}
 </p>
 );
 }

 return (
 <ol className="flex flex-col gap-3">
 {list.map((entry) => (
 <li
 key={entry.turnId + ":" + entry.ts}
 className={cn(
 "flex flex-col gap-1 text-sm",
 entry.turnComplete ? "opacity-80" : "opacity-100",
 )}
 >
 <span
 className={cn(
 "text-[11px] font-semibold uppercase tracking-wider",
 entry.role === "user" ? "text-accent" : "text-fg-muted",
 )}
 >
 {entry.role === "user"
 ? LIVE_SESSION_PAGE_LABELS.transcript.userLabel
 : LIVE_SESSION_PAGE_LABELS.transcript.modelLabel}
 {!entry.turnComplete && (
 <span className="ml-1.5 inline-block size-1.5 animate-pulse rounded-full bg-accent align-middle" />
 )}
 </span>
 <p className="leading-relaxed text-fg">{entry.text}</p>
 </li>
 ))}
 <div ref={bottomRef} />
 </ol>
 );
}
