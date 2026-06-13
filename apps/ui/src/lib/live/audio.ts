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
 this.frameCount = 0;
 // Tell the main thread we're alive.
 this.port.postMessage({ kind: "ready" });
 }

 process(inputs) {
 const input = inputs[0];
 if (!input || input.length === 0) return true;
 const channel = input[0];
 if (!channel) return true;
 this.frameCount++;
 if (this.frameCount === 1 || this.frameCount % 250 === 0) {
 this.port.postMessage({ kind: "tick", frame: this.frameCount });
 }
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
 }
 }
 if (isSpeech && this.silenceFrames === 0) {
 // Emit "speech" only on the rising edge of a speech burst — the
 // main thread ignores subsequent speech messages until silence
 // arrives, so this naturally debounces.
 this.port.postMessage({ kind: "speech" });
 }
 this.port.postMessage({ kind: "pcm", buffer: copy });
 }
}
registerProcessor("pcm-emitter", PCMEmitter);
`;

export async function startCapture(cb: CaptureCallbacks): Promise<CaptureHandle> {
 const stream = await navigator.mediaDevices.getUserMedia({
 audio: {
 sampleRate: SAMPLE_RATE_IN,
 channelCount: CHANNELS,
 echoCancellation: true,
 noiseSuppression: true,
 },
 });

 const ctx = new AudioContext({ sampleRate: SAMPLE_RATE_IN });
 // The AudioContext may start suspended; resume it so the worklet
 // processes. Browsers require a user-gesture for this, but the click
 // on "Start live session" counts.
 if (ctx.state === "suspended") {
 await ctx.resume();
 }

 const blob = new Blob([PCM_WORKLET_SOURCE], { type: "application/javascript" });
 const url = URL.createObjectURL(blob);
 await ctx.audioWorklet.addModule(url);
 URL.revokeObjectURL(url);

 const source = ctx.createMediaStreamSource(stream);
 const node = new AudioWorkletNode(ctx, "pcm-emitter", {
 // Keep the node alive even if no downstream connection is made.
 processorOptions: {},
 });
 source.connect(node);
 // Important: an AudioWorkletNode whose output is not connected to
 // `ctx.destination` may be silently suspended in some browsers. We
 // wire a muted gain to keep the audio graph "live" without feeding
 // the mic back to the speakers (which would cause a feedback loop).
 const sink = ctx.createGain();
 sink.gain.value = 0;
 node.connect(sink);
 sink.connect(ctx.destination);

 const samplesPerFrame = (SAMPLE_RATE_IN * CAPTURE_FRAME_MS) / 1000;
 let buffer = new Float32Array(samplesPerFrame * 4);
 let buffered = 0;

 node.port.onmessage = (ev: MessageEvent<{ kind: string; buffer?: Float32Array; frame?: number }>) => {
 const msg = ev.data;
 if (msg.kind === "ready") {
 console.log("[live] worklet ready");
 return;
 }
 if (msg.kind === "tick") {
 console.log(`[live] worklet frame ${msg.frame}`);
 return;
 }
 if (msg.kind === "speech") {
 cb.onSpeechStart?.();
 return;
 }
 if (msg.kind !== "pcm" || !msg.buffer) return;

 const chunk = msg.buffer;
 const needed = samplesPerFrame;
 if (buffered + chunk.length < needed) {
 // Grow buffer if needed.
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
 // Keep ctx running by playing silence if it gets suspended on user-gesture boundaries.
 const queue: AudioBufferSourceNode[] = [];

 function playNext() {
 const next = queue.shift();
 if (!next) return;
 next.start();
 next.onended = () => {
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
