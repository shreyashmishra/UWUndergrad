export type CourseStatus = "COMPLETED" | "IN_PROGRESS" | "PLANNED";

export interface Course {
  id: string;
  code: string;
  title: string;
  description: string;
  creditWeight: number;
  offeredTerms: string | null;
  antirequisites: string | null;
}

export interface PrerequisiteIssue {
  courseCode: string;
  blockingCourseCodes: string[];
  message: string;
  isSatisfied: boolean;
}

export interface RoadmapEntry {
  id: string;
  termID: string;
  status: CourseStatus;
  sortOrder: number;
  course: Course;
  prerequisiteIssue: PrerequisiteIssue | null;
}

export interface PlannerTerm {
  id: string;
  label: string;
  season: string;
  year: number;
  sequenceNumber: number;
  entries: RoadmapEntry[];
}

export interface ProgressSummary {
  completedCredits: number;
  inProgressCredits: number;
  plannedCredits: number;
  totalCourses: number;
}

export interface PlannerUser {
  id: string;
  email: string;
  displayName: string;
  universityID: string;
}

export interface Planner {
  user: PlannerUser;
  terms: PlannerTerm[];
  progressSummary: ProgressSummary;
  catalogCourseCount: number;
}

export interface DependencyView {
  courseCode: string;
  directPrerequisites: string[];
  message: string;
}
