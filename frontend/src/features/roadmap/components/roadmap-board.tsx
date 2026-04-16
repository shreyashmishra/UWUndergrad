import { cn } from "@/lib/utils/cn";
import { CourseCard } from "@/features/roadmap/components/course-card";
import { ElectiveGroupCard } from "@/features/roadmap/components/elective-group-card";
import type { CourseStatus, Roadmap, TermRoadmap } from "@/types/roadmap";

interface RoadmapBoardProps {
  roadmap: Roadmap | null;
  onCourseStatusChange: (courseCode: string, status: CourseStatus) => void;
  onElectiveSelect: (groupCode: string, courseCode: string) => void;
  onElectiveStatusChange: (groupCode: string, courseCode: string, status: CourseStatus) => void;
  onElectiveClear: (groupCode: string) => void;
}

// Assign short labels like "1A", "1B" based on ordering within each year.
// Falls back to the backend label if no year data is present.
function buildTermLabels(terms: TermRoadmap[]): Map<string, string> {
  const labels = new Map<string, string>();
  const sorted = [...terms].sort((a, b) => a.sequence - b.sequence);
  const byYear = new Map<number, TermRoadmap[]>();

  for (const t of sorted) {
    const y = t.year > 0 ? t.year : 0;
    const arr = byYear.get(y) ?? [];
    arr.push(t);
    byYear.set(y, arr);
  }

  for (const [year, yearTerms] of byYear) {
    yearTerms.forEach((t, i) => {
      const letter = String.fromCharCode(65 + i); // A, B, C …
      labels.set(t.code, year > 0 ? `${year}${letter}` : t.label);
    });
  }

  return labels;
}

// Group terms by year for display, preserving sequence order.
function groupByYear(
  terms: TermRoadmap[],
): { year: number | null; items: TermRoadmap[] }[] {
  const sorted = [...terms].sort((a, b) => a.sequence - b.sequence);
  const hasYears = sorted.some((t) => t.year > 0);

  if (!hasYears) {
    return [{ year: null, items: sorted }];
  }

  const map = new Map<number, TermRoadmap[]>();
  for (const t of sorted) {
    const y = t.year > 0 ? t.year : 0;
    const arr = map.get(y) ?? [];
    arr.push(t);
    map.set(y, arr);
  }

  return Array.from(map.entries())
    .sort(([a], [b]) => a - b)
    .map(([year, items]) => ({ year, items }));
}

function termCompletion(term: TermRoadmap) {
  return term.totalCount > 0 ? term.completedCount / term.totalCount : 0;
}

export function RoadmapBoard({
  roadmap,
  onCourseStatusChange,
  onElectiveSelect,
  onElectiveStatusChange,
  onElectiveClear,
}: RoadmapBoardProps) {
  if (!roadmap) {
    return (
      <section className="rounded-[2rem] border border-white/60 bg-white/80 p-12 text-center shadow-panel backdrop-blur">
        <div className="mx-auto max-w-sm">
          <div className="mx-auto mb-5 flex h-16 w-16 items-center justify-center rounded-2xl bg-cloud">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              <path
                d="M6 8h20M6 14h14M6 20h10"
                stroke="#102336"
                strokeOpacity="0.35"
                strokeWidth="2"
                strokeLinecap="round"
              />
            </svg>
          </div>
          <h2 className="font-display text-2xl text-ink">
            Your academic pathway appears here
          </h2>
          <p className="mt-3 text-sm leading-6 text-ink/55">
            Select a program on the left to load your personalized roadmap from
            1A to 4B.
          </p>
        </div>
      </section>
    );
  }

  const sorted = [...roadmap.terms].sort((a, b) => a.sequence - b.sequence);
  const termLabels = buildTermLabels(sorted);
  const groups = groupByYear(sorted);

  return (
    <section className="space-y-5">
      {/* Pathway strip */}
      <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
        <p className="mb-4 text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
          Academic Pathway
        </p>
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
          {sorted.map((term, i) => {
            const label = termLabels.get(term.code) ?? term.label;
            const pct = termCompletion(term);
            const isDone = pct >= 1;
            const isActive = pct > 0 && pct < 1;

            return (
              <div key={term.code} className="flex flex-shrink-0 items-center gap-1.5">
                <div
                  className={cn(
                    "flex min-w-[52px] flex-col items-center gap-1 rounded-xl px-3 py-2.5 text-center",
                    isDone
                      ? "bg-teal text-white"
                      : isActive
                        ? "bg-sand/80 text-amber"
                        : "border border-ink/8 bg-cloud text-ink/50",
                  )}
                >
                  <span className="text-xs font-bold tracking-wide">{label}</span>
                  <span className="text-[10px] opacity-70">
                    {term.completedCount}/{term.totalCount}
                  </span>
                </div>
                {i < sorted.length - 1 && (
                  <svg
                    className="flex-shrink-0 text-ink/20"
                    width="14"
                    height="14"
                    viewBox="0 0 14 14"
                    fill="none"
                  >
                    <path
                      d="M3 7h8M8 4l3 3-3 3"
                      stroke="currentColor"
                      strokeWidth="1.4"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Terms grouped by year */}
      {groups.map(({ year, items }) => (
        <div key={year ?? "all"} className="space-y-3">
          {year !== null && (
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-ink px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.2em] text-white">
                Year {year}
              </div>
              <div className="h-px flex-1 bg-ink/10" />
            </div>
          )}

          <div className="grid gap-4 xl:grid-cols-2">
            {items.map((term) => {
              const shortLabel = termLabels.get(term.code) ?? term.label;
              const pct = termCompletion(term);

              return (
                <article
                  key={term.code}
                  className="flex flex-col rounded-[1.7rem] border border-ink/8 bg-cloud p-5"
                >
                  {/* Term header */}
                  <div className="mb-4 flex items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-ink text-sm font-bold text-white">
                        {shortLabel}
                      </div>
                      <div>
                        <h3 className="font-display text-xl leading-tight text-ink">
                          {term.label}
                        </h3>
                        {term.year > 0 && (
                          <p className="mt-0.5 text-xs text-ink/50">
                            {term.season === "FALL"
                              ? "Fall"
                              : term.season === "WINTER"
                                ? "Winter"
                                : "Spring"}{" "}
                            · Year {term.year}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="rounded-2xl bg-white px-3 py-2 text-right shadow-sm">
                      <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-ink/40">
                        Done
                      </p>
                      <p className="mt-0.5 font-display text-xl text-ink">
                        {term.completedCount}/{term.totalCount}
                      </p>
                    </div>
                  </div>

                  {/* Term progress bar */}
                  <div className="mb-4 h-1.5 overflow-hidden rounded-full bg-ink/8">
                    <div
                      className="h-full rounded-full bg-gradient-to-r from-teal to-amber transition-all duration-500"
                      style={{ width: `${pct * 100}%` }}
                    />
                  </div>

                  {/* Requirements */}
                  <div className="space-y-3">
                    {term.requirements.map((req) => {
                      if (req.course) {
                        return (
                          <CourseCard
                            key={`${term.code}-${req.course.code}`}
                            course={req.course}
                            variant="required"
                            onStatusChange={(status) =>
                              onCourseStatusChange(req.course!.code, status)
                            }
                          />
                        );
                      }

                      if (req.group) {
                        return (
                          <ElectiveGroupCard
                            key={`${term.code}-${req.group.code}`}
                            group={req.group}
                            onSelectOption={(courseCode) =>
                              onElectiveSelect(req.group!.code, courseCode)
                            }
                            onOptionStatusChange={(courseCode, status) =>
                              onElectiveStatusChange(
                                req.group!.code,
                                courseCode,
                                status,
                              )
                            }
                            onClearSelection={() =>
                              onElectiveClear(req.group!.code)
                            }
                          />
                        );
                      }

                      return null;
                    })}
                  </div>
                </article>
              );
            })}
          </div>
        </div>
      ))}
    </section>
  );
}
