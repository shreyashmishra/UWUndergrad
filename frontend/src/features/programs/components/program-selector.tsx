import type { Program } from "@/types/roadmap";

interface ProgramSelectorProps {
  programs: Program[];
  selectedCode: string | null;
  onChange: (value: string) => void;
}

export function ProgramSelector({
  programs,
  selectedCode,
  onChange,
}: ProgramSelectorProps) {
  return (
    <label className="block space-y-2">
      <span className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/60">
        Program
      </span>
      <select
        value={selectedCode ?? ""}
        onChange={(event) => onChange(event.target.value)}
        className="w-full rounded-2xl border border-white/70 bg-white/90 px-4 py-3 text-sm font-medium text-ink shadow-sm outline-none transition focus:border-teal focus:ring-2 focus:ring-teal/20"
      >
        <option value="" disabled>
          Select a program
        </option>
        {programs.map((program) => (
          <option key={program.code} value={program.code}>
            {program.name}
          </option>
        ))}
      </select>
    </label>
  );
}
