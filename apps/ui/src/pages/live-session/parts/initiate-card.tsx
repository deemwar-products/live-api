/**
 * Initiate card — shown when the session is idle or connecting.
 * A single CTA opens the WS, requests mic permission, and starts
 * the live session. Copy and CTA label resolve through the page
 * labels so they stay in lockstep.
 */

import { Mic } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { LIVE_SESSION_PAGE_LABELS } from "@/labels/live-session";

export function InitiateCard({
 status,
 onStart,
}: {
 status: "idle" | "connecting";
 onStart: () => void;
}) {
 const isConnecting = status === "connecting";
 return (
 <Card className="mx-auto w-full max-w-xl">
 <CardBody className="flex flex-col items-center gap-4 py-10 text-center">
 <span className="inline-flex size-12 items-center justify-center rounded-full bg-accent-soft text-accent-strong">
 <Mic className="size-5" strokeWidth={1.75} />
 </span>
 <div className="space-y-1">
 <p className="text-[11px] font-semibold uppercase tracking-wider text-accent">
 {LIVE_SESSION_PAGE_LABELS.initiate.eyebrow}
 </p>
 <h2 className="text-xl font-semibold tracking-tight text-fg">
 {LIVE_SESSION_PAGE_LABELS.initiate.title}
 </h2>
 <p className="mx-auto max-w-md text-sm text-fg-muted">
 {LIVE_SESSION_PAGE_LABELS.initiate.body}
 </p>
 </div>
 <Button
 size="lg"
 onClick={onStart}
 disabled={isConnecting}
 leading={<Mic className="size-4" />}
 >
 {isConnecting
 ? LIVE_SESSION_PAGE_LABELS.initiate.ctaConnecting
 : LIVE_SESSION_PAGE_LABELS.initiate.ctaIdle}
 </Button>
 <p className="text-[11px] text-fg-subtle">
 {LIVE_SESSION_PAGE_LABELS.initiate.permissionHint}
 </p>
 </CardBody>
 </Card>
 );
}
