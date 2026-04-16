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

// Return the short display label for a term.
// The backend now sets label = "1A", "2B", etc. for proper academic terms.
// For requirement sections (year=0) we use the full label.
function shortLabel(term: TermRoadmap): string {
  return term.label;
}

// Build year-grouped buckets.
// year=0  → requirement sections without a real term structure (e.g. CS "Required Courses")
// year>0  → real academic terms (1A, 1B, 2A … 4B)
function groupByYear(
  terms: TermRoadmap[],
): { year: number; label: string; items: TermRoadmap[] }[] {
  const sorted = [...terms].sort((a, b) => a.sequence - b.sequence);
  const map = new Map<number, TermRoadmap[]>();

  for (const t of sorted) {
    const arr = map.get(t.year) ?? [];
    arr.push(t);
    map.set(t.year, arr);
  }

  return Array.from(map.entries())
    .sort(([a], [b]) => a - b)
    .map(([year, items]) => ({
      year,
      label: year === 0 ? "Requirements" : `Year ${year}`,
      items,
    }));
}

function termCompletion(term: TermRoadmap): number {
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
  const groups = groupByYear(sorted);

  // Only real academic terms (year > 0) appear in the pathway strip.
  const pathwayTerms = sorted.filter((t) => t.year > 0);
  const hasPathway = pathwayTerms.length > 0;

  return (
    <section className="space-y-5">
      {/* Pathway strip — only shown when the program has real term data */}
      {hasPathway && (
        <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
          <p className="mb-4 text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
            Academic Pathway
          </p>
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
            {pathwayTerms.map((term, i) => {
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
                    <span className="text-xs font-bold tracking-wide">
                      {shortLabel(term)}
                    </span>
                    <span className="text-[10px] opacity-70">
                      {term.completedCount}/{term.totalCount}
                    </span>
                  </div>
                  {i < pathwayTerms.length - 1 && (
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
      )}

      {/* Terms grouped by year */}
      {groups.map(({ year, label: groupLabel, items }) => {
        // Skip groups with no requirements
        if (items.every((t) => t.requirements.length === 0)) {
          return null;
        }

        return (
          <div key={year} className="space-y-3">
            <div className="flex items-center gap-3">
              <div
                className={cn(
                  "rounded-full px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.2em]",
                  year === 0
                    ? "bg-ink/8 text-ink/60"
                    : "bg-ink text-white",
                )}
              >
                {groupLabel}
              </div>
              <div className="h-px flex-1 bg-ink/10" />
            </div>

            <div
              className={cn(
                "grid gap-4",
                // Academic years: 2 terms side-by-side; plan notes: single column
                year > 0 ? "xl:grid-cols-2" : "xl:grid-cols-1",
              )}
            >
              {items.map((term) => {
                const pct = termCompletion(term);
                const isAcademic = term.year > 0;

                return (
                  <article
                    key={term.code}
                    className="flex flex-col rounded-[1.7rem] border border-ink/8 bg-cloud p-5"
                  >
                    {/* Term header */}
                    <div className="mb-4 flex items-center justify-between gap-3">
                      <div className="flex items-center gap-3">
                        {isAcademic ? (
                          <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-ink text-sm font-bold text-white">
                            {shortLabel(term)}
                          </div>
                        ) : (
                          <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-ink/8">
                            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
                              <path
                                d="M3 5h12M3 9h8M3 13h10"
                                stroke="#102336"
                                strokeOpacity="0.4"
                                strokeWidth="1.5"
                                strokeLinecap="round"
                              />
                            </svg>
                          </div>
                        )}
                        <div>
                          <h3 className="font-display text-xl leading-tight text-ink">
                            {isAcademic
                              ? term.season === "FALL"
                                ? `Fall · Year ${term.year}`
                                : term.season === "WINTER"
                                  ? `Winter · Year ${term.year}`
                                  : `Spring · Year ${term.year}`
                              : term.label}
                          </h3>
                        </div>
                      </div>

                      {term.totalCount > 0 && (
                        <div className="rounded-2xl bg-white px-3 py-2 text-right shadow-sm">
                          <p className="text-[10px] font-semibold uppercase tracking-[0.16em] text-ink/40">
                            Done
                          </p>
                          <p className="mt-0.5 font-display text-xl text-ink">
                            {term.completedCount}/{term.totalCount}
                          </p>
                        </div>
                      )}
                    </div>

                    {/* Progress bar */}
                    {term.totalCount > 0 && (
                      <div className="mb-4 h-1.5 overflow-hidden rounded-full bg-ink/8">
                        <div
                          className="h-full rounded-full bg-gradient-to-r from-teal to-amber transition-all duration-500"
                          style={{ width: `${pct * 100}%` }}
                        />
                      </div>
                    )}

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
        );
      })}
    </section>
  );
}
