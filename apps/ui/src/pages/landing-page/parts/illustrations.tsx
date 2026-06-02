/**
 * Custom SVG illustrations for the landing page. Hand-built, no external
 * assets, themed via currentColor. Use these inside <ParallaxIllustration>
 * or anywhere ambient visual depth is needed.
 *
 * Each illustration is `aria-hidden` — they're decorative, not informational.
 */
import { cn } from "@/lib/cn";

type Props = {
  className?: string;
};

/* A constellation of dots and connecting lines — evokes a knowledge graph.
   The organic shape suggests relationships between ideas. */
export function KnowledgeGraph({ className }: Props) {
  const nodes = [
    { x: 60, y: 30 },
    { x: 180, y: 70 },
    { x: 110, y: 140 },
    { x: 250, y: 50 },
    { x: 220, y: 170 },
    { x: 320, y: 110 },
    { x: 360, y: 200 },
    { x: 90, y: 220 },
  ];
  const edges: [number, number][] = [
    [0, 1],
    [0, 2],
    [1, 3],
    [1, 4],
    [2, 4],
    [3, 5],
    [4, 5],
    [4, 6],
    [5, 6],
    [2, 7],
    [6, 7],
  ];
  return (
    <svg
      viewBox="0 0 400 260"
      className={cn("h-full w-auto opacity-60", className)}
      aria-hidden
    >
      <g stroke="currentColor" strokeWidth="0.75" fill="none" opacity="0.5">
        {edges.map(([a, b], i) => {
          const A = nodes[a]!;
          const B = nodes[b]!;
          return <line key={i} x1={A.x} y1={A.y} x2={B.x} y2={B.y} />;
        })}
      </g>
      <g fill="currentColor">
        {nodes.map((n, i) => (
          <circle key={i} cx={n.x} cy={n.y} r="3" />
        ))}
      </g>
    </svg>
  );
}

/* Concentric rings with a small offset center — suggests the live-monitor
   health score visualization. Calm, geometric, slightly hypnotic. */
export function ConcentricRings({ className }: Props) {
  return (
    <svg viewBox="0 0 200 200" className={cn("h-full w-auto opacity-50", className)} aria-hidden>
      <g stroke="currentColor" fill="none" strokeWidth="0.5">
        <circle cx="100" cy="100" r="20" />
        <circle cx="100" cy="100" r="40" />
        <circle cx="100" cy="100" r="60" />
        <circle cx="100" cy="100" r="80" />
        <circle cx="100" cy="100" r="95" />
      </g>
      <circle cx="106" cy="94" r="3" fill="currentColor" />
    </svg>
  );
}

/* A horizontal flow of dots, varying heights — abstract "voice activity"
   but calmer than the hero waveform. Good for ambient backgrounds. */
export function FlowDots({ className }: Props) {
  const dots = Array.from({ length: 48 }, (_, i) => {
    const t = (i / 48) * Math.PI * 6;
    const h = 30 + 30 * Math.abs(Math.sin(t * 0.7) * Math.cos(t * 0.4));
    return { x: i * 18, y: 80 - h / 2, h };
  });
  return (
    <svg viewBox="0 0 880 80" className={cn("h-full w-auto opacity-50", className)} aria-hidden>
      <g fill="currentColor">
        {dots.map((d, i) => (
          <rect key={i} x={d.x} y={d.y} width="3" height={d.h} rx="1.5" />
        ))}
      </g>
    </svg>
  );
}

/* A subtle dotted grid, fading at the edges. The kind of "we are
   technical and precise" background without being loud about it. */
export function DotGrid({ className }: Props) {
  return (
    <svg viewBox="0 0 400 200" className={cn("h-full w-auto opacity-40", className)} aria-hidden>
      <defs>
        <radialGradient id="dotgrid-fade" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="currentColor" stopOpacity="1" />
          <stop offset="100%" stopColor="currentColor" stopOpacity="0" />
        </radialGradient>
        <mask id="dotgrid-mask">
          <rect width="400" height="200" fill="url(#dotgrid-fade)" />
        </mask>
      </defs>
      <g fill="currentColor" mask="url(#dotgrid-mask)">
        {Array.from({ length: 20 }).map((_, r) =>
          Array.from({ length: 40 }).map((_, c) => (
            <circle key={`${r}-${c}`} cx={c * 10 + 5} cy={r * 10 + 5} r="0.9" />
          ))
        )}
      </g>
    </svg>
  );
}
