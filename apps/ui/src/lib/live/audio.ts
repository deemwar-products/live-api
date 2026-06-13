/**
 * Browser audio: 16 kHz mono Int16 PCM capture (mic) and 24 kHz mono
 * Int16 PCM playback (model). PCM is base64-encoded on the wire — see
 * `protocol.ts`.
 *
 * Capture pipeline:
 * getUserMedia → AudioContext(16kHz) → AudioWorklet(pcm-emitter)
 * → main thread encodes Float32 → Int16 → base64 → WS frame every ~20 ms.
 *
 * Playback pipeline:
 * WS frame (b64 Int16) → Int16Array → Float32Array → AudioBuffer(24kHz)
 * → queue + small jitter buffer → destination.
 */

import {
 CHANNELS,
 ENCODING,
 SAMPLE_RATE_IN,
 SAMPLE_RATE_OUT,
} from "@/lib/live/protocol";
import { logger } from "@/lib/logger";

const CAPTURE_FRAME_MS = 20;
const PLAYBACK_JITTER_FRAMES = 3;

// ---------- base64 helpers ----------

function int16ToBase64(samples: Int16Array): string {
 const view = new Uint8Array(samples.buffer, samples.byteOffset, samples.byteLength);
 let binary = "";
 const chunk = 0x8000;
 for (let i = 0; i < view.length; i += chunk) {
 binary += String.fromCharCode.apply(
 null,
 Array.from(view.subarray(i, i + chunk)),
 );
 }
 return btoa(binary);
}

function base64ToInt16(b64: string): Int16Array {
 const binary = atob(b64);
 const view = new Uint8Array(binary.length);
 for (let i = 0; i < binary.length; i++) view[i] = binary.charCodeAt(i);
 return new Int16Array(view.buffer, view.byteOffset, view.byteLength / 2);
}

// ---------- capture ----------

export interface CaptureHandle {
 /** Flush any pending frame and stop the mic + worklet. */
 stop(): void;
}

export interface CaptureCallbacks {
 /** Called every ~20 ms with a base64 PCM frame. */
 onFrame(b64Pcm: string, durationMs: number): void;
 /**
 * Called the first time in a while the mic crosses the speech
 * threshold. Fires once per "speech onset" — re-arms only after a
 * silence gap of ~400ms.
 */
 onSpeechStart?(): void;
 /** Called on getUserMedia or worklet failure. */
 onError(err: Error): void;
}

const PCM_WORKLET_SOURCE = `
const SPEECH_RMS_THRESHOLD = 0.02;
const SPEECH_END_FRAMES = 20; // ~400ms at 20ms/frame before re-arming

class PCMEmitter extends AudioWorkletProcessor {
 constructor() {
 super();
 this.silenceFrames = 0;
 this.warnedNoInput = false;
 this.speechArmed = true;
 }

 process(inputs) {
 const input = inputs[0];
 if (!input || input.length === 0) {
 // The MediaStreamTrack delivered zero channels for this render
 // quantum — typically a virtual/loopback input device (e.g.
 // BlackHole) with no signal feeding it, rather than a real mic.
 // Warn once so this doesn't look like a silent infinite hang.
 if (!this.warnedNoInput) {
 this.warnedNoInput = true;
 this.port.postMessage({ kind: "no_input" });
 }
 return true;
 }
 const channel = input[0];
 if (!channel) return true;
 // Copy because the underlying buffer is reused.
 const copy = new Float32Array(channel);

 // RMS over this render quantum (128 samples at 16kHz ≈ 8ms).
 let sumSq = 0;
 for (let i = 0; i < copy.length; i++) sumSq += copy[i] * copy[i];
 const rms = Math.sqrt(sumSq / copy.length);

 let isSpeech = false;
 if (rms >= SPEECH_RMS_THRESHOLD) {
 this.silenceFrames = 0;
 isSpeech = true;
 } else {
 this.silenceFrames += 1;
 // Re-arm: only after ~400ms of silence will we fire another onset.
 if (this.silenceFrames === SPEECH_END_FRAMES) {
 // Tell the main thread that speech ended, so it can update UI.
 this.port.postMessage({ kind: "silence" });
 this.speechArmed = true;
 }
 }
 if (isSpeech && this.speechArmed) {
 // Emit "speech" only on the rising edge of a speech burst — without
 // the armed flag this fires on every ~8ms render quantum while the
 // user is talking, flooding the main thread with speech-start events.
 this.port.postMessage({ kind: "speech" });
 this.speechArmed = false;
 }
 this.port.postMessage({ kind: "pcm", buffer: copy });
 // A falsy return tells the browser this node is done and it can stop
 // calling process() — without this, Chrome calls process() exactly
 // once and the worklet goes silent forever (no more "pcm" messages).
 return true;
 }
}
registerProcessor("pcm-emitter", PCMEmitter);
`;

