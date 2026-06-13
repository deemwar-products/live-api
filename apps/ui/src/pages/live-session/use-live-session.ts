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
import { logger } from "@/lib/logger";

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
 console.log("[live] start: connecting WS to", wsUrl());
 liveStore.reset();
 liveStore.setStatus("connecting");
 setError(null);

 logger.info("live: connecting", { url: wsUrl() });
 const client = connectLive(wsUrl());
 clientRef.current = client;

 // ----- Server messages -----
 let ignoreUntil = 0;
 let audioOutFrames = 0;
 let audioInFrames = 0;
 let frameDebugCount = 0;
 const offMsg = client.onServerMessage((msg: ServerMsg) => {
 switch (msg.type) {
 case "ready":
 logger.info("live: ready", { ...msg.payload });
 liveStore.setReady(msg.payload);
 liveStore.setStatus("live");
 // Start capture + playback after Ready.
 void beginMedia();
 break;
 case "audio_out": {
 audioOutFrames += 1;
 if (audioOutFrames === 1 || audioOutFrames % 50 === 0) {
 logger.info("live: audio_out", {
 frame: audioOutFrames,
 bytes: msg.payload.pcm.length,
 durationMs: msg.payload.durationMs,
 final: msg.payload.final,
 });
 }
 // Drop audio_out frames received during the post-interrupt ignore
 // window so the queued model audio doesn't keep playing.
 if (performance.now() < ignoreUntil) {
 break;
 }
 playbackRef.current?.enqueue(msg.payload.pcm);
 liveStore.noteModelAudioActive();
 liveStore.setModelActive(true);
 // Mark "speaking" cleared shortly after the last frame; we'll reset
 // the timer on every new frame.
 scheduleModelIdle();
 break;
 }
 case "transcript":
 logger.info("live: transcript", { ...msg.payload });
 liveStore.appendTranscript(msg.payload);
 // Once Gemini's first transcript chunk arrives, user has heard
 // something — flag user activity to drop sub-state back to listening
 // when model finishes.
 if (msg.payload.role === "user") liveStore.noteUserActivity();
 break;
 case "status": {
 logger.info("live: status", { ...msg.payload });
 const prevStatus = liveStore.getState().status;
 liveStore.setStatus(msg.payload.state);
 // Server told us we were interrupted: drop queued audio and
 // start a brief ignore window for in-flight frames.
 if (msg.payload.state === "interrupted" && prevStatus !== "interrupted") {
 playbackRef.current?.stop();
 playbackRef.current = null;
 ignoreUntil = performance.now() + 500;
 liveStore.noteInterrupted();
 // Re-init playback so the next model turn can speak.
 playbackRef.current = startPlayback();
 } else if (msg.payload.state === "live" && prevStatus === "interrupted") {
 // Server resumed us; sub-state already reconciled by setStatus.
 }
 break;
 }
 case "error": {
 logger.error("live: server error", { ...msg.payload });
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

 client.onOpen(() => {
 logger.info("live: ws open");
 // Kick the session off; server is permissive — start is a no-op today.
 client.sendStart({});
 });

 client.onError((ev) => {
 logger.error("live: ws error", { event: ev.type });
 });

 client.onClose((ev) => {
 logger.info("live: ws closed", { code: ev.code, reason: ev.reason });
 teardown();
 liveStore.setStatus("ended");
 });

 async function beginMedia() {
 logger.info("live: requesting microphone");
 try {
 const capture = await startCapture({
 onFrame: (b64, durationMs) => {
 audioInFrames += 1;
 if (audioInFrames === 1 || audioInFrames % 50 === 0) {
 logger.info("live: audio_in", { frame: audioInFrames, bytes: b64.length, durationMs });
 }
 if (mutedRef.current) return;
 if (frameDebugCount < 3) {
 console.log(`[live] audio_in frame #${frameDebugCount++} bytes=${b64.length} durMs=${durationMs}`);
 } else if (frameDebugCount === 3) {
 console.log("[live] audio_in: further frames suppressed (set frameDebugCount=Infinity in console to re-enable)");
 frameDebugCount = Infinity;
 }
 clientRef.current?.sendAudioIn({
 pcm: b64,
 sampleRate: 16000,
 encoding: "pcm_s16le",
 channels: 1,
 durationMs,
 });
 },
 onSpeechStart: () => {
 logger.info("live: speech start (VAD)");
 // VAD fired — flip sub-state to listening unless the model is
 // currently speaking, in which case barge in.
 const sub = liveStore.getState().subState;
 liveStore.setMicActive(true);
 if (sub === "speaking" && !mutedRef.current) {
 clientRef.current?.sendInterrupt({ reason: "user_barge_in" });
 liveStore.noteClientInterrupt();
 } else {
 liveStore.noteUserActivity();
 }
 // Decay the active flag after a short window if no further speech fires.
 // The worklet also posts a "silence" event after ~400ms; we mirror that
 // here to keep the waveform snappy.
 window.setTimeout(() => liveStore.setMicActive(false), 400);
 },
 onError: (err) => {
 logger.error("live: capture error", { message: err.message });
 setError({ code: "mic_error", message: err.message });
 liveStore.setError({ code: "mic_error", message: err.message, fatal: false });
 },
 });
 console.log("[live] capture started");
 captureRef.current = capture;
 logger.info("live: microphone capture started");
 } catch (e) {
 const message = e instanceof Error ? e.message : "microphone unavailable";
 logger.error("live: microphone capture failed", { message });
 setError({ code: "mic_denied", message });
 liveStore.setError({ code: "mic_denied", message, fatal: false });
 // Don't tear down the WS — let the user see Gemini's text even without mic.
 }
 playbackRef.current = startPlayback();
 }

 // modelIdleTimer debounces model-audio-stopped into a single sub-state
 // flip. Reset on every audio_out frame.
 let modelIdleTimer: number | null = null;
 function scheduleModelIdle() {
 if (modelIdleTimer != null) window.clearTimeout(modelIdleTimer);
 modelIdleTimer = window.setTimeout(() => {
 modelIdleTimer = null;
 liveStore.noteModelAudioIdle();
 liveStore.setModelActive(false);
 }, 300);
 }

 // Periodic summary so it's obvious from the console alone whether the
 // mic is producing frames and whether the model is responding — the
 // fastest way to tell "browser not sending" from "backend not replying".
 const heartbeat = window.setInterval(() => {
 const s = liveStore.getState();
 logger.info("live: heartbeat", {
 wsState: clientRef.current?.state,
 audioInFrames,
 audioOutFrames,
 micActive: s.micActive,
 modelActive: s.modelActive,
 subState: s.subState,
 });
 }, 2000);

 function teardown() {
 window.clearInterval(heartbeat);
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
