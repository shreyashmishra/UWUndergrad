import type { CourseStatus, ProgressSnapshot } from "@/types/roadmap";
import { readStorageValue, writeStorageValue } from "@/lib/storage/local-storage";

interface ProgressScope {
  universityCode: string;
  programCode: string;
}

const buildKey = ({ universityCode, programCode }: ProgressScope): string =>
  `planahead.progress.v1.${universityCode}.${programCode}`;

const EMPTY_PROGRESS: ProgressSnapshot = {
  courseStatuses: {},
  electiveSelections: {},
};

export class ProgressStorageService {
  static get(scope: ProgressScope): ProgressSnapshot {
    return readStorageValue(buildKey(scope), EMPTY_PROGRESS);
  }

  static set(scope: ProgressScope, snapshot: ProgressSnapshot): ProgressSnapshot {
    writeStorageValue(buildKey(scope), snapshot);
    return snapshot;
  }

  static bootstrap(scope: ProgressScope, snapshot: ProgressSnapshot): ProgressSnapshot {
    const existing = this.get(scope);
    if (Object.keys(existing.courseStatuses).length || Object.keys(existing.electiveSelections).length) {
      return existing;
    }

    return this.set(scope, snapshot);
  }

  static updateCourseStatus(scope: ProgressScope, courseCode: string, status: CourseStatus): ProgressSnapshot {
    const current = this.get(scope);
    const nextCourseStatuses = { ...current.courseStatuses };

    if (status === "NOT_STARTED") {
      delete nextCourseStatuses[courseCode];
    } else {
      nextCourseStatuses[courseCode] = status;
    }

    return this.set(scope, {
      ...current,
      courseStatuses: nextCourseStatuses,
    });
  }

  static selectElective(scope: ProgressScope, groupCode: string, courseCode: string): ProgressSnapshot {
    const current = this.get(scope);
    return this.set(scope, {
      ...current,
      electiveSelections: {
        ...current.electiveSelections,
        [groupCode]: courseCode,
      },
    });
  }

  static clearElectiveSelection(scope: ProgressScope, groupCode: string): ProgressSnapshot {
    const current = this.get(scope);
    const nextSelections = { ...current.electiveSelections };
    delete nextSelections[groupCode];

    return this.set(scope, {
      ...current,
      electiveSelections: nextSelections,
    });
  }
}
