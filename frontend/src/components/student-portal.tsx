"use client";

import { useEffect, useState } from "react";

import { ProgressSummary } from "@/features/progress/components/progress-summary";
import { ProgramSelector } from "@/features/programs/components/program-selector";
import { RoadmapBoard } from "@/features/roadmap/components/roadmap-board";
import { RoadmapGraph } from "@/features/roadmap/components/roadmap-graph";
import { AuthModal } from "./auth-modal";
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

  const [viewMode, setViewMode] = useState<"board" | "graph">("graph");
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [userName, setUserName] = useState<string | null>(null);
  const [currentUserKey, setCurrentUserKey] = useState<string | null>(null);
  const [isGuest, setIsGuest] = useState(false);

  const [activeTab, setActiveTab] = useState<"explore" | "my-plans">("explore");
  const [savedPlans, setSavedPlans] = useState<string[]>([]);

  // Load saved plans for the user
  useEffect(() => {
    if (currentUserKey) {
      try {
        const stored = localStorage.getItem(`savedPlans_${currentUserKey}`);
        if (stored) {
          setSavedPlans(JSON.parse(stored));
        } else {
          setSavedPlans([]);
        }
      } catch {
        setSavedPlans([]);
      }
    } else {
      setSavedPlans([]);
    }
  }, [currentUserKey]);

  const toggleSavePlan = (code: string) => {
    if (!currentUserKey) return;
    const isSaved = savedPlans.includes(code);
    const nextSaved = isSaved ? savedPlans.filter(p => p !== code) : [...savedPlans, code];
    setSavedPlans(nextSaved);
    localStorage.setItem(`savedPlans_${currentUserKey}`, JSON.stringify(nextSaved));
  };

  useEffect(() => {
    const storedSelection = ProgramSelectionStorageService.get();
    const nextSelection = {
      universityCode: WATERLOO_UNIVERSITY_CODE,
      programCode: storedSelection.programCode,
    };
    ProgramSelectionStorageService.set(nextSelection);
    setSelection(nextSelection);
    
    // Check auth
    if (typeof window !== "undefined") {
      const externalKey = localStorage.getItem("externalKey");
      if (externalKey) {
        setUserName("Student"); 
        setCurrentUserKey(externalKey);
      }
    }

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
      if (isGuest) {
        setProgress(EMPTY_PROGRESS);
        return;
      }
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
  }, [selection.programCode, selection.universityCode, isGuest]);

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
    if (!isGuest) {
      ProgramSelectionStorageService.set(nextSelection);
    }
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
    if (isGuest) {
      setProgress(prev => {
        const next = { ...prev, courseStatuses: { ...prev.courseStatuses } };
        if (status === "NOT_STARTED") {
          delete next.courseStatuses[courseCode];
        } else {
          next.courseStatuses[courseCode] = status;
        }
        return next;
      });
      return;
    }
    void syncProgress(() =>
      updateCourseStatus(selection.universityCode!, selection.programCode!, courseCode, status),
    );
  };

  const handleElectiveSelect = (groupCode: string, courseCode: string) => {
    if (isGuest) {
      setProgress(prev => {
        const next = { ...prev, electiveSelections: { ...prev.electiveSelections } };
        next.electiveSelections[groupCode] = courseCode;
        return next;
      });
      return;
    }
    void syncProgress(() =>
      selectElective(selection.universityCode!, selection.programCode!, groupCode, courseCode),
    );
  };

  const handleElectiveStatusChange = (
    groupCode: string,
    courseCode: string,
    status: CourseStatus,
  ) => {
    if (isGuest) {
      setProgress(prev => {
        const next = { ...prev, electiveSelections: { ...prev.electiveSelections }, courseStatuses: { ...prev.courseStatuses } };
        next.electiveSelections[groupCode] = courseCode;
        if (status === "NOT_STARTED") {
          delete next.courseStatuses[courseCode];
        } else {
          next.courseStatuses[courseCode] = status;
        }
        return next;
      });
      return;
    }
    void syncProgress(async () => {
      await selectElective(selection.universityCode!, selection.programCode!, groupCode, courseCode);
      return updateCourseStatus(selection.universityCode!, selection.programCode!, courseCode, status);
    });
  };

  const handleElectiveClear = (groupCode: string) => {
    if (isGuest) {
      setProgress(prev => {
        const next = { ...prev, electiveSelections: { ...prev.electiveSelections } };
        delete next.electiveSelections[groupCode];
        return next;
      });
      return;
    }
    void syncProgress(() =>
      clearElectiveSelection(selection.universityCode!, selection.programCode!, groupCode),
    );
  };

  const selectedProgram = programs.find((p) => p.code === selection.programCode) ?? null;
  const isBusy = isLoading || isRoadmapLoading || isProgressMutating;

  if (!isLoading && !currentUserKey && !isGuest) {
    return (
      <main className="min-h-screen bg-hero-gradient px-4 py-8 text-ink flex flex-col items-center justify-center">
        {showAuthModal && (
          <AuthModal
            onClose={() => setShowAuthModal(false)}
            onSuccess={(name) => {
              setUserName(name);
              setCurrentUserKey(localStorage.getItem("externalKey"));
              setShowAuthModal(false);
              // Trigger reload
              setRoadmap(null);
              setProgress(EMPTY_PROGRESS);
              setSelection({ ...selection });
            }}
          />
        )}
        <div className="mx-auto max-w-4xl text-center">
          <div className="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-[2rem] bg-ink text-white shadow-xl">
             <svg width="40" height="40" viewBox="0 0 18 18" fill="none">
                <path d="M2 14L9 4l7 10H2z" fill="white" fillOpacity="0.9" />
             </svg>
          </div>
          <h1 className="mb-6 font-display text-5xl font-bold tracking-tight text-ink md:text-7xl">
            Map your future with <span className="text-teal">PlanAhead</span>
          </h1>
          <p className="mb-10 text-lg leading-relaxed text-ink/70 md:text-xl max-w-2xl mx-auto">
            The ultimate degree planner designed specifically for University of Waterloo students. Track your prerequisites, plot your terms effortlessly on an interactive web, and visually build your path to graduation.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <button
              onClick={() => setShowAuthModal(true)}
              className="rounded-2xl bg-teal px-8 py-4 font-display text-lg font-semibold text-white shadow-lg transition duration-200 hover:-translate-y-1 hover:bg-teal/90 hover:shadow-xl w-full sm:w-auto"
            >
              Log In to Start Planning
            </button>
            <button
              onClick={() => {
                localStorage.removeItem("token");
                localStorage.removeItem("externalKey");
                setUserName(null);
                setCurrentUserKey(null);
                setIsGuest(true);
                setActiveTab("explore");
              }}
              className="rounded-2xl bg-white px-8 py-4 font-display text-lg font-semibold text-ink shadow-lg transition duration-200 hover:-translate-y-1 hover:bg-cloud hover:shadow-xl w-full sm:w-auto"
            >
              Plan without an account
            </button>
          </div>
        </div>
      </main>
    );
  }

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

          <div className="flex items-center gap-4">
            {userName ? (
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-ink">Hi, {userName}</span>
                <button
                  onClick={() => {
                    localStorage.removeItem("token");
                    localStorage.removeItem("externalKey");
                    setUserName(null);
                    setCurrentUserKey(null);
                    setRoadmap(null);
                    setProgress(EMPTY_PROGRESS);
                  }}
                  className="rounded-xl px-4 py-2 text-sm font-semibold text-rose transition hover:bg-rose/10"
                >
                  Log Out
                </button>
              </div>
            ) : (
              <button
                onClick={() => setShowAuthModal(true)}
                className="rounded-xl bg-teal px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-teal/90"
              >
                Log In
              </button>
            )}
          </div>
        </div>

        {showAuthModal && (
          <AuthModal
            onClose={() => setShowAuthModal(false)}
            onSuccess={(name) => {
              setUserName(name);
              setCurrentUserKey(localStorage.getItem("externalKey"));
              setShowAuthModal(false);
              // Trigger reload
              setRoadmap(null);
              setProgress(EMPTY_PROGRESS);
              setSelection({ ...selection }); // prompt refetch
            }}
          />
        )}

        <section className="rounded-[2.4rem] border border-white/60 bg-white/55 p-5 shadow-panel backdrop-blur sm:p-6">
          <div className="grid gap-5 xl:grid-cols-[320px,minmax(0,1fr)]">

            {/* Sidebar */}
            <aside className="space-y-4">

              {/* Tabs */}
              {userName && (
                <div className="flex w-full items-center gap-2 rounded-[2rem] border border-white/70 bg-white/80 p-2 shadow-panel backdrop-blur">
                  <button
                    onClick={() => setActiveTab("explore")}
                    className={`flex-1 rounded-3xl py-2 text-sm font-semibold transition ${
                      activeTab === "explore" ? "bg-teal text-white shadow" : "text-ink/60 hover:bg-ink/5 hover:text-ink"
                    }`}
                  >
                    Explore
                  </button>
                  <button
                    onClick={() => setActiveTab("my-plans")}
                    className={`flex-1 rounded-3xl py-2 text-sm font-semibold transition ${
                      activeTab === "my-plans" ? "bg-teal text-white shadow" : "text-ink/60 hover:bg-ink/5 hover:text-ink"
                    }`}
                  >
                    My Plans
                  </button>
                </div>
              )}

              {activeTab === "my-plans" && userName ? (
                <div className="relative z-10 rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                  <h1 className="mb-4 font-display text-2xl leading-tight text-ink">
                    My Saved Plans
                  </h1>
                  {savedPlans.length === 0 ? (
                    <p className="text-sm text-ink/60">You haven't saved any plans yet.</p>
                  ) : (
                    <div className="space-y-3">
                      {savedPlans.map(code => {
                        const prog = programs.find(p => p.code === code);
                        return (
                          <div 
                            key={code} 
                            onClick={() => handleProgramChange(code)}
                            className={`cursor-pointer rounded-xl border p-4 transition hover:border-teal/30 hover:bg-teal/5 ${
                              selection.programCode === code ? "border-teal/50 bg-teal/5 shadow-sm" : "border-ink/10 bg-white"
                            }`}
                          >
                            <h3 className="font-semibold text-ink">{prog?.name ?? code}</h3>
                            {prog?.degree && <p className="text-xs text-ink/60">{prog.degree}</p>}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              ) : (
                <div className="relative z-10 rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
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
              )}

              {/* Active program info */}
              {(selectedProgram || isProgramsLoading) && activeTab === "explore" && (
                <div className="rounded-[2rem] border border-white/70 bg-white/80 p-5 shadow-panel backdrop-blur">
                  <div className="flex items-start justify-between gap-3">
                    <p className="text-xs font-semibold uppercase tracking-[0.24em] text-ink/50">
                      Active plan
                    </p>
                    {userName && selectedProgram && (
                      <button 
                        onClick={() => toggleSavePlan(selectedProgram.code)}
                        className={`rounded-lg px-3 py-1 text-xs font-semibold transition ${
                          savedPlans.includes(selectedProgram.code) 
                            ? "bg-teal/10 text-teal" 
                            : "bg-ink/5 text-ink/60 hover:bg-ink/10 hover:text-ink"
                        }`}
                      >
                        {savedPlans.includes(selectedProgram.code) ? "Saved" : "Save Plan"}
                      </button>
                    )}
                  </div>
                  <h2 className="mt-2 font-display text-2xl leading-tight text-ink">
                    {selectedProgram?.name ?? "Loading…"}
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
                  <div className="flex items-center gap-3">
                    {roadmap && (
                      <div className="flex items-center gap-1 rounded-2xl bg-ink/5 p-1">
                        <button
                          onClick={() => setViewMode("board")}
                          className={`rounded-xl px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.1em] transition ${
                            viewMode === "board" ? "bg-white text-ink shadow-sm" : "text-ink/50 hover:text-ink/80"
                          }`}
                        >
                          Board
                        </button>
                        <button
                          onClick={() => setViewMode("graph")}
                          className={`rounded-xl px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.1em] transition ${
                            viewMode === "graph" ? "bg-white text-ink shadow-sm" : "text-ink/50 hover:text-ink/80"
                          }`}
                        >
                          Graph
                        </button>
                      </div>
                    )}
                    {roadmap && (
                      <div className="flex items-center gap-2 rounded-2xl bg-cloud px-4 py-2">
                        <div className="h-2 w-2 rounded-full bg-teal" />
                        <span className="text-xs font-semibold text-teal">
                          {Math.round(roadmap.summary.completionPercentage)}% complete
                        </span>
                      </div>
                    )}
                  </div>
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

              {viewMode === "board" ? (
                <RoadmapBoard
                  roadmap={roadmap}
                  onCourseStatusChange={handleCourseStatusChange}
                  onElectiveSelect={handleElectiveSelect}
                  onElectiveStatusChange={handleElectiveStatusChange}
                  onElectiveClear={handleElectiveClear}
                  currentUserKey={currentUserKey}
                  isGuest={isGuest}
                />
              ) : (
                <RoadmapGraph roadmap={roadmap} />
              )}
            </section>
          </div>
        </section>
      </div>
    </main>
  );
}
