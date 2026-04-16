"use client";

import { useEffect, useState } from "react";

import { ProgressSummary } from "@/features/progress/components/progress-summary";
import { ProgramSelector } from "@/features/programs/components/program-selector";
import { UniversitySelector } from "@/features/programs/components/university-selector";
import { RoadmapBoard } from "@/features/roadmap/components/roadmap-board";
import { ProgramSelectionStorageService } from "@/features/storage/program-selection-storage-service";
import {
  clearElectiveSelection,
  fetchAvailableUniversities,
  fetchProgramsByUniversity,
  fetchRoadmap,
  fetchStudentProgress,
  selectElective,
  updateCourseStatus,
} from "@/lib/graphql/client";
import type {
  CourseStatus,
  ProgressSnapshot,
  Program,
  ProgramSelection,
  Roadmap,
  StudentProgress,
  University,
} from "@/types/roadmap";

const EMPTY_PROGRESS: ProgressSnapshot = {
  courseStatuses: {},
  electiveSelections: {},
};

function toSnapshot(progress: StudentProgress): ProgressSnapshot {
  return {
    courseStatuses: Object.fromEntries(
      progress.courseStatuses.map((entry) => [entry.courseCode, entry.status]),
    ),
    electiveSelections: Object.fromEntries(
      progress.electiveSelections.map((entry) => [entry.groupCode, entry.courseCode]),
    ),
  };
}

