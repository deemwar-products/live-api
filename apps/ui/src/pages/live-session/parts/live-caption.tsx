/**
 * Live caption — shows the user's in-progress speech transcript with a
 * pulsing "speaking" indicator. Only renders while the mic is active or
 * there's an unfrozen "user" transcript entry to show.
 */

import { useEffect, useState } from "react";
import { cn } from "@/lib/cn";
import { liveStore } from "@/lib/live/store";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";

function useUserCaption(): { text: string; active: boolean } {
 const derive = () => {
 const s = liveStore.getState();
 const last = s.transcript[s.transcript.length - 1];
 const text = last && last.role === "user" && !last.turnComplete ? last.text : "";
 return { text, active: s.micActive };
 };
 const [state, setState] = useState(derive);
 useEffect(() => liveStore.subscribe(() => setState(derive())), []);
 return state;
}

export function LiveCaption({ className }: { className?: string }) {
 const { text, active } = useUserCaption();

 if (!text && !active) return null;

 return (
 <div
 className={cn(
 "flex items-center gap-2 rounded-lg bg-bg-muted px-3 py-2 text-sm text-fg-muted",
 className,
 )}
 >
 <span className="relative flex size-2 shrink-0">
 {active && (
 <span className="absolute inline-flex size-full animate-ping rounded-full bg-success opacity-75" />
 )}
 <span className={cn("relative inline-flex size-2 rounded-full", active ? "bg-success" : "bg-fg-subtle")} />
 </span>
 <p className="flex-1 truncate">
 {text || LIVE_SESSION_PAGE_LABELS.caption.listening}
 </p>
 </div>
 );
}
