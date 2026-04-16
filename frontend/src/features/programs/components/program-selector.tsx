"use client";

import { useEffect, useRef, useState } from "react";

import { cn } from "@/lib/utils/cn";
import type { Program } from "@/types/roadmap";

interface ProgramSelectorProps {
  programs: Program[];
  selectedCode: string | null;
  onChange: (value: string) => void;
  disabled?: boolean;
  loading?: boolean;
}

export function ProgramSelector({
  programs,
  selectedCode,
  onChange,
  disabled = false,
  loading = false,
}: ProgramSelectorProps) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const selected = programs.find((p) => p.code === selectedCode) ?? null;

  const filtered = query.trim()
    ? programs.filter((p) =>
        `${p.code} ${p.name} ${p.degree}`
          .toLowerCase()
          .includes(query.toLowerCase()),
      )
    : programs;

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
        setQuery("");
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSelect = (code: string) => {
    onChange(code);
    setOpen(false);
    setQuery("");
  };

  const displayValue = open
    ? query
    : selected
      ? `${selected.code} – ${selected.name}`
      : "";

  return (
    <div ref={containerRef} className="relative">
      <label className="block">
        <span className="mb-2 block text-xs font-semibold uppercase tracking-[0.24em] text-ink/55">
          Program
        </span>
        <div className="relative">
          <input
            type="text"
            value={displayValue}
            placeholder={
              loading
                ? "Loading programs…"
                : "Search, e.g. BMATH – Data Science"
            }
            disabled={disabled}
            onChange={(e) => {
              setQuery(e.target.value);
              setOpen(true);
            }}
            onFocus={() => setOpen(true)}
            className="w-full rounded-2xl border border-white/70 bg-white/90 px-4 py-3 pr-10 text-sm font-medium text-ink shadow-sm outline-none transition placeholder:font-normal placeholder:text-ink/40 focus:border-teal focus:ring-2 focus:ring-teal/20 disabled:cursor-not-allowed disabled:opacity-50"
          />
          <svg
            className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-ink/35"
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
          >
            <path
              d="M4 6l4 4 4-4"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
      </label>

      {open && filtered.length > 0 && (
        <ul className="absolute left-0 right-0 top-full z-50 mt-2 max-h-64 overflow-y-auto rounded-2xl border border-white/60 bg-white shadow-xl">
          {filtered.map((p, i) => (
            <li key={p.code}>
              <button
                type="button"
                onClick={() => handleSelect(p.code)}
                className={cn(
                  "flex w-full flex-col gap-0.5 px-4 py-3 text-left transition hover:bg-cloud",
                  i === 0 && "rounded-t-2xl",
                  i === filtered.length - 1 && "rounded-b-2xl",
                  selected?.code === p.code && "bg-mint/40",
                )}
              >
                <span className="text-sm font-semibold text-ink">
                  {p.code} – {p.name}
                </span>
                <span className="text-xs text-ink/50">{p.degree}</span>
              </button>
            </li>
          ))}
        </ul>
      )}

      {open && !loading && filtered.length === 0 && (
        <div className="absolute left-0 right-0 top-full z-50 mt-2 rounded-2xl border border-white/60 bg-white px-4 py-4 text-sm text-ink/50 shadow-xl">
          No programs match &ldquo;{query}&rdquo;
        </div>
      )}
    </div>
  );
}
