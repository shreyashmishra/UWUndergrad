export type CourseStatus = "NOT_STARTED" | "PLANNED" | "IN_PROGRESS" | "COMPLETED";
export type RequirementKind = "COURSE" | "ELECTIVE_GROUP";
export type RequirementGroupKind = "CORE" | "ELECTIVE" | "ONE_OF";
export type TermSeason = "FALL" | "WINTER" | "SPRING";

export interface University {
  code: string;
  name: string;
}

export interface Program {
  code: string;
  name: string;
  degree: string;
  description: string;
  universityCode: string;
}

export interface RoadmapCourse {
  code: string;
  title: string;
  credits: number;
  description: string | null;
  subject: string | null;
  status: CourseStatus;
  prerequisitesMet: boolean;
  prerequisiteMessage: string | null;
  notes: string | null;
  isSelected: boolean;
}

export interface ElectiveGroup {
  code: string;
  title: string;
  description: string;
  kind: RequirementGroupKind;
  minSelections: number;
  maxSelections: number;
  selectedCourseCode: string | null;
  status: CourseStatus;
  isSatisfied: boolean;
  notes: string | null;
  options: RoadmapCourse[];
}

export interface TermRequirement {
  kind: RequirementKind;
  sequence: number;
  course: RoadmapCourse | null;
  group: ElectiveGroup | null;
}

export interface RequirementSummary {
  totalRequirements: number;
  completedRequirements: number;
  inProgressRequirements: number;
  plannedRequirements: number;
  remainingRequirements: number;
  selectedElectives: number;
  completionPercentage: number;
}

export interface TermRoadmap {
  code: string;
  label: string;
  year: number;
  season: TermSeason;
  sequence: number;
  completedCount: number;
  totalCount: number;
  requirements: TermRequirement[];
}

export interface Roadmap {
  universityCode: string;
  programCode: string;
  programName: string;
  degree: string;
  description: string;
  summary: RequirementSummary;
  terms: TermRoadmap[];
}

export interface CourseStatusEntry {
  courseCode: string;
  status: CourseStatus;
}

export interface ElectiveSelectionEntry {
  groupCode: string;
  courseCode: string;
}

export interface StudentProgress {
  studentExternalKey: string;
  courseStatuses: CourseStatusEntry[];
  electiveSelections: ElectiveSelectionEntry[];
}

export interface ProgressSnapshot {
  courseStatuses: Record<string, CourseStatus>;
  electiveSelections: Record<string, string>;
}

export interface ProgramSelection {
  universityCode: string | null;
  programCode: string | null;
}
