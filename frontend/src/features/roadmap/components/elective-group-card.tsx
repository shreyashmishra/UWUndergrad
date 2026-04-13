import { CourseCard } from "@/features/roadmap/components/course-card";
import type { CourseStatus, ElectiveGroup } from "@/types/roadmap";

interface ElectiveGroupCardProps {
  group: ElectiveGroup;
  onSelectOption: (courseCode: string) => void;
  onOptionStatusChange: (courseCode: string, status: CourseStatus) => void;
  onClearSelection: () => void;
}

export function ElectiveGroupCard({
  group,
  onSelectOption,
  onOptionStatusChange,
  onClearSelection,
}: ElectiveGroupCardProps) {
  return (
    <section className="rounded-[1.9rem] border border-dashed border-teal/20 bg-white/75 p-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <span className="rounded-full bg-teal/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-teal">
            Choose {group.minSelections}
          </span>
          <h4 className="mt-3 font-display text-2xl text-ink">{group.title}</h4>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-ink/70">
            {group.description}
          </p>
        </div>

        <button
          type="button"
          onClick={onClearSelection}
          className="rounded-full border border-ink/10 bg-white px-3 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-ink/60 transition hover:border-rose/20 hover:text-rose"
        >
          Clear choice
        </button>
      </div>

      <div className="mt-5 grid gap-4 xl:grid-cols-2">
        {group.options.map((option) => (
          <CourseCard
            key={option.code}
            course={option}
            variant="elective"
            onSelect={() => onSelectOption(option.code)}
            onStatusChange={(status) => onOptionStatusChange(option.code, status)}
          />
        ))}
      </div>
    </section>
  );
}
