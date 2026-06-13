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
const TARGET_RATE = 16000;
const FRAME_MS = 20;
const SAMPLES_PER_FRAME = TARGET_RATE * FRAME_MS / 1000; // 320

class PCMEmitter extends AudioWorkletProcessor {
  constructor() {
    super();
    this.silenceFrames = 0;
    this.warnedNoInput = false;
    this.speechArmed = true;
    this.frameCount = 0;
    this.outBuffer = new Float32Array(SAMPLES_PER_FRAME * 4);
    this.outBuffered = 0;
    this.port.postMessage({ kind: "ready" });
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
    this.frameCount++;
    if (this.frameCount === 1 || this.frameCount % 250 === 0) {
      this.port.postMessage({ kind: "tick", frame: this.frameCount, ctxRate: sampleRate });
    }

    // Resample from sampleRate to 16000. We use a 3-tap averaging
    // decimation for the 48k -> 16k case (each output is the mean of 3
    // input samples, which acts as a simple low-pass filter to avoid
    // aliasing). For other rates, linear interpolation.
    const ratio = sampleRate / TARGET_RATE;
    let copy;
    if (Math.abs(ratio - 1) < 0.01) {
      copy = new Float32Array(channel);
    } else if (Math.abs(ratio - 3) < 0.05) {
      // 48k -> 16k: average every 3 samples. Acts as a 3-tap box filter
      // with cutoff around sampleRate/(2*3) ≈ 8kHz — adequate for speech.
      const groups = Math.floor(channel.length / 3);
      copy = new Float32Array(groups);
      for (let g = 0; g < groups; g++) {
        const i = g * 3;
        copy[g] = (channel[i] + channel[i + 1] + channel[i + 2]) / 3;
      }
    } else {
      // Generic linear interpolation fallback.
      const len = Math.floor(channel.length / ratio);
      copy = new Float32Array(len);
      for (let i = 0; i < len; i++) {
        const srcIdx = i * ratio;
        const lo = Math.floor(srcIdx);
        const hi = Math.min(lo + 1, channel.length - 1);
        const frac = srcIdx - lo;
        copy[i] = channel[lo] * (1 - frac) + channel[hi] * frac;
      }
    }

    // RMS on the resampled block.
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

    // Accumulate into outBuffer and emit fixed-size frames.
    if (this.outBuffered + copy.length > this.outBuffer.length) {
      const grown = new Float32Array(Math.max(this.outBuffer.length * 2, this.outBuffered + copy.length));
      grown.set(this.outBuffer.subarray(0, this.outBuffered));
      this.outBuffer = grown;
    }
    this.outBuffer.set(copy, this.outBuffered);
    this.outBuffered += copy.length;

    while (this.outBuffered >= SAMPLES_PER_FRAME) {
      const frame = this.outBuffer.slice(0, SAMPLES_PER_FRAME);
      this.port.postMessage({ kind: "pcm", buffer: frame });
      this.outBuffer.copyWithin(0, SAMPLES_PER_FRAME, this.outBuffered);
      this.outBuffered -= SAMPLES_PER_FRAME;
    }
    return true;
  }
}
registerProcessor("pcm-emitter", PCMEmitter);
`;

export async function startCapture(cb: CaptureCallbacks): Promise<CaptureHandle> {
  console.log("[live] startCapture: beginning mic request");
  const stream = await navigator.mediaDevices.getUserMedia({
    audio: {
      // `ideal` (not a hard value) — most hardware mics don't natively
      // support 16kHz and would throw OverconstrainedError on a hard
      // constraint. The worklet resamples to SAMPLE_RATE_IN regardless
      // of what the device/context delivers.
      sampleRate: { ideal: SAMPLE_RATE_IN },
      channelCount: CHANNELS,
      echoCancellation: true,
      noiseSuppression: true,
    },
  });
  console.log("[live] getUserMedia granted");

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
    // The worklet resamples to 16kHz internally regardless of the actual
    // context rate, so a mismatch here is informational only.
    logger.info("live: audio context running at non-standard rate — worklet will resample", {
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
  console.log("[live] AudioWorkletNode created");
  source.connect(node);
  // Important: an AudioWorkletNode whose output is not connected to
  // `ctx.destination` may be silently suspended in some browsers. Wire a
  // muted gain to keep the audio graph "live" without feeding the mic
  // back to the speakers (which would cause a feedback loop).
  const sink = ctx.createGain();
  sink.gain.value = 0;
  node.connect(sink);
  sink.connect(ctx.destination);

  node.port.onmessage = (ev: MessageEvent<{ kind: string; buffer?: Float32Array; frame?: number; ctxRate?: number }>) => {
    const msg = ev.data;
    if (msg.kind === "ready") {
      console.log("[live] worklet ready");
      return;
    }
    if (msg.kind === "tick") {
      console.log(`[live] worklet frame ${msg.frame} ctxRate=${msg.ctxRate}`);
      return;
    }
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

    // Worklet emits fixed-size 20ms frames at 16kHz — encode and ship.
    const int16 = floatToInt16(msg.buffer);
    cb.onFrame(int16ToBase64(int16), CAPTURE_FRAME_MS);
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
    // Resume the context on first audio to avoid autoplay restrictions.
    if (ctx.state === "suspended") {
      ctx.resume().then(() => next.start());
    } else {
      next.start();
    }
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
