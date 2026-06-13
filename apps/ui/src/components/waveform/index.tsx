/**
 * Live waveform/level meter. Two small equalizer-style bar groups — mic
 * (green, "You") and model (blue, "Gemini"). Each group's bars rise
 * together on activity with a slight per-bar stagger, then decay back
 * to a calm baseline. The shape is intentionally compact so the
 * transcript card is the focus.
 */

import { cn } from "@/lib/cn";
import { useEffect, useState } from "react";

/**
 * Hook that returns a [0,1] activity level that decays toward 0 after
 * every "ping" call. Use ping() to bump activity; the bars will return
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

// Relative amplitudes per bar — gives the group a "wave" silhouette
// instead of all bars moving in lockstep.
const BAR_PROFILE = [0.35, 0.7, 1, 0.65, 0.4];
const MIN_HEIGHT = 0.16;

function BarGroup({
  level,
  barColor,
  idleColor,
}: {
  level: number;
  barColor: string;
  idleColor: string;
}) {
  return (
    <div className="flex h-5 flex-1 items-center justify-center gap-[3px]">
      {BAR_PROFILE.map((amp, i) => (
        <span
          key={i}
          className={cn("w-1 rounded-full transition-all ease-out", level > 0 ? barColor : idleColor)}
          style={{
            height: `${Math.max(MIN_HEIGHT, level * amp) * 100}%`,
            transitionDuration: "150ms",
            transitionDelay: `${i * 20}ms`,
          }}
        />
      ))}
    </div>
  );
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

  return (
    <div className={cn("flex w-full items-center gap-4 rounded-xl bg-bg-muted/50 px-4 py-2.5", className)}>
      <div className="flex flex-1 items-center gap-2.5">
        <span className="text-[11px] font-medium text-fg-muted">You</span>
        <BarGroup level={mic.level} barColor="bg-success" idleColor="bg-success/25" />
      </div>
      <div className="h-5 w-px bg-border" />
      <div className="flex flex-1 items-center gap-2.5">
        <span className="text-[11px] font-medium text-fg-muted">Gemini</span>
        <BarGroup level={model.level} barColor="bg-info" idleColor="bg-info/25" />
      </div>
    </div>
  );
}
