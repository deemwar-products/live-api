import { useCallback, useEffect, useRef, useState } from "react";
import { Mic, MicOff, Radio } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { Card, CardBody, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { COMMON_LABELS } from "@/labels";
import { cn } from "@/lib/cn";

const L = COMMON_LABELS.liveSession;

type SessionState = "idle" | "connecting" | "listening" | "speaking" | "error";

type ServerMsg =
  | { type: "audio"; data: string }
  | { type: "transcript"; data: string }
  | { type: "turn_end" }
  | { type: "error"; data: string };

const WS_URL = `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.hostname}:8080/ws`;

export function LiveSessionPanel() {
  const [state, setState] = useState<SessionState>("idle");
  const [transcript, setTranscript] = useState<string[]>([]);
  const [errorMsg, setErrorMsg] = useState<string>("");

  const wsRef = useRef<WebSocket | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioCtxRef = useRef<AudioContext | null>(null);
  const audioQueueRef = useRef<ArrayBuffer[]>([]);
  const playingRef = useRef(false);

  // ── audio playback ────────────────────────────────────────────────────────

  const playNextChunk = useCallback(async () => {
    if (playingRef.current || audioQueueRef.current.length === 0) return;
    playingRef.current = true;
    const chunk = audioQueueRef.current.shift()!;
    const ctx = audioCtxRef.current!;
    try {
      const buf = await ctx.decodeAudioData(chunk);
      const src = ctx.createBufferSource();
      src.buffer = buf;
      src.connect(ctx.destination);
      src.onended = () => {
        playingRef.current = false;
        playNextChunk();
      };
      src.start();
    } catch {
      playingRef.current = false;
      playNextChunk();
    }
  }, []);

  const enqueueAudio = useCallback(
    (rawBytes: string) => {
      // rawBytes are raw PCM bytes sent as a binary string
      const bytes = Uint8Array.from(rawBytes, (c) => c.charCodeAt(0));
      audioQueueRef.current.push(bytes.buffer);
      setState("speaking");
      playNextChunk();
    },
    [playNextChunk]
  );

  // ── WebSocket lifecycle ───────────────────────────────────────────────────

  const connect = useCallback(async () => {
    setState("connecting");
    setErrorMsg("");
    setTranscript([]);

    let stream: MediaStream;
    try {
      stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    } catch {
      setState("error");
      setErrorMsg(L.permissionDenied);
      return;
    }

    audioCtxRef.current = new AudioContext({ sampleRate: 16000 });

    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      setState("listening");
      const recorder = new MediaRecorder(stream, { mimeType: "audio/webm;codecs=opus" });
      mediaRecorderRef.current = recorder;

      recorder.ondataavailable = (e) => {
        if (e.data.size > 0 && ws.readyState === WebSocket.OPEN) {
          e.data.arrayBuffer().then((buf) => {
            const bytes = new Uint8Array(buf);
            const raw = String.fromCharCode(...bytes);
            ws.send(JSON.stringify({ type: "audio", data: raw }));
          });
        }
      };
      recorder.start(250); // 250 ms chunks
    };

    ws.onmessage = (e) => {
      let msg: ServerMsg;
      try {
        msg = JSON.parse(e.data) as ServerMsg;
      } catch {
        return;
      }
      switch (msg.type) {
        case "audio":
          enqueueAudio(msg.data);
          break;
        case "transcript":
          setTranscript((t) => [...t, msg.data]);
          break;
        case "turn_end":
          setState("listening");
          break;
        case "error":
          setState("error");
          setErrorMsg(msg.data);
          break;
      }
    };

    ws.onerror = () => {
      setState("error");
      setErrorMsg(L.error);
    };

    ws.onclose = () => {
      stream.getTracks().forEach((t) => t.stop());
      mediaRecorderRef.current?.stop();
      if (audioCtxRef.current?.state !== "closed") {
        audioCtxRef.current?.close();
      }
      setState("idle");
    };
  }, [enqueueAudio]);

  const disconnect = useCallback(() => {
    mediaRecorderRef.current?.stop();
    wsRef.current?.close();
  }, []);

  useEffect(() => () => { wsRef.current?.close(); }, []);

  // ── rendering ─────────────────────────────────────────────────────────────

  const isConnected = state === "listening" || state === "speaking";
  const statusLabel =
    state === "connecting" ? L.connecting
    : state === "listening" ? L.listening
    : state === "speaking"  ? L.speaking
    : state === "error"     ? errorMsg || L.error
    : L.idle;

  return (
    <Card className="flex flex-col gap-0 overflow-hidden">
      <CardHeader
        title={L.title}
        caption={L.caption}
        action={
          <StatusDot state={state} />
        }
      />
      <CardBody className="flex flex-col gap-5">
        {/* Visualiser orb */}
        <div className="flex items-center justify-center py-4">
          <Orb state={state} />
        </div>

        {/* Status label */}
        <p className="text-center text-sm text-fg-muted">{statusLabel}</p>

        {/* Connect / disconnect */}
        <div className="flex justify-center">
          {isConnected ? (
            <Button
              variant="secondary"
              leading={<MicOff className="size-4" />}
              onClick={disconnect}
            >
              {L.disconnect}
            </Button>
          ) : (
            <Button
              leading={<Mic className="size-4" />}
              onClick={connect}
              disabled={state === "connecting"}
            >
              {state === "connecting" ? L.connecting : L.connect}
            </Button>
          )}
        </div>

        {/* Transcript */}
        <TranscriptArea lines={transcript} />
      </CardBody>
    </Card>
  );
}