export async function startCapture(cb: CaptureCallbacks): Promise<CaptureHandle> {
 const stream = await navigator.mediaDevices.getUserMedia({
 audio: {
 // `ideal` (not a hard value) — most hardware mics don't natively
 // support 16kHz and would throw OverconstrainedError on a hard
 // constraint. The AudioContext below resamples to SAMPLE_RATE_IN
 // regardless of what the device delivers.
 sampleRate: { ideal: SAMPLE_RATE_IN },
 channelCount: CHANNELS,
 echoCancellation: true,
 noiseSuppression: true,
 },
 });

 const track = stream.getAudioTracks()[0];
 logger.info("live: mic track acquired", {
 label: track?.label,
 settings: track?.getSettings(),
 });

 const ctx = new AudioContext({ sampleRate: SAMPLE_RATE_IN });
 // Creating an AudioContext after `await getUserMedia(...)` breaks the
 // user-gesture chain in some browsers, so it can start "suspended" —
 // the worklet's process() callback then never fires and no audio_in
 // frames are ever produced (with no error). Resume explicitly.
 if (ctx.state === "suspended") {
 await ctx.resume();
 }
 logger.info("live: capture audio context", { state: ctx.state, sampleRate: ctx.sampleRate });
 if (ctx.sampleRate !== SAMPLE_RATE_IN) {
 // We tell the server every frame is SAMPLE_RATE_IN regardless — if the
 // browser ignored the requested rate, audio_in frames are mislabeled
 // and Gemini will hear sped-up/slowed-down audio.
 logger.warn("live: audio context sample rate mismatch — audio_in frames will be mislabeled", {
 requested: SAMPLE_RATE_IN,
 actual: ctx.sampleRate,
 });
 }
 const blob = new Blob([PCM_WORKLET_SOURCE], { type: "application/javascript" });
 const url = URL.createObjectURL(blob);
 await ctx.audioWorklet.addModule(url);
 URL.revokeObjectURL(url);
 logger.info("live: audio worklet module loaded");

 const source = ctx.createMediaStreamSource(stream);
 const node = new AudioWorkletNode(ctx, "pcm-emitter");
 source.connect(node);
 // The worklet never writes to its output, but it still needs to be
 // connected into the graph reaching `destination` — otherwise the
 // audio renderer treats it as inactive and process() is never called,
 // so no "pcm"/"speech" messages are posted. This routes silence only.
 node.connect(ctx.destination);

 const samplesPerFrame = (SAMPLE_RATE_IN * CAPTURE_FRAME_MS) / 1000;
 let buffer = new Float32Array(samplesPerFrame * 4);
 let buffered = 0;

 node.port.onmessage = (ev: MessageEvent<{ kind: string; buffer?: Float32Array }>) => {
 const msg = ev.data;
 if (msg.kind === "no_input") {
 cb.onError(
 new Error(
 "Selected microphone is delivering no audio channels — it may be a virtual/loopback device (e.g. BlackHole) with no input. Switch to a real microphone in your browser/OS sound settings.",
 ),
 );
 return;
 }
 if (msg.kind === "speech") {
 cb.onSpeechStart?.();
 return;
 }
 if (msg.kind !== "pcm" || !msg.buffer) return;

 const chunk = msg.buffer;
 const needed = samplesPerFrame;
 if (buffer.length < buffered + chunk.length) {
 // Grow buffer if it's too small to hold the incoming chunk.
 const grown = new Float32Array(Math.max(buffer.length * 2, buffered + chunk.length));
 grown.set(buffer.subarray(0, buffered));
 buffer = grown;
 }
 buffer.set(chunk, buffered);
 buffered += chunk.length;

 while (buffered >= needed) {
 const frame = buffer.subarray(0, needed);
 const int16 = floatToInt16(frame);
 cb.onFrame(int16ToBase64(int16), CAPTURE_FRAME_MS);
 // Shift remaining.
 buffer.copyWithin(0, needed, buffered);
 buffered -= needed;
 }
 };

 node.port.onmessageerror = () => cb.onError(new Error("worklet message error"));

 return {
 stop() {
 try {
 node.disconnect();
 source.disconnect();
 } catch {
 // ignore
 }
 stream.getTracks().forEach((t) => t.stop());
 ctx.close().catch(() => undefined);
 },
 };
}

function floatToInt16(input: Float32Array): Int16Array {
 const out = new Int16Array(input.length);
 for (let i = 0; i < input.length; i++) {
 const s = Math.max(-1, Math.min(1, input[i]));
 out[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
 }
 return out;
}

// ---------- playback ----------

export interface PlaybackHandle {
 /** Queue a base64 PCM frame (24 kHz mono Int16 LE). */
 enqueue(b64Pcm: string): void;
 /** Stop playback and clear the queue. */
 stop(): void;
}

export function startPlayback(): PlaybackHandle {
 const ctx = new AudioContext({ sampleRate: SAMPLE_RATE_OUT });
 // Same suspended-on-creation issue as startCapture — resume explicitly
 // so queued audio_out frames actually play.
 if (ctx.state === "suspended") {
 void ctx.resume();
 }
 logger.info("live: playback audio context", { state: ctx.state, sampleRate: ctx.sampleRate });
 const queue: AudioBufferSourceNode[] = [];
 let playing = false;

 function playNext() {
 // Guard against overlapping playback: enqueue() can call playNext()
 // while a source is still playing (e.g. queue length returns to 1
 // right after the previous frame started). Without this guard, each
 // new frame starts its own source on top of the current one.
 if (playing) return;
 const next = queue.shift();
 if (!next) return;
 playing = true;
 next.start();
 next.onended = () => {
 playing = false;
 if (queue.length > 0) playNext();
 };
 }

 return {
 enqueue(b64: string) {
 const int16 = base64ToInt16(b64);
 const f32 = new Float32Array(int16.length);
 for (let i = 0; i < int16.length; i++) f32[i] = int16[i] / (int16[i] < 0 ? 0x8000 : 0x7fff);
 const buf = ctx.createBuffer(CHANNELS, f32.length, SAMPLE_RATE_OUT);
 buf.copyToChannel(f32, 0);
 const src = ctx.createBufferSource();
 src.buffer = buf;
 src.connect(ctx.destination);
 queue.push(src);
 if (queue.length === 1 || queue.length === PLAYBACK_JITTER_FRAMES) playNext();
 },
 stop() {
 while (queue.length) {
 const n = queue.shift();
 try {
 n?.stop();
 } catch {
 // ignore
 }
 }
 ctx.close().catch(() => undefined);
 },
 };
}

// Silence the linter — encoding is asserted in protocol.ts.
void ENCODING;
