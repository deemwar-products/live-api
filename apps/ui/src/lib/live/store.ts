/**
 * Local live session store — no Zustand, no external dep.
 *
 * A small pub/sub store that lives in memory. We *also* mirror
 * important bits to localStorage so a page reload mid-session can
 * recover the session id + last status. Audio buffers and live
 * transcript are not persisted (they expire on close).
 *
 * Public surface is intentionally tiny: getState, setState, subscribe,
 * plus typed action helpers (setStatus, appendTranscript, etc).
 */

import type {
 ErrorPayload,
 ReadyPayload,
 StatusState,
 TranscriptPayload,
} from "@/lib/live/protocol";

export interface TranscriptEntry {
 turnId: string;
 role: "user" | "model";
 text: string; // current best text for this turn
 turnComplete: boolean;
 ts: number;
}

export interface SessionRecord {
 sessionId?: string;
 model?: string;
 startedAt?: number;
 endedAt?: number;
 /** Last observed coarse status, including "idle". Slightly wider than
 * the wire StatusState because the store has its own "idle" anchor. */
 lastStatus?: "idle" | StatusState;
 lastError?: { code: string; message: string };
}

export interface LiveState {
 status: "idle" | "connecting" | "live" | "interrupted" | "ended";
 subState: "idle" | "listening" | "speaking" | "paused" | "dropped";
 ready: ReadyPayload | null;
 transcript: TranscriptEntry[];
 session: SessionRecord;
 micMuted: boolean;
 /** True while the mic VAD is reporting speech. Drives the waveform. */
 micActive: boolean;
 /** True while model audio_out frames are arriving. Drives the waveform. */
 modelActive: boolean;
}

const STORAGE_KEY = "live-api:session";

const initial: LiveState = {
 status: "idle",
 subState: "idle",
 ready: null,
 transcript: [],
 session: {},
 micMuted: false,
 micActive: false,
 modelActive: false,
};

type Listener = (s: LiveState) => void;

class LiveStore {
 private state: LiveState = initial;
 private listeners = new Set<Listener>();

 getState(): LiveState {
 return this.state;
 }

 setState(patch: Partial<LiveState> | ((s: LiveState) => Partial<LiveState>)): void {
 const next =
 typeof patch === "function" ? patch(this.state) : patch;
 this.state = { ...this.state, ...next };
 this.persist();
 this.listeners.forEach((l) => l(this.state));
 }

 subscribe(l: Listener): () => void {
 this.listeners.add(l);
 return () => this.listeners.delete(l);
 }

 // ---------- typed action helpers ----------

 setStatus(status: LiveState["status"]): void {
 const session: SessionRecord = { ...this.state.session };
 if (status === "live" && !session.startedAt) {
 session.startedAt = Date.now();
 }
 if (status === "ended" && !session.endedAt) {
 session.endedAt = Date.now();
 }
 session.lastStatus = status;

 // Reconcile sub-state with coarse status so the pill always reads true.
 let subState: LiveState["subState"] = this.state.subState;
 if (status === "ended" || status === "idle" || status === "connecting") {
 subState = status === "connecting" ? this.state.subState : "idle";
 } else if (status === "live") {
 subState = this.state.micMuted ? "paused" : "listening";
 } else if (status === "interrupted") {
 subState = this.state.micMuted ? "paused" : "dropped";
 }
 this.setState({ status, session, subState });
 }

 setReady(ready: ReadyPayload): void {
 this.setState({
 ready,
 session: { ...this.state.session, sessionId: ready.sessionId, model: ready.model },
 });
 }

 appendTranscript(p: TranscriptPayload): void {
 const list = this.state.transcript.slice();
 const last = list[list.length - 1];

 if (last && last.turnId === p.turnId && !last.turnComplete) {
 // Update the in-flight turn.
 list[list.length - 1] = {
 ...last,
 text: p.text,
 turnComplete: p.turnComplete,
 ts: Date.now(),
 };
 } else {
 list.push({
 turnId: p.turnId,
 role: p.role,
 text: p.text,
 turnComplete: p.turnComplete,
 ts: Date.now(),
 });
 }
 this.setState({ transcript: list });
 }

