/**
 * Live waveform/level meter. Two thin animated bars: mic (green) and
 * model (blue). Each bar pulses on activity — a quick "thump" on the
 * rising edge, then decays. The shape is intentionally minimal so
 * the transcript card is the focus.
 */

import { cn } from "@/lib/cn";
import { useEffect, useState } from "react";

/**
 * Hook that returns a [0,1] activity level that decays toward 0 after
 * every "ping" call. Use ping() to bump activity; the bar will return
 * to baseline over `decayMs` even with no further pings.
 */
function useDecayLevel(decayMs = 250): { level: number; ping: (v?: number) => void } {
 const [level, setLevel] = useState(0);
 useEffect(() => {
 if (level <= 0) return;
 const id = window.setTimeout(() => setLevel((l) => Math.max(0, l - 0.1)), decayMs / 10);
 return () => window.clearTimeout(id);
 }, [level, decayMs]);
 return {
 level,
 ping: (v = 1) => setLevel(Math.min(1, v)),
 };
}

export function Waveform({
 micActive = false,
 modelActive = false,
 className,
}: {
 micActive?: boolean;
 modelActive?: boolean;
 className?: string;
}) {
 const mic = useDecayLevel();
 const model = useDecayLevel();

 // Bump activity when props say "active" — parent controls when to
 // pulse by setting these flags for at least one render frame.
 useEffect(() => {
 if (micActive) mic.ping();
 }, [micActive, mic]);
 useEffect(() => {
 if (modelActive) model.ping();
 }, [modelActive, model]);

 const bar = (level: number, activeColor: string, idleColor: string) => (
 <div className="relative h-1 flex-1 overflow-hidden rounded-full bg-bg-muted">
 <div
 className={cn("absolute inset-y-0 left-0 rounded-full transition-all duration-100 ease-out", activeColor)}
 style={{ width: `${Math.max(8, level * 100)}%`, opacity: level > 0 ? 1 : 0.4 }}
 />
 <div
 className={cn("absolute inset-y-0 right-0 h-full w-1 rounded-full", idleColor)}
 style={{ opacity: level > 0 ? 0.9 : 0.3 }}
 />
 </div>
 );

 return (
 <div className={cn("flex w-full items-center gap-2", className)}>
 {bar(mic.level, "bg-success", "bg-success/40")}
 {bar(model.level, "bg-info", "bg-info/40")}
 </div>
 );
}