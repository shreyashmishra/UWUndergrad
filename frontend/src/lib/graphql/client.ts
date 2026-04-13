import type {
  ProgressSnapshot,
  Program,
  Roadmap,
  StudentProgress,
  University,
} from "@/types/roadmap";

const GRAPHQL_ENDPOINT =
  process.env.NEXT_PUBLIC_GRAPHQL_API_URL ?? "http://localhost:8000/graphql";

async function graphqlRequest<TData>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<TData> {
  const response = await fetch(GRAPHQL_ENDPOINT, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ query, variables }),
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`GraphQL request failed with status ${response.status}.`);
  }

  const payload = (await response.json()) as {
    data?: TData;
    errors?: Array<{ message: string }>;
  };

  if (payload.errors?.length) {
    throw new Error(payload.errors.map((error) => error.message).join(", "));
  }

  if (!payload.data) {
    throw new Error("GraphQL request returned no data.");
  }

  return payload.data;
}

function toProgressInput(snapshot: ProgressSnapshot): Record<string, unknown> {
  return {
    courseStatuses: Object.entries(snapshot.courseStatuses).map(([courseCode, status]) => ({
      courseCode,
      status,
    })),
    electiveSelections: Object.entries(snapshot.electiveSelections).map(([groupCode, courseCode]) => ({
      groupCode,
      courseCode,
    })),
  };
}

export async function fetchAvailableUniversities(): Promise<University[]> {
  const data = await graphqlRequest<{ availableUniversities: University[] }>(`
    query AvailableUniversities {
      availableUniversities {
        code
        name
      }
    }
  `);

  return data.availableUniversities;
}

export async function fetchProgramsByUniversity(universityCode: string): Promise<Program[]> {
  const data = await graphqlRequest<{ programsByUniversity: Program[] }>(
    `
      query ProgramsByUniversity($universityCode: String!) {
        programsByUniversity(universityCode: $universityCode) {
          code
          name
          degree
          description
          universityCode
        }
      }
    `,
    { universityCode },
  );

  return data.programsByUniversity;
}

export async function fetchRoadmap(
  universityCode: string,
  programCode: string,
  snapshot: ProgressSnapshot,
): Promise<Roadmap> {
  const data = await graphqlRequest<{ roadmapByProgram: Roadmap }>(
    `
      query RoadmapByProgram(
        $universityCode: String!
        $programCode: String!
        $progressInput: ProgressInput
      ) {
        roadmapByProgram(
          universityCode: $universityCode
          programCode: $programCode
          progressInput: $progressInput
        ) {
          universityCode
          programCode
          programName
          degree
          description
          summary {
            totalRequirements
            completedRequirements
            inProgressRequirements
            plannedRequirements
            remainingRequirements
            selectedElectives
            completionPercentage
          }
          terms {
            code
            label
            year
            season
            sequence
            completedCount
            totalCount
            requirements {
              kind
              sequence
              course {
                code
                title
                credits
                description
                subject
                status
                prerequisitesMet
                prerequisiteMessage
                notes
                isSelected
              }
              group {
                code
                title
                description
                kind
                minSelections
                maxSelections
                selectedCourseCode
                status
                isSatisfied
                notes
                options {
                  code
                  title
                  credits
                  description
                  subject
                  status
                  prerequisitesMet
                  prerequisiteMessage
                  notes
                  isSelected
                }
              }
            }
          }
        }
      }
    `,
    {
      universityCode,
      programCode,
      progressInput: toProgressInput(snapshot),
    },
  );

  return data.roadmapByProgram;
}

export async function fetchStudentProgress(
  universityCode: string,
  programCode: string,
): Promise<StudentProgress> {
  const data = await graphqlRequest<{ studentProgress: StudentProgress }>(
    `
      query StudentProgress($universityCode: String!, $programCode: String!) {
        studentProgress(universityCode: $universityCode, programCode: $programCode) {
          studentExternalKey
          courseStatuses {
            courseCode
            status
          }
          electiveSelections {
            groupCode
            courseCode
          }
        }
      }
    `,
    { universityCode, programCode },
  );

  return data.studentProgress;
}
