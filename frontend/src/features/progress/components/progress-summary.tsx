import type { RequirementSummary } from "@/types/roadmap";

interface ProgressSummaryProps {
  summary: RequirementSummary | null;
}

const summaryCards = [
  { key: "completedRequirements", label: "Completed", accent: "bg-teal/12 text-teal" },
  { key: "inProgressRequirements", label: "In Progress", accent: "bg-sand text-amber" },
  { key: "plannedRequirements", label: "Planned", accent: "bg-mint text-teal" },
  { key: "remainingRequirements", label: "Remaining", accent: "bg-rose/10 text-rose" },
] as const;

export function ProgressSummary({ summary }: ProgressSummaryProps) {
  return (
    <section className="rounded-[2rem] border border-white/60 bg-white/80 p-5 shadow-panel backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/55">
            Requirement Summary
          </p>
          <h2 className="mt-2 font-display text-3xl text-ink">
            {summary ? `${Math.round(summary.completionPercentage)}% complete` : "Waiting for roadmap"}
          </h2>
        </div>
        <div className="rounded-full bg-ink px-3 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-white">
          {summary ? `${summary.selectedElectives} elective picks` : "0 elective picks"}
        </div>
      </div>

      <div className="mt-4 h-3 overflow-hidden rounded-full bg-ink/10">
        <div
          className="h-full rounded-full bg-gradient-to-r from-teal to-amber transition-all duration-500"
          style={{ width: `${summary?.completionPercentage ?? 0}%` }}
        />
      </div>

      <div className="mt-5 grid grid-cols-2 gap-3">
        {summaryCards.map((item) => (
          <div key={item.key} className="rounded-2xl border border-ink/8 bg-cloud px-4 py-3">
            <div className={`inline-flex rounded-full px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] ${item.accent}`}>
              {item.label}
            </div>
            <p className="mt-3 font-display text-3xl text-ink">
              {summary ? summary[item.key] : 0}
            </p>
          </div>
        ))}
      </div>

      <div className="mt-4 rounded-2xl bg-ink px-4 py-4 text-white">
        <p className="text-xs uppercase tracking-[0.2em] text-white/65">Total requirement slots</p>
        <p className="mt-2 text-3xl font-display">
          {summary?.totalRequirements ?? 0}
        </p>
      </div>
    </section>
  );
}
