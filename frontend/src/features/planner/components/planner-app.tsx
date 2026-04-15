"use client";

import { startTransition, useDeferredValue, useEffect, useMemo, useState } from "react";

import {
  addCourseToTerm,
  createTerm,
  fetchDependencyView,
  fetchPlanner,
  moveRoadmapCourse,
  removeRoadmapCourse,
  searchCourses,
  updateRoadmapCourseStatus,
} from "@/features/planner/lib/api";
import type {
  Course,
  CourseStatus,
  DependencyView,
  Planner,
  PlannerTerm,
} from "@/features/planner/lib/types";

const STATUS_OPTIONS: CourseStatus[] = ["COMPLETED", "IN_PROGRESS", "PLANNED"];
const SEASON_OPTIONS = ["FALL", "WINTER", "SPRING"] as const;

function formatCredits(value: number) {
  return `${value.toFixed(1)} credits`;
}

function statusTone(status: CourseStatus) {
  switch (status) {
    case "COMPLETED":
      return "bg-teal/15 text-teal";
    case "IN_PROGRESS":
      return "bg-sand text-amber";
    default:
      return "bg-ink/10 text-ink";
  }
}

export function PlannerApp() {
  const [planner, setPlanner] = useState<Planner | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Course[]>([]);
  const [activeTermID, setActiveTermID] = useState<string>("");
  const [selectedCourseCode, setSelectedCourseCode] = useState<string | null>(null);
  const [dependencyView, setDependencyView] = useState<DependencyView | null>(null);
  const [newTermSeason, setNewTermSeason] = useState<(typeof SEASON_OPTIONS)[number]>("FALL");
  const [newTermYear, setNewTermYear] = useState<number>(2027);
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [isMutating, setIsMutating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const deferredSearchQuery = useDeferredValue(searchQuery);

  useEffect(() => {
    const loadPlanner = async () => {
      try {
        const nextPlanner = await fetchPlanner();
        setPlanner(nextPlanner);
        setActiveTermID(nextPlanner.terms[0]?.id ?? "");
        setError(null);
      } catch (loadError) {
        setError(
          loadError instanceof Error
            ? loadError.message
            : "Unable to load planner data from the Go API.",
        );
      } finally {
        setIsBootstrapping(false);
      }
    };

    void loadPlanner();
  }, []);

  useEffect(() => {
    const loadSearchResults = async () => {
      if (!deferredSearchQuery.trim()) {
        setSearchResults([]);
        return;
      }

      try {
        const results = await searchCourses(deferredSearchQuery);
        setSearchResults(results);
        setError(null);
      } catch (searchError) {
        setError(
          searchError instanceof Error
            ? searchError.message
            : "Unable to search the course catalog.",
        );
      }
    };

    void loadSearchResults();
  }, [deferredSearchQuery]);

  useEffect(() => {
    const loadDependencyView = async () => {
      if (!selectedCourseCode) {
        setDependencyView(null);
        return;
      }

      try {
        const nextDependencyView = await fetchDependencyView(selectedCourseCode);
        setDependencyView(nextDependencyView);
      } catch {
        setDependencyView(null);
      }
    };

    void loadDependencyView();
  }, [selectedCourseCode]);

  const totalWarnings = useMemo(
    () =>
      planner?.terms.reduce(
        (count, term) =>
          count +
          term.entries.filter(
            (entry) => entry.prerequisiteIssue && !entry.prerequisiteIssue.isSatisfied,
          ).length,
        0,
      ) ?? 0,
    [planner],
  );

  const runMutation = async (operation: () => Promise<Planner>) => {
    try {
      setIsMutating(true);
      const nextPlanner = await operation();
      startTransition(() => {
        setPlanner(nextPlanner);
        if (!activeTermID) {
          setActiveTermID(nextPlanner.terms[0]?.id ?? "");
        }
      });
      setError(null);
    } catch (mutationError) {
      setError(
        mutationError instanceof Error
          ? mutationError.message
          : "Planner update failed.",
      );
    } finally {
      setIsMutating(false);
    }
  };

  const handleCreateTerm = async () => {
    await runMutation(() => createTerm(newTermSeason, newTermYear));
  };

  const handleAddCourse = async (courseCode: string) => {
    if (!activeTermID) {
      setError("Choose a target term before adding a course.");
      return;
    }

    await runMutation(() => addCourseToTerm(activeTermID, courseCode));
  };

  const handleRemoveCourse = async (entryID: string) => {
    await runMutation(() => removeRoadmapCourse(entryID));
  };

  const handleMoveCourse = async (entryID: string, targetTermID: string) => {
    if (!targetTermID) {
      return;
    }
    await runMutation(() => moveRoadmapCourse(entryID, targetTermID));
  };

  const handleStatusChange = async (entryID: string, status: CourseStatus) => {
    await runMutation(() => updateRoadmapCourseStatus(entryID, status));
  };

  if (isBootstrapping) {
    return (
      <main className="min-h-screen bg-hero-gradient px-4 py-8 text-ink sm:px-6">
        <div className="mx-auto max-w-7xl rounded-[2rem] border border-white/60 bg-white/70 p-10 shadow-panel backdrop-blur">
          <p className="text-sm uppercase tracking-[0.25em] text-teal">PlanAhead</p>
          <h1 className="mt-3 font-display text-4xl font-semibold">Loading planner…</h1>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-hero-gradient px-4 py-6 text-ink sm:px-6 lg:px-8">
      <div className="mx-auto grid max-w-7xl gap-6 xl:grid-cols-[360px,minmax(0,1fr)]">
        <aside className="space-y-5">
          <section className="rounded-[2rem] border border-white/70 bg-white/80 p-6 shadow-panel backdrop-blur">
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-teal">
              University of Waterloo
            </p>
            <h1 className="mt-3 font-display text-4xl font-semibold leading-tight">
              Degree roadmap planner
            </h1>
            <p className="mt-3 text-sm leading-6 text-ink/70">
              Plan term by term, spot prerequisite sequencing problems early, and keep a clear view of
              your completed, active, and future coursework.
            </p>
            <div className="mt-5 grid gap-3">
              <StatCard label="Completed" value={formatCredits(planner?.progressSummary.completedCredits ?? 0)} />
              <StatCard label="In progress" value={formatCredits(planner?.progressSummary.inProgressCredits ?? 0)} />
              <StatCard label="Planned" value={formatCredits(planner?.progressSummary.plannedCredits ?? 0)} />
              <StatCard label="Warnings" value={`${totalWarnings}`} accent={totalWarnings > 0 ? "rose" : "teal"} />
            </div>
          </section>

          <section className="rounded-[2rem] border border-white/70 bg-white/80 p-6 shadow-panel backdrop-blur">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-teal">
                  Catalog search
                </p>
                <h2 className="mt-2 font-display text-2xl font-semibold">Add future courses</h2>
              </div>
              <span className="rounded-full bg-mint px-3 py-1 text-xs font-medium text-teal">
                {planner?.catalogCourseCount ?? 0} catalog courses
              </span>
            </div>

            <label className="mt-5 block text-sm font-medium text-ink/70">
              Target term
              <select
                value={activeTermID}
                onChange={(event) => setActiveTermID(event.target.value)}
                className="mt-2 w-full rounded-2xl border border-ink/10 bg-cloud px-4 py-3 outline-none transition focus:border-teal"
              >
                {(planner?.terms ?? []).map((term) => (
                  <option key={term.id} value={term.id}>
                    {term.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="mt-4 block text-sm font-medium text-ink/70">
              Course search
              <input
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
                placeholder="Search CS 246 or Object-Oriented"
                className="mt-2 w-full rounded-2xl border border-ink/10 bg-cloud px-4 py-3 outline-none transition focus:border-teal"
              />
            </label>

            <div className="mt-4 space-y-3">
              {searchResults.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-ink/15 bg-cloud px-4 py-5 text-sm text-ink/55">
                  Search by code or title to pull courses into the roadmap.
                </div>
              ) : (
                searchResults.map((course) => (
                  <div
                    key={course.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => setSelectedCourseCode(course.code)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        setSelectedCourseCode(course.code);
                      }
                    }}
                    className="block w-full cursor-pointer rounded-2xl border border-ink/10 bg-cloud p-4 text-left transition hover:border-teal/40 hover:bg-white"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="font-display text-lg font-semibold">{course.code}</p>
                        <p className="text-sm text-ink/75">{course.title}</p>
                      </div>
                      <button
                        type="button"
                        onClick={(event) => {
                          event.stopPropagation();
                          void handleAddCourse(course.code);
                        }}
                        className="rounded-full bg-teal px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.16em] text-white"
                      >
                        Add
                      </button>
                    </div>
                    <p className="mt-3 text-sm leading-6 text-ink/60">{course.description}</p>
                  </div>
                ))
              )}
            </div>
          </section>

          <section className="rounded-[2rem] border border-white/70 bg-white/80 p-6 shadow-panel backdrop-blur">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-teal">Create term</p>
            <div className="mt-4 grid grid-cols-2 gap-3">
              <select
                value={newTermSeason}
                onChange={(event) => setNewTermSeason(event.target.value as (typeof SEASON_OPTIONS)[number])}
                className="rounded-2xl border border-ink/10 bg-cloud px-4 py-3 outline-none transition focus:border-teal"
              >
                {SEASON_OPTIONS.map((season) => (
                  <option key={season} value={season}>
                    {season}
                  </option>
                ))}
              </select>
              <input
                type="number"
                min={2026}
                max={2035}
                value={newTermYear}
                onChange={(event) => setNewTermYear(Number(event.target.value))}
                className="rounded-2xl border border-ink/10 bg-cloud px-4 py-3 outline-none transition focus:border-teal"
              />
            </div>
            <button
              type="button"
              onClick={() => void handleCreateTerm()}
              className="mt-4 w-full rounded-2xl bg-ink px-4 py-3 font-semibold text-white"
            >
              Add term to roadmap
            </button>
          </section>

          <section className="rounded-[2rem] border border-white/70 bg-white/80 p-6 shadow-panel backdrop-blur">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-teal">Dependencies</p>
            <h2 className="mt-2 font-display text-2xl font-semibold">
              {dependencyView?.courseCode ?? "Select a course"}
            </h2>
            <p className="mt-3 text-sm leading-6 text-ink/70">
              {dependencyView?.message ??
                "Click a roadmap course or a search result to inspect its direct prerequisite chain."}
            </p>
            {dependencyView?.directPrerequisites.length ? (
              <div className="mt-4 flex flex-wrap gap-2">
                {dependencyView.directPrerequisites.map((courseCode) => (
                  <span
                    key={courseCode}
                    className="rounded-full bg-mint px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-teal"
                  >
                    {courseCode}
                  </span>
                ))}
              </div>
            ) : null}
          </section>

          {error ? (
            <section className="rounded-[2rem] border border-rose/25 bg-white/80 p-4 text-sm text-rose shadow-panel">
              {error}
            </section>
          ) : null}
        </aside>

        <section className="rounded-[2rem] border border-white/70 bg-white/70 p-6 shadow-panel backdrop-blur">
          <div className="flex items-end justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-teal">Roadmap</p>
              <h2 className="mt-2 font-display text-3xl font-semibold">Semester-by-semester plan</h2>
            </div>
            <div className="rounded-full bg-cloud px-4 py-2 text-sm text-ink/70">
              {planner?.user.displayName} · {planner?.terms.length ?? 0} terms
            </div>
          </div>

          <div className="mt-6 flex gap-5 overflow-x-auto pb-2">
            {(planner?.terms ?? []).map((term) => (
              <TermColumn
                key={term.id}
                term={term}
                allTerms={planner?.terms ?? []}
                selectedCourseCode={selectedCourseCode}
                isMutating={isMutating}
                onSelectCourse={setSelectedCourseCode}
                onRemoveCourse={handleRemoveCourse}
                onMoveCourse={handleMoveCourse}
                onStatusChange={handleStatusChange}
              />
            ))}
          </div>
        </section>
      </div>
    </main>
  );
}

function StatCard({
  label,
  value,
  accent = "teal",
}: {
  label: string;
  value: string;
  accent?: "teal" | "rose";
}) {
  return (
    <div className="rounded-2xl border border-ink/10 bg-cloud px-4 py-3">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-ink/55">{label}</p>
      <p className={`mt-2 font-display text-2xl font-semibold ${accent === "rose" ? "text-rose" : "text-ink"}`}>
        {value}
      </p>
    </div>
  );
}

function TermColumn({
  term,
  allTerms,
  selectedCourseCode,
  isMutating,
  onSelectCourse,
  onRemoveCourse,
  onMoveCourse,
  onStatusChange,
}: {
  term: PlannerTerm;
  allTerms: PlannerTerm[];
  selectedCourseCode: string | null;
  isMutating: boolean;
  onSelectCourse: (courseCode: string) => void;
  onRemoveCourse: (entryID: string) => Promise<void>;
  onMoveCourse: (entryID: string, targetTermID: string) => Promise<void>;
  onStatusChange: (entryID: string, status: CourseStatus) => Promise<void>;
}) {
  return (
    <div className="min-w-[320px] flex-1 rounded-[1.75rem] border border-ink/10 bg-white p-5">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-teal">{term.season}</p>
          <h3 className="mt-1 font-display text-2xl font-semibold">{term.label}</h3>
        </div>
        <div className="rounded-full bg-cloud px-3 py-1 text-xs font-semibold uppercase tracking-[0.16em] text-ink/60">
          {term.entries.length} courses
        </div>
      </div>

      <div className="mt-5 space-y-4">
        {term.entries.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-ink/15 bg-cloud px-4 py-8 text-center text-sm text-ink/55">
            No courses planned yet.
          </div>
        ) : (
          term.entries.map((entry) => (
            <article
              key={entry.id}
              className={`rounded-2xl border p-4 transition ${
                selectedCourseCode === entry.course.code
                  ? "border-teal bg-mint/35"
                  : "border-ink/10 bg-cloud"
              }`}
            >
              <button
                type="button"
                onClick={() => onSelectCourse(entry.course.code)}
                className="block w-full text-left"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-display text-lg font-semibold">{entry.course.code}</p>
                    <p className="text-sm text-ink/75">{entry.course.title}</p>
                  </div>
                  <span className={`rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] ${statusTone(entry.status)}`}>
                    {entry.status.replace("_", " ")}
                  </span>
                </div>
              </button>

              <p className="mt-3 text-sm leading-6 text-ink/65">{entry.course.description}</p>

              {entry.prerequisiteIssue ? (
                <div
                  className={`mt-4 rounded-2xl px-3 py-2 text-sm ${
                    entry.prerequisiteIssue.isSatisfied
                      ? "bg-teal/10 text-teal"
                      : "bg-rose/10 text-rose"
                  }`}
                >
                  {entry.prerequisiteIssue.message}
                </div>
              ) : null}

              <div className="mt-4 grid gap-3">
                <label className="text-xs font-semibold uppercase tracking-[0.14em] text-ink/55">
                  Status
                  <select
                    value={entry.status}
                    disabled={isMutating}
                    onChange={(event) =>
                      void onStatusChange(entry.id, event.target.value as CourseStatus)
                    }
                    className="mt-2 w-full rounded-xl border border-ink/10 bg-white px-3 py-2 text-sm outline-none transition focus:border-teal"
                  >
                    {STATUS_OPTIONS.map((status) => (
                      <option key={status} value={status}>
                        {status.replace("_", " ")}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="text-xs font-semibold uppercase tracking-[0.14em] text-ink/55">
                  Move course
                  <select
                    defaultValue=""
                    disabled={isMutating}
                    onChange={(event) => {
                      if (!event.target.value) {
                        return;
                      }
                      void onMoveCourse(entry.id, event.target.value);
                    }}
                    className="mt-2 w-full rounded-xl border border-ink/10 bg-white px-3 py-2 text-sm outline-none transition focus:border-teal"
                  >
                    <option value="">Choose another term</option>
                    {allTerms
                      .filter((candidate) => candidate.id !== term.id)
                      .map((candidate) => (
                        <option key={candidate.id} value={candidate.id}>
                          {candidate.label}
                        </option>
                      ))}
                  </select>
                </label>
              </div>

              <div className="mt-4 flex items-center justify-between gap-3">
                <span className="text-xs font-semibold uppercase tracking-[0.14em] text-ink/50">
                  {entry.course.creditWeight.toFixed(1)} credits
                </span>
                <button
                  type="button"
                  disabled={isMutating}
                  onClick={() => void onRemoveCourse(entry.id)}
                  className="rounded-full border border-rose/20 px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.16em] text-rose"
                >
                  Remove
                </button>
              </div>
            </article>
          ))
        )}
      </div>
    </div>
  );
}
