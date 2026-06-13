/**
 * Live session lifecycle hook. Owns the WS client, the capture and
 * playback handles, and the glue between the store and the network.
 *
 * The page component just calls useLiveSession() and reads state.
 * All teardown is automatic on unmount or when end() is called.
 */

import { useCallback, useEffect, useRef, useState } from "react";
import { connectLive, type LiveClient } from "@/lib/live/ws";
import { liveStore } from "@/lib/live/store";
import { startCapture, startPlayback, type CaptureHandle, type PlaybackHandle } from "@/lib/live/audio";
import type { ErrorPayload, ServerMsg, StatusState } from "@/lib/live/protocol";

const WS_URL =
 (import.meta.env.VITE_API_WS_URL as string | undefined) ?? "ws://localhost:8080/v1/live";

function wsUrl(): string {
 return WS_URL;
}

function mapError(p: ErrorPayload): { fatal: boolean } {
 return { fatal: p.fatal };
}

export function useLiveSession() {
 const [status, setStatus] = useState<StatusState | "idle">(() => liveStore.getState().status);
 const [error, setError] = useState<{ code: string; message: string } | null>(null);
 const clientRef = useRef<LiveClient | null>(null);
 const captureRef = useRef<CaptureHandle | null>(null);
 const playbackRef = useRef<PlaybackHandle | null>(null);
 const mutedRef = useRef<boolean>(false);

 // Mirror store status into local React state.
 useEffect(() => {
 return liveStore.subscribe((s) => setStatus(s.status));
 }, []);

 const start = useCallback(async () => {
 if (status === "connecting" || status === "live") return;
 liveStore.reset();
 liveStore.setStatus("connecting");
 setError(null);

 const client = connectLive(wsUrl());
 clientRef.current = client;

 // ----- Server messages -----
 const offMsg = client.onServerMessage((msg: ServerMsg) => {
 switch (msg.type) {
 case "ready":
 liveStore.setReady(msg.payload);
 liveStore.setStatus("live");
 // Start capture + playback after Ready.
 void beginMedia();
 break;
 case "audio_out":
 playbackRef.current?.enqueue(msg.payload.pcm);
 break;
 case "transcript":
 liveStore.appendTranscript(msg.payload);
 break;
 case "status":
 liveStore.setStatus(msg.payload.state);
 break;
 case "error": {
 const { fatal } = mapError(msg.payload);
 liveStore.setError(msg.payload);
 setError({ code: msg.payload.code, message: msg.payload.message });
 if (fatal) {
 teardown();
 }
 break;
 }
 }
 });

 client.onClose(() => {
 teardown();
 liveStore.setStatus("ended");
 });

 // Kick the session off; server is permissive — start is a no-op today.
 client.onOpen(() => client.sendStart({}));

 async function beginMedia() {
 const muted = mutedRef.current;
 try {
 const capture = await startCapture({
 onFrame: (b64, durationMs) => {
 if (muted) return;
 clientRef.current?.sendAudioIn({
 pcm: b64,
 sampleRate: 16000,
 encoding: "pcm_s16le",
 channels: 1,
 durationMs,
 });
 },
 onError: (err) => {
 setError({ code: "mic_error", message: err.message });
 liveStore.setError({ code: "mic_error", message: err.message, fatal: false });
 },
 });
 captureRef.current = capture;
 } catch (e) {
 const message = e instanceof Error ? e.message : "microphone unavailable";
 setError({ code: "mic_denied", message });
 liveStore.setError({ code: "mic_denied", message, fatal: false });
 // Don't tear down the WS — let the user see Gemini's text even without mic.
 }
 playbackRef.current = startPlayback();
 }

 function teardown() {
 offMsg();
 try {
 clientRef.current?.close();
 } catch {
 // ignore
 }
 clientRef.current = null;
 try {
 captureRef.current?.stop();
 } catch {
 // ignore
 }
 captureRef.current = null;
 try {
 playbackRef.current?.stop();
 } catch {
 // ignore
 }
 playbackRef.current = null;
 }

 // end is exposed via the ref so onClose can call it.
 (clientRef as unknown as { _teardown?: () => void })._teardown = teardown;
 }, [status]);

 const end = useCallback(() => {
 try {
 clientRef.current?.sendEnd({ reason: "user_ended" });
 } catch {
 // ignore
 }
 try {
 captureRef.current?.stop();
 } catch {
 // ignore
 }
 captureRef.current = null;
 try {
 playbackRef.current?.stop();
 } catch {
 // ignore
 }
 playbackRef.current = null;
 try {
 clientRef.current?.close();
 } catch {
 // ignore
 }
 clientRef.current = null;
 liveStore.setStatus("ended");
 }, []);

 const toggleMute = useCallback(() => {
 const next = !mutedRef.current;
 mutedRef.current = next;
 liveStore.setMicMuted(next);
 }, []);

 // Clean up on unmount.
 useEffect(() => {
 return () => {
 try {
 captureRef.current?.stop();
 } catch {
 // ignore
 }
 try {
 playbackRef.current?.stop();
 } catch {
 // ignore
 }
 try {
 clientRef.current?.close();
 } catch {
 // ignore
 }
 };
 }, []);

 return { status, error, start, end, toggleMute };
}
