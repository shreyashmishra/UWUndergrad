import type { CourseStatus, DependencyView, Planner, Course } from "@/features/planner/lib/types";

const GRAPHQL_ENDPOINT =
  process.env.NEXT_PUBLIC_PLANNER_API_URL ?? "http://localhost:8080/graphql";

export const MOCK_USER_ID = "demo-user";

const plannerFields = `
  user {
    id
    email
    displayName
    universityID
  }
  catalogCourseCount
  progressSummary {
    completedCredits
    inProgressCredits
    plannedCredits
    totalCourses
  }
  terms {
    id
    label
    season
    year
    sequenceNumber
    entries {
      id
      termID
      status
      sortOrder
      course {
        id
        code
        title
        description
        creditWeight
        offeredTerms
        antirequisites
      }
      prerequisiteIssue {
        courseCode
        blockingCourseCodes
        message
        isSatisfied
      }
    }
  }
`;

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

export async function fetchPlanner(): Promise<Planner> {
  const data = await graphqlRequest<{ planner: Planner }>(
    `
      query Planner($userID: ID) {
        planner(userID: $userID) {
          ${plannerFields}
        }
      }
    `,
    { userID: MOCK_USER_ID },
  );

  return data.planner;
}

export async function searchCourses(query: string): Promise<Course[]> {
  const trimmed = query.trim();
  if (!trimmed) {
    return [];
  }

  const data = await graphqlRequest<{ searchCourses: Course[] }>(
    `
      query SearchCourses($query: String!) {
        searchCourses(query: $query) {
          id
          code
          title
          description
          creditWeight
          offeredTerms
          antirequisites
        }
      }
    `,
    { query: trimmed },
  );

  return data.searchCourses;
}

export async function fetchDependencyView(courseCode: string): Promise<DependencyView> {
  const data = await graphqlRequest<{ dependencyView: DependencyView }>(
    `
      query DependencyView($courseCode: String!) {
        dependencyView(courseCode: $courseCode) {
          courseCode
          directPrerequisites
          message
        }
      }
    `,
    { courseCode },
  );

  return data.dependencyView;
}

export async function createTerm(season: string, year: number): Promise<Planner> {
  const data = await graphqlRequest<{ createTerm: Planner }>(
    `
      mutation CreateTerm($input: CreateTermInput!) {
        createTerm(input: $input) {
          ${plannerFields}
        }
      }
    `,
    { input: { userID: MOCK_USER_ID, season, year } },
  );

  return data.createTerm;
}

export async function addCourseToTerm(termID: string, courseCode: string): Promise<Planner> {
  const data = await graphqlRequest<{ addCourseToTerm: Planner }>(
    `
      mutation AddCourseToTerm($input: AddCourseToTermInput!) {
        addCourseToTerm(input: $input) {
          ${plannerFields}
        }
      }
    `,
    { input: { userID: MOCK_USER_ID, termID, courseCode } },
  );

  return data.addCourseToTerm;
}

export async function removeRoadmapCourse(entryID: string): Promise<Planner> {
  const data = await graphqlRequest<{ removeRoadmapCourse: Planner }>(
    `
      mutation RemoveRoadmapCourse($input: RemoveRoadmapCourseInput!) {
        removeRoadmapCourse(input: $input) {
          ${plannerFields}
        }
      }
    `,
    { input: { userID: MOCK_USER_ID, entryID } },
  );

  return data.removeRoadmapCourse;
}

export async function moveRoadmapCourse(entryID: string, targetTermID: string): Promise<Planner> {
  const data = await graphqlRequest<{ moveRoadmapCourse: Planner }>(
    `
      mutation MoveRoadmapCourse($input: MoveRoadmapCourseInput!) {
        moveRoadmapCourse(input: $input) {
          ${plannerFields}
        }
      }
    `,
    { input: { userID: MOCK_USER_ID, entryID, targetTermID } },
  );

  return data.moveRoadmapCourse;
}

export async function updateRoadmapCourseStatus(
  entryID: string,
  status: CourseStatus,
): Promise<Planner> {
  const data = await graphqlRequest<{ updateRoadmapCourseStatus: Planner }>(
    `
      mutation UpdateRoadmapCourseStatus($input: UpdateRoadmapCourseStatusInput!) {
        updateRoadmapCourseStatus(input: $input) {
          ${plannerFields}
        }
      }
    `,
    { input: { userID: MOCK_USER_ID, entryID, status } },
  );

  return data.updateRoadmapCourseStatus;
}