 setError(p: ErrorPayload): void {
 this.setState({
 session: {
 ...this.state.session,
 lastError: { code: p.code, message: p.message },
 },
 });
 }

 setMicMuted(muted: boolean): void {
 // Mute flips sub-state to "paused" so the pill reflects reality. We only
 // flip when the session is in a state that cares (live / interrupted) —
 // the idle/ended states keep their subState untouched.
 const cur = this.state;
 const next: Partial<LiveState> = { micMuted: muted };
 if (muted) next.micActive = false;
 if (muted && (cur.status === "live" || cur.status === "interrupted")) {
 next.subState = "paused";
 } else if (!muted && cur.subState === "paused") {
 next.subState = cur.status === "live" ? "listening" : "idle";
 }
 this.setState(next);
 }

 setMicActive(active: boolean): void {
 if (this.state.micMuted) return;
 if (this.state.micActive === active) return;
 this.setState({ micActive: active });
 }

 setModelActive(active: boolean): void {
 if (this.state.modelActive === active) return;
 this.setState({ modelActive: active });
 }

 // ---------- sub-state helpers ----------

 /**
  * Mark that the model is producing audio. Called every time a new
  * audio_out frame arrives. The sub-state moves to "speaking" if the
  * session is live. Use clearModelAudio when audio stops for >300ms.
  */
 noteModelAudioActive(): void {
 if (this.state.status !== "live") return;
 if (this.state.subState === "speaking") return;
 this.setState({ subState: "speaking" });
 }

 /**
  * Mark that model audio has stopped (no new audio_out for ~300ms or
  * the server signaled turnComplete / interrupted). Sub-state returns
  * to "listening" unless muted or session is no longer live.
  */
 noteModelAudioIdle(): void {
 if (this.state.subState !== "speaking") return;
 const next = this.deriveListeningSubState();
 this.setState({ subState: next });
 }

 /**
  * Mark the user is currently speaking (VAD fired). Only used as a
  * hint to flip the sub-state back to "listening" after a model turn.
  */
 noteUserActivity(): void {
 if (this.state.micMuted) return;
 if (this.state.status !== "live" && this.state.status !== "interrupted") return;
 if (this.state.subState === "speaking") return;
 this.setState({ subState: "listening" });
 }

 /**
  * Server told us we were interrupted. Sub-state moves to "dropped"
  * (the pill says "Interrupted — listening again").
  */
 noteInterrupted(): void {
 this.setState({ subState: this.state.micMuted ? "paused" : "dropped" });
 }

 /**
  * User-initiated interrupt via barge-in or stop button. Same effect
  * as noteInterrupted from the UI's perspective.
 */
 noteClientInterrupt(): void {
 this.noteInterrupted();
 }

 /**
  * Derive the "what should the sub-state be when not speaking"
  * answer from the current coarse status + mute.
  */
 private deriveListeningSubState(): "listening" | "paused" | "idle" {
 if (this.state.status !== "live" && this.state.status !== "interrupted") {
 return "idle";
 }
 if (this.state.micMuted) return "paused";
 return "listening";
 }

 reset(): void {
 this.state = initial;
 this.persist();
 this.listeners.forEach((l) => l(this.state));
 }

 // ---------- localStorage mirror ----------

 private persist(): void {
 if (typeof window === "undefined") return;
 try {
 const { session, micMuted, status } = this.state;
 window.localStorage.setItem(
 STORAGE_KEY,
 JSON.stringify({ session, micMuted, status }),
 );
 } catch {
 // Quota or disabled — non-fatal.
 }
 }
}

export const liveStore = new LiveStore();

// Hydrate from localStorage on first import.
if (typeof window !== "undefined") {
 try {
 const raw = window.localStorage.getItem(STORAGE_KEY);
 if (raw) {
 const parsed = JSON.parse(raw) as Partial<LiveState>;
 if (parsed.session) {
 liveStore.setState({ session: parsed.session });
 }
 if (typeof parsed.micMuted === "boolean") {
 liveStore.setState({ micMuted: parsed.micMuted });
 }
 }
 } catch {
 // ignore
 }
}
