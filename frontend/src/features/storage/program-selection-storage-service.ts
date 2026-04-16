import { readStorageValue, writeStorageValue } from "@/lib/storage/local-storage";
import type { ProgramSelection } from "@/types/roadmap";

const PROGRAM_SELECTION_KEY = "planahead.program-selection.v1";

const EMPTY_SELECTION: ProgramSelection = {
  universityCode: null,
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
