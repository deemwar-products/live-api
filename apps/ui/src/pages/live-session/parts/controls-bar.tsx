/**
 * Controls bar — mute toggle + end-session button. Rounded-full
 * buttons matching the design system, with a destructive end
 * button.
 */

import { Mic, MicOff, PhoneOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";

export function ControlsBar({
 muted,
 onToggleMute,
 onEnd,
}: {
 muted: boolean;
 onToggleMute: () => void;
 onEnd: () => void;
}) {
 return (
 <div className="flex items-center justify-center gap-2">
 <Button
 variant="secondary"
 onClick={onToggleMute}
 leading={muted ? <MicOff className="size-4" /> : <Mic className="size-4" />}
 >
 {muted
 ? LIVE_SESSION_PAGE_LABELS.controls.unmute
 : LIVE_SESSION_PAGE_LABELS.controls.mute}
 </Button>
 <Button
 variant="primary"
 tone="inverted"
 onClick={onEnd}
 leading={<PhoneOff className="size-4" />}
 >
 {LIVE_SESSION_PAGE_LABELS.controls.end}
 </Button>
 </div>
 );
}
