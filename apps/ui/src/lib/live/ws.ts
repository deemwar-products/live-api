import type {
 AudioInPayload,
 ClientMsg,
 EndPayload,
 InterruptPayload,
 PingPayload,
 ServerMsg,
 StartPayload,
} from "@/lib/live/protocol";
import { PROTOCOL_VERSION } from "@/lib/live/protocol";

/**
 * Tiny WebSocket client for /v1/live.
 *
 * - One browser session == one WS.
 * - No auto-reconnect (per SPEC); the page decides when to retry.
 * - The server is the source of truth for state transitions; this
 * client just forwards messages and exposes typed send helpers.
 */

type ServerHandler = (msg: ServerMsg) => void;
type OpenHandler = () => void;
type CloseHandler = (ev: CloseEvent) => void;
type ErrorHandler = (err: Event) => void;

export interface LiveClient {
 sendStart(p?: StartPayload): void;
 sendAudioIn(p: AudioInPayload): void;
 sendInterrupt(p: InterruptPayload): void;
 sendEnd(p?: EndPayload): void;
 sendPing(p?: PingPayload): void;
 close(): void;
 onServerMessage(h: ServerHandler): () => void;
 onOpen(h: OpenHandler): () => void;
 onClose(h: CloseHandler): () => void;
 onError(h: ErrorHandler): () => void;
 readonly state: "idle" | "connecting" | "open" | "closed";
}

function newId(): string {
 if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
 return crypto.randomUUID();
 }
 return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

export function connectLive(url: string): LiveClient {
 const ws = new WebSocket(url);
 ws.binaryType = "arraybuffer";

 const serverHandlers = new Set<ServerHandler>();
 const openHandlers = new Set<OpenHandler>();
 const closeHandlers = new Set<CloseHandler>();
 const errorHandlers = new Set<ErrorHandler>();

 let state: LiveClient["state"] = "connecting";

 ws.addEventListener("open", () => {
 state = "open";
 openHandlers.forEach((h) => h());
 });

 ws.addEventListener("message", (ev) => {
 if (typeof ev.data !== "string") return; // we only accept text/JSON
 let parsed: ServerMsg;
 try {
 parsed = JSON.parse(ev.data) as ServerMsg;
 } catch {
 return; // ignore malformed frames; server logs the bad_message
 }
 serverHandlers.forEach((h) => h(parsed));
 });

 ws.addEventListener("close", (ev) => {
 state = "closed";
 closeHandlers.forEach((h) => h(ev));
 });

 ws.addEventListener("error", (ev) => {
 errorHandlers.forEach((h) => h(ev));
 });

 function send<T extends ClientMsg["type"]>(
 type: T,
 payload: Extract<ClientMsg, { type: T }>["payload"],
 ): void {
 if (ws.readyState !== WebSocket.OPEN) return;
 const msg = {
 v: PROTOCOL_VERSION,
 type,
 id: newId(),
 ts: Date.now(),
 payload,
 };
 ws.send(JSON.stringify(msg));
 }

 return {
 state,
 sendStart(p = {}) {
 send("start", p);
 },
 sendAudioIn(p: AudioInPayload) {
 send("audio_in", p);
 },
 sendInterrupt(p: InterruptPayload) {
 send("interrupt", p);
 },
 sendEnd(p = { reason: "user_ended" }) {
 send("end", p);
 },
 sendPing(p = {}) {
 send("ping", p);
 },
 close() {
 try {
 ws.close(1000, "client_ended");
 } catch {
 // ignore
 }
 },
 onServerMessage(h) {
 serverHandlers.add(h);
 return () => serverHandlers.delete(h);
 },
 onOpen(h) {
 openHandlers.add(h);
 return () => openHandlers.delete(h);
 },
 onClose(h) {
 closeHandlers.add(h);
 return () => closeHandlers.delete(h);
 },
 onError(h) {
 errorHandlers.add(h);
 return () => errorHandlers.delete(h);
 },
 };
}
