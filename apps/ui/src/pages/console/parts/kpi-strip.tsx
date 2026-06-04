import { motion } from "framer-motion";
import { KpiTile } from "@/components/ui/kpi-tile";
import { RevealStagger, revealItem } from "@/components/ui/reveal";
import { getKpiSnapshot } from "@/mocks/business/console";

const ease = [0.22, 1, 0.36, 1] as const;

export function KpiStrip() {
  const kpis = getKpiSnapshot();
  return (
    <RevealStagger className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {kpis.map((kpi) => (
        <motion.div
          key={kpi.id}
          variants={revealItem}
          transition={{ duration: 0.6, ease }}
        >
          <KpiTile data={kpi} />
        </motion.div>
      ))}
    </RevealStagger>
  );
}