// ── sub-components ──────────────────────────────────────────────────────────

function Orb({ state }: { state: SessionState }) {
  const pulse = state === "listening" || state === "speaking";
  const color =
    state === "speaking"   ? "bg-brand-500"
    : state === "listening" ? "bg-brand-400"
    : state === "error"     ? "bg-danger"
    : "bg-border";

  return (
    <div className="relative flex size-20 items-center justify-center">
      <AnimatePresence>
        {pulse && (
          <motion.div
            key="pulse"
            className={cn("absolute inset-0 rounded-full opacity-30", color)}
            initial={{ scale: 1 }}
            animate={{ scale: 1.6 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1.2, repeat: Infinity, repeatType: "reverse", ease: "easeInOut" }}
          />
        )}
      </AnimatePresence>
      <div className={cn("relative z-10 flex size-14 items-center justify-center rounded-full shadow-md", color)}>
        <Radio className="size-6 text-fg-inverse" />
      </div>
    </div>
  );
}

function StatusDot({ state }: { state: SessionState }) {
  const color =
    state === "listening" || state === "speaking" ? "bg-success"
    : state === "connecting"                        ? "bg-warning"
    : state === "error"                             ? "bg-danger"
    : "bg-border";
  return <span className={cn("inline-block size-2 rounded-full", color)} />;
}

function TranscriptArea({ lines }: { lines: string[] }) {
  const endRef = useRef<HTMLDivElement>(null);
  useEffect(() => { endRef.current?.scrollIntoView({ behavior: "smooth" }); }, [lines]);

  return (
    <div className="flex flex-col gap-1">
      <p className="text-xs font-medium uppercase tracking-wider text-fg-subtle">
        {COMMON_LABELS.liveSession.transcript}
      </p>
      <div className="h-32 overflow-y-auto rounded-lg bg-bg-muted p-3 text-sm text-fg-muted">
        {lines.length === 0 ? (
          <span className="text-fg-subtle">{COMMON_LABELS.liveSession.noTranscript}</span>
        ) : (
          lines.map((l, i) => (
            <p key={i} className="leading-relaxed">{l}</p>
          ))
        )}
        <div ref={endRef} />
      </div>
    </div>
  );
}
