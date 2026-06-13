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
 lastStatus?: StatusState;
 lastError?: { code: string; message: string };
}

export interface LiveState {
 status: "idle" | "connecting" | "live" | "interrupted" | "ended";
 ready: ReadyPayload | null;
 transcript: TranscriptEntry[];
 session: SessionRecord;
 micMuted: boolean;
}

const STORAGE_KEY = "live-api:session";

const initial: LiveState = {
 status: "idle",
 ready: null,
 transcript: [],
 session: {},
 micMuted: false,
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
 this.setState({ status, session });
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
 this.setState({ micMuted: muted });
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
