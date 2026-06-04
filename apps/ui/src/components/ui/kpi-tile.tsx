import { ArrowDownRight, ArrowUpRight, Minus } from "lucide-react";
import type { KpiTile as KpiTileData } from "@/mocks/business/console";
import { cn } from "@/lib/cn";

/**
 * Single KPI tile — large value, label, caption, delta, and a tiny inline
 * sparkline (7 bars). Sparkline is SVG-only — no chart library.
 */
export function KpiTile({ data, className }: { data: KpiTileData; className?: string }) {
  const DirectionIcon =
    data.delta.direction === "up"
      ? ArrowUpRight
      : data.delta.direction === "down"
        ? ArrowDownRight
        : Minus;

  return (
    <div
      className={cn(
        "rounded-2xl border border-border bg-bg-elevated/60 p-5 shadow-sm shadow-fg/[0.02] backdrop-blur-sm",
        className
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-medium tracking-tight text-fg-subtle">{data.label}</p>
          <p className="mt-2 text-3xl font-semibold tracking-tight text-fg">{data.value}</p>
        </div>
        <Sparkline values={data.spark} />
      </div>
      <div className="mt-3 flex items-center gap-1.5 text-[11px] text-fg-muted">
        <DirectionIcon className="size-3" />
        <span className="font-medium text-fg-muted">{data.delta.value}</span>
        <span className="text-fg-subtle">· {data.caption}</span>
      </div>
    </div>
  );
}

function Sparkline({ values }: { values: number[] }) {
  const w = 72;
  const h = 28;
  const gap = 2;
  const barW = (w - gap * (values.length - 1)) / values.length;
  return (
    <svg
      width={w}
      height={h}
      viewBox={`0 0 ${w} ${h}`}
      aria-hidden
      className="shrink-0 text-accent"
    >
      {values.map((v, i) => {
        const bh = Math.max(2, v * h);
        return (
          <rect
            key={i}
            x={i * (barW + gap)}
            y={h - bh}
            width={barW}
            height={bh}
            rx={1}
            fill="currentColor"
            opacity={0.35 + v * 0.5}
          />
        );
      })}
    </svg>
  );
}
