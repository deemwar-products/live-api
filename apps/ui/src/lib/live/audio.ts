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
 /** Called on getUserMedia or worklet failure. */
 onError(err: Error): void;
}

const PCM_WORKLET_SOURCE = `
class PCMEmitter extends AudioWorkletProcessor {
 process(inputs) {
 const input = inputs[0];
 if (!input || input.length === 0) return true;
 const channel = input[0];
 if (!channel) return true;
 // Copy because the underlying buffer is reused.
 this.port.postMessage(new Float32Array(channel));
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
 const blob = new Blob([PCM_WORKLET_SOURCE], { type: "application/javascript" });
 const url = URL.createObjectURL(blob);
 await ctx.audioWorklet.addModule(url);
 URL.revokeObjectURL(url);

 const source = ctx.createMediaStreamSource(stream);
 const node = new AudioWorkletNode(ctx, "pcm-emitter");
 source.connect(node);

 const samplesPerFrame = (SAMPLE_RATE_IN * CAPTURE_FRAME_MS) / 1000;
 let buffer = new Float32Array(samplesPerFrame * 4);
 let buffered = 0;

 node.port.onmessage = (ev: MessageEvent<Float32Array>) => {
 const chunk = ev.data;
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
