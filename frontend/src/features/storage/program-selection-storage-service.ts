import { readStorageValue, writeStorageValue } from "@/lib/storage/local-storage";
import type { ProgramSelection } from "@/types/roadmap";

const PROGRAM_SELECTION_KEY = "planahead.program-selection.v1";
export const WATERLOO_UNIVERSITY_CODE = "WATERLOO";

const EMPTY_SELECTION: ProgramSelection = {
  universityCode: WATERLOO_UNIVERSITY_CODE,
  programCode: null,
};

export class ProgramSelectionStorageService {
  static get(): ProgramSelection {
    return readStorageValue(PROGRAM_SELECTION_KEY, EMPTY_SELECTION);
  }

  static set(selection: ProgramSelection): ProgramSelection {
    writeStorageValue(PROGRAM_SELECTION_KEY, selection);
    return selection;
  }
}