export function StudentPortal() {
  const [selection, setSelection] = useState<ProgramSelection>({
    universityCode: null,
    programCode: null,
  });
  const [universities, setUniversities] = useState<University[]>([]);
  const [programs, setPrograms] = useState<Program[]>([]);
  const [progress, setProgress] = useState<ProgressSnapshot>(EMPTY_PROGRESS);
  const [roadmap, setRoadmap] = useState<Roadmap | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRoadmapLoading, setIsRoadmapLoading] = useState(false);
  const [isProgressMutating, setIsProgressMutating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    const storedSelection = ProgramSelectionStorageService.get();

    const loadUniversities = async () => {
      try {
        const availableUniversities = await fetchAvailableUniversities();
        if (cancelled) {
          return;
        }

        setUniversities(availableUniversities);
        const fallbackUniversityCode =
          storedSelection.universityCode &&
          availableUniversities.some(
            (university) => university.code === storedSelection.universityCode,
          )
            ? storedSelection.universityCode
            : availableUniversities[0]?.code ?? null;

        const nextSelection = {
          universityCode: fallbackUniversityCode,
          programCode: storedSelection.programCode,
        };
        ProgramSelectionStorageService.set(nextSelection);
        setSelection(nextSelection);
        setError(null);
      } catch (loadError) {
        if (!cancelled) {
          setError(
            loadError instanceof Error
              ? loadError.message
              : "Unable to load universities from the backend.",
          );
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void loadUniversities();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!selection.universityCode) {
      setPrograms([]);
      return;
    }

    let cancelled = false;
    const loadPrograms = async () => {
      try {
        const universityPrograms = await fetchProgramsByUniversity(selection.universityCode!);
        if (cancelled) {
          return;
        }

        setPrograms(universityPrograms);
        const fallbackProgramCode =
          selection.programCode &&
          universityPrograms.some((item) => item.code === selection.programCode)
            ? selection.programCode
            : universityPrograms[0]?.code ?? null;

        const nextSelection = {
          universityCode: selection.universityCode,
          programCode: fallbackProgramCode,
        };
        ProgramSelectionStorageService.set(nextSelection);
        setSelection(nextSelection);
      } catch (loadError) {
        if (!cancelled) {
          setError(
            loadError instanceof Error
              ? loadError.message
              : "Unable to load programs from the backend.",
          );
        }
      }
    };

    void loadPrograms();
    return () => {
      cancelled = true;
    };
  }, [selection.universityCode]);

  useEffect(() => {
    if (!selection.universityCode || !selection.programCode) {
      setProgress(EMPTY_PROGRESS);
      setRoadmap(null);
      return;
    }

    let cancelled = false;
    const loadProgress = async () => {
      try {
        const seededProgress = await fetchStudentProgress(
          selection.universityCode!,
          selection.programCode!,
        );
        if (cancelled) {
          return;
        }
        setProgress(toSnapshot(seededProgress));
      } catch {
        if (!cancelled) {
          setProgress(EMPTY_PROGRESS);
        }
      }
    };

    void loadProgress();
    return () => {
      cancelled = true;
    };
  }, [selection.programCode, selection.universityCode]);

  useEffect(() => {
    if (!selection.universityCode || !selection.programCode) {
      return;
    }

    let cancelled = false;
    const loadRoadmap = async () => {
      try {
        setIsRoadmapLoading(true);
        const nextRoadmap = await fetchRoadmap(
          selection.universityCode!,
          selection.programCode!,
          progress,
        );
        if (!cancelled) {
          setRoadmap(nextRoadmap);
          setError(null);
        }
      } catch (loadError) {
        if (!cancelled) {
          setError(
            loadError instanceof Error
              ? loadError.message
              : "Unable to load roadmap data from the backend.",
          );
        }
      } finally {
        if (!cancelled) {
          setIsRoadmapLoading(false);
        }
      }
    };

    void loadRoadmap();
    return () => {
      cancelled = true;
    };
  }, [progress, selection.programCode, selection.universityCode]);

  const handleUniversityChange = (universityCode: string) => {
    const nextSelection = {
      universityCode,
      programCode: null,
    };
    ProgramSelectionStorageService.set(nextSelection);
    setPrograms([]);
    setProgress(EMPTY_PROGRESS);
    setRoadmap(null);
    setSelection(nextSelection);
  };

  const handleProgramChange = (programCode: string) => {
    const nextSelection = {
      universityCode: selection.universityCode,
      programCode,
    };
    ProgramSelectionStorageService.set(nextSelection);
    setRoadmap(null);
    setProgress(EMPTY_PROGRESS);
    setSelection(nextSelection);
  };

  const syncProgress = async (operation: () => Promise<StudentProgress>) => {
    if (!selection.universityCode || !selection.programCode) {
      return;
    }

    try {
      setIsProgressMutating(true);
      const nextProgress = await operation();
      setProgress(toSnapshot(nextProgress));
      setError(null);
    } catch (updateError) {
      setError(
        updateError instanceof Error
          ? updateError.message
          : "Unable to update roadmap progress.",
      );
    } finally {
      setIsProgressMutating(false);
    }
  };

  const handleCourseStatusChange = (courseCode: string, status: CourseStatus) => {
    void syncProgress(() =>
      updateCourseStatus(
        selection.universityCode!,
        selection.programCode!,
        courseCode,
        status,
      ),
    );
  };

  const handleElectiveSelect = (groupCode: string, courseCode: string) => {
    void syncProgress(() =>
      selectElective(
        selection.universityCode!,
        selection.programCode!,
        groupCode,
        courseCode,
      ),
    );
  };

  const handleElectiveStatusChange = (
    groupCode: string,
    courseCode: string,
    status: CourseStatus,
  ) => {
    void syncProgress(async () => {
      await selectElective(
        selection.universityCode!,
        selection.programCode!,
        groupCode,
        courseCode,
      );
      return updateCourseStatus(
        selection.universityCode!,
        selection.programCode!,
        courseCode,
        status,
      );
    });
  };

  const handleElectiveClear = (groupCode: string) => {
    void syncProgress(() =>
      clearElectiveSelection(
        selection.universityCode!,
        selection.programCode!,
        groupCode,
      ),
    );
  };

  const selectedProgram = programs.find((program) => program.code === selection.programCode) ?? null;
  const selectedUniversity =
    universities.find((university) => university.code === selection.universityCode) ?? null;

  return (
    <main className="min-h-screen bg-hero-gradient px-4 py-6 text-ink sm:px-6 lg:px-8">
      <div className="mx-auto max-w-[1500px]">
        <section className="rounded-[2.4rem] border border-white/60 bg-white/55 p-6 shadow-panel backdrop-blur sm:p-8">
          <div className="grid gap-6 xl:grid-cols-[360px,minmax(0,1fr)]">
            <aside className="space-y-5">
              <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                <span className="rounded-full bg-ink px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-white">
                  PlanAhead MVP
                </span>
                <h1 className="mt-4 font-display text-4xl leading-tight text-ink">
                  Degree requirement tracking that feels like a real semester plan.
                </h1>
                <p className="mt-4 text-sm leading-7 text-ink/70">
                  Start in the portal, choose a university and program, and map completed,
                  in-progress, or planned work directly onto a term-by-term roadmap.
                </p>
              </div>

              <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                <div className="space-y-4">
                  <div className="rounded-[1.6rem] bg-cloud px-4 py-4">
                    <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/55">
                      Catalog source
                    </p>
                    <h2 className="mt-3 font-display text-2xl text-ink">
                      {selectedUniversity?.name ?? "Loading universities"}
                    </h2>
                    <p className="mt-2 text-sm leading-6 text-ink/70">
                      The Go backend now drives the active university and program list from MySQL, so the planner no longer boots into a single hardcoded roadmap.
                    </p>
                  </div>
                  <UniversitySelector
                    universities={universities}
                    selectedCode={selection.universityCode}
                    onChange={handleUniversityChange}
                  />
                  <ProgramSelector
                    programs={programs}
                    selectedCode={selection.programCode}
                    onChange={handleProgramChange}
                  />
                </div>

                <div className="mt-5 rounded-[1.6rem] bg-cloud px-4 py-4">
                  <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/55">
                    Active plan
                  </p>
                  <h2 className="mt-3 font-display text-2xl text-ink">
                    {selectedProgram?.name ?? "Select a program"}
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-ink/70">
                    {selectedProgram?.degree ??
                      "Program details appear here once a selection is made."}
                  </p>
                </div>
              </div>

              <ProgressSummary summary={roadmap?.summary ?? null} />
            </aside>

            <section className="space-y-5">
              <div className="rounded-[2rem] border border-white/70 bg-white/80 p-6 shadow-panel backdrop-blur">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
                      Dashboard
                    </p>
                    <h2 className="mt-2 font-display text-3xl text-ink">
                      {roadmap?.programName ?? "Roadmap loading"}
                    </h2>
                  </div>
                  <div className="rounded-full bg-mint px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-teal">
                    {isLoading || isRoadmapLoading || isProgressMutating
                      ? "Refreshing roadmap"
                      : "Live MySQL-backed progress"}
                  </div>
                </div>
                <p className="mt-4 max-w-3xl text-sm leading-7 text-ink/70">
                  {roadmap?.description ??
                    "Once the roadmap is loaded, the Go API evaluates remaining requirements and unmet prerequisites against the MySQL-backed student snapshot."}
                </p>
              </div>

              {error ? (
                <div className="rounded-[1.6rem] border border-rose/25 bg-rose/10 px-5 py-4 text-sm text-rose">
                  {error}
                </div>
              ) : null}

              <RoadmapBoard
                roadmap={roadmap}
                onCourseStatusChange={handleCourseStatusChange}
                onElectiveSelect={handleElectiveSelect}
                onElectiveStatusChange={handleElectiveStatusChange}
                onElectiveClear={handleElectiveClear}
              />
            </section>
          </div>
        </section>
      </div>
    </main>
  );
}
