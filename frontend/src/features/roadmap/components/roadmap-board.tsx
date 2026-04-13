import { CourseCard } from "@/features/roadmap/components/course-card";
import { ElectiveGroupCard } from "@/features/roadmap/components/elective-group-card";
import type { CourseStatus, Roadmap } from "@/types/roadmap";

interface RoadmapBoardProps {
  roadmap: Roadmap | null;
  onCourseStatusChange: (courseCode: string, status: CourseStatus) => void;
  onElectiveSelect: (groupCode: string, courseCode: string) => void;
  onElectiveStatusChange: (groupCode: string, courseCode: string, status: CourseStatus) => void;
  onElectiveClear: (groupCode: string) => void;
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
      <section className="rounded-[2rem] border border-white/60 bg-white/80 p-8 text-center shadow-panel backdrop-blur">
        <p className="text-sm font-semibold uppercase tracking-[0.24em] text-ink/50">
          Roadmap
        </p>
        <h2 className="mt-3 font-display text-3xl text-ink">
          Select a university and program to load the roadmap
        </h2>
      </section>
    );
  }

  const years = roadmap.terms.reduce<Record<number, typeof roadmap.terms>>((accumulator, term) => {
    accumulator[term.year] = [...(accumulator[term.year] ?? []), term];
    return accumulator;
  }, {});

  return (
    <div className="space-y-8">
      {Object.entries(years).map(([year, terms]) => (
        <section
          key={year}
          className="rounded-[2rem] border border-white/60 bg-white/80 p-6 shadow-panel backdrop-blur"
        >
          <div className="flex items-end justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
                Semester Roadmap
              </p>
              <h2 className="mt-2 font-display text-3xl text-ink">Year {year}</h2>
            </div>
            <p className="max-w-xl text-sm leading-6 text-ink/65">
              Track courses term by term. Required courses are fixed, while elective groups let the student choose a direction without losing the overall graduation picture.
            </p>
          </div>

          <div className="mt-6 grid gap-5 xl:grid-cols-3">
            {terms.map((term) => (
              <article
                key={term.code}
                className="flex h-full flex-col rounded-[1.7rem] border border-ink/8 bg-cloud p-5"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <span className="rounded-full bg-ink px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-white">
                      {term.season}
                    </span>
                    <h3 className="mt-3 font-display text-2xl text-ink">{term.label}</h3>
                  </div>
                  <div className="rounded-2xl bg-white px-3 py-2 text-right shadow-sm">
                    <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-ink/45">
                      Complete
                    </p>
                    <p className="mt-1 font-display text-2xl text-ink">
                      {term.completedCount}/{term.totalCount}
                    </p>
                  </div>
                </div>

                <div className="mt-5 space-y-4">
                  {term.requirements.map((requirement) => {
                    if (requirement.course) {
                      return (
                        <CourseCard
                          key={`${term.code}-${requirement.course.code}`}
                          course={requirement.course}
                          variant="required"
                          onStatusChange={(status) =>
                            onCourseStatusChange(requirement.course!.code, status)
                          }
                        />
                      );
                    }

                    if (requirement.group) {
                      return (
                        <ElectiveGroupCard
                          key={`${term.code}-${requirement.group.code}`}
                          group={requirement.group}
                          onSelectOption={(courseCode) =>
                            onElectiveSelect(requirement.group!.code, courseCode)
                          }
                          onOptionStatusChange={(courseCode, status) =>
                            onElectiveStatusChange(requirement.group!.code, courseCode, status)
                          }
                          onClearSelection={() => onElectiveClear(requirement.group!.code)}
                        />
                      );
                    }

                    return null;
                  })}
                </div>
              </article>
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
