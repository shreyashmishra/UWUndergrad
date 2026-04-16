"use client";

import { useEffect, useState } from "react";

import { ProgressSummary } from "@/features/progress/components/progress-summary";
import { ProgramSelector } from "@/features/programs/components/program-selector";
import { RoadmapBoard } from "@/features/roadmap/components/roadmap-board";
import {
  ProgramSelectionStorageService,
  WATERLOO_UNIVERSITY_CODE,
} from "@/features/storage/program-selection-storage-service";
import {
  clearElectiveSelection,
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
    universityCode: WATERLOO_UNIVERSITY_CODE,
    programCode: null,
  });
  const [programs, setPrograms] = useState<Program[]>([]);
  const [progress, setProgress] = useState<ProgressSnapshot>(EMPTY_PROGRESS);
  const [roadmap, setRoadmap] = useState<Roadmap | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isProgramsLoading, setIsProgramsLoading] = useState(true);
  const [isRoadmapLoading, setIsRoadmapLoading] = useState(false);
  const [isProgressMutating, setIsProgressMutating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const storedSelection = ProgramSelectionStorageService.get();
    const nextSelection = {
      universityCode: WATERLOO_UNIVERSITY_CODE,
      programCode: storedSelection.programCode,
    };
    ProgramSelectionStorageService.set(nextSelection);
    setSelection(nextSelection);
    setIsLoading(false);
  }, []);

  useEffect(() => {
    if (!selection.universityCode) {
      setPrograms([]);
      return;
    }

    let cancelled = false;
    const loadPrograms = async () => {
      try {
        setIsProgramsLoading(true);
        const universityPrograms = await fetchProgramsByUniversity(selection.universityCode!);
        if (cancelled) return;

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
      } finally {
        if (!cancelled) setIsProgramsLoading(false);
      }
    };

    void loadPrograms();
    return () => { cancelled = true; };
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
        if (cancelled) return;
        setProgress(toSnapshot(seededProgress));
      } catch {
        if (!cancelled) setProgress(EMPTY_PROGRESS);
      }
    };

    void loadProgress();
    return () => { cancelled = true; };
  }, [selection.programCode, selection.universityCode]);

  useEffect(() => {
    if (!selection.universityCode || !selection.programCode) return;

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
        if (!cancelled) setIsRoadmapLoading(false);
      }
    };

    void loadRoadmap();
    return () => { cancelled = true; };
  }, [progress, selection.programCode, selection.universityCode]);

  const handleProgramChange = (programCode: string) => {
    const nextSelection = { universityCode: WATERLOO_UNIVERSITY_CODE, programCode };
    ProgramSelectionStorageService.set(nextSelection);
    setRoadmap(null);
    setProgress(EMPTY_PROGRESS);
    setSelection(nextSelection);
  };

  const syncProgress = async (operation: () => Promise<StudentProgress>) => {
    if (!selection.universityCode || !selection.programCode) return;
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
      updateCourseStatus(selection.universityCode!, selection.programCode!, courseCode, status),
    );
  };

  const handleElectiveSelect = (groupCode: string, courseCode: string) => {
    void syncProgress(() =>
      selectElective(selection.universityCode!, selection.programCode!, groupCode, courseCode),
    );
  };

  const handleElectiveStatusChange = (
    groupCode: string,
    courseCode: string,
    status: CourseStatus,
  ) => {
    void syncProgress(async () => {
      await selectElective(selection.universityCode!, selection.programCode!, groupCode, courseCode);
      return updateCourseStatus(selection.universityCode!, selection.programCode!, courseCode, status);
    });
  };

  const handleElectiveClear = (groupCode: string) => {
    void syncProgress(() =>
      clearElectiveSelection(selection.universityCode!, selection.programCode!, groupCode),
    );
  };

  const selectedProgram = programs.find((p) => p.code === selection.programCode) ?? null;
  const isBusy = isLoading || isRoadmapLoading || isProgressMutating;

  return (
    <main className="min-h-screen bg-hero-gradient px-4 py-6 text-ink sm:px-6 lg:px-8">
      <div className="mx-auto max-w-[1500px]">

        {/* Top header bar */}
        <div className="mb-5 flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-ink">
              <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
                <path
                  d="M2 14L9 4l7 10H2z"
                  fill="white"
                  fillOpacity="0.9"
                />
              </svg>
            </div>
            <span className="font-display text-xl text-ink">PlanAhead</span>
          </div>
          <div
            className={`rounded-full px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.18em] transition ${
              isProgramsLoading
                ? "bg-sand text-amber"
                : isBusy
                  ? "bg-mint text-teal"
                  : "bg-ink/8 text-ink/60"
            }`}
          >
            {isProgramsLoading
              ? "Loading catalog"
              : isBusy
                ? "Refreshing"
                : "University of Waterloo"}
          </div>
        </div>

        <section className="rounded-[2.4rem] border border-white/60 bg-white/55 p-5 shadow-panel backdrop-blur sm:p-6">
          <div className="grid gap-5 xl:grid-cols-[320px,minmax(0,1fr)]">

            {/* Sidebar */}
            <aside className="space-y-4">

              {/* Program selector */}
              <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                <h1 className="mb-1 font-display text-2xl leading-tight text-ink">
                  Plan your degree
                </h1>
                <p className="mb-5 text-sm leading-6 text-ink/60">
                  Choose your Waterloo program and track your progress from 1A to graduation.
                </p>
                <ProgramSelector
                  programs={programs}
                  selectedCode={selection.programCode}
                  onChange={handleProgramChange}
                  disabled={isProgramsLoading || programs.length === 0}
                  loading={isProgramsLoading}
                />
              </div>

              {/* Active program info */}
              {(selectedProgram || isProgramsLoading) && (
                <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                  <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
                    Active plan
                  </p>
                  <h2 className="mt-2 font-display text-2xl leading-tight text-ink">
                    {selectedProgram
                      ? `${selectedProgram.code} – ${selectedProgram.name}`
                      : "Loading…"}
                  </h2>
                  {selectedProgram?.degree && (
                    <p className="mt-1.5 text-sm text-ink/60">{selectedProgram.degree}</p>
                  )}
                  {selectedProgram?.description && (
                    <p className="mt-3 text-sm leading-6 text-ink/55">
                      {selectedProgram.description}
                    </p>
                  )}
                </div>
              )}

              {/* Progress summary */}
              <ProgressSummary summary={roadmap?.summary ?? null} />
            </aside>

            {/* Main roadmap area */}
            <section className="space-y-4">
              {/* Roadmap header */}
              <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
                      Roadmap
                    </p>
                    <h2 className="mt-1.5 font-display text-2xl text-ink">
                      {roadmap?.programName ??
                        (isProgramsLoading
                          ? "Loading programs…"
                          : "Select a program to begin")}
                    </h2>
                  </div>
                  {roadmap && (
                    <div className="flex items-center gap-2 rounded-2xl bg-cloud px-4 py-2">
                      <div className="h-2 w-2 rounded-full bg-teal" />
                      <span className="text-xs font-semibold text-teal">
                        {roadmap.summary.completionPercentage}% complete
                      </span>
                    </div>
                  )}
                </div>
                {roadmap?.description && (
                  <p className="mt-3 max-w-3xl text-sm leading-6 text-ink/60">
                    {roadmap.description}
                  </p>
                )}
              </div>

              {error && (
                <div className="rounded-[1.6rem] border border-rose/25 bg-rose/10 px-5 py-4 text-sm text-rose">
                  {error}
                </div>
              )}

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
