package graph

const Schema = `
schema {
	query: Query
	mutation: Mutation
}

type Query {
	availableUniversities: [University!]!
	programsByUniversity(universityCode: String!): [Program!]!
	roadmapByProgram(universityCode: String!, programCode: String!, progressInput: ProgressInput): Roadmap!
	studentProgress(universityCode: String!, programCode: String!, studentExternalKey: String): StudentProgress!
	requirementSummary(universityCode: String!, programCode: String!, progressInput: ProgressInput): RequirementSummary!
}

type Mutation {
	updateCourseStatus(input: UpdateCourseStatusInput!): StudentProgress!
	selectElective(input: SelectElectiveInput!): StudentProgress!
	clearElectiveSelection(input: ClearElectiveSelectionInput!): StudentProgress!

	register(email: String!, fullName: String!, password: String!): AuthPayload!
	login(email: String!, password: String!): AuthPayload!
}

type AuthPayload {
	token: String!
	studentName: String!
	externalKey: String!
}

type University {
	code: String!
	name: String!
}

type Program {
	code: String!
	name: String!
	degree: String!
	description: String!
	universityCode: String!
}

type Roadmap {
	universityCode: String!
	programCode: String!
	programName: String!
	degree: String!
	description: String!
	summary: RequirementSummary!
	terms: [TermRoadmap!]!
}

type RequirementSummary {
	totalRequirements: Int!
	completedRequirements: Int!
	inProgressRequirements: Int!
	plannedRequirements: Int!
	remainingRequirements: Int!
	selectedElectives: Int!
	completionPercentage: Float!
}

type TermRoadmap {
	code: String!
	label: String!
	year: Int!
	season: String!
	sequence: Int!
	completedCount: Int!
	totalCount: Int!
	requirements: [TermRequirement!]!
}

type TermRequirement {
	kind: String!
	sequence: Int!
	course: RoadmapCourse
	group: ElectiveGroup
}

type RoadmapCourse {
	code: String!
	title: String!
	credits: Float!
	description: String
	subject: String
	status: String!
	prerequisitesMet: Boolean!
	prerequisiteMessage: String
	notes: String
	isSelected: Boolean!
}

type ElectiveGroup {
	code: String!
	title: String!
	description: String!
	kind: String!
	minSelections: Int!
	maxSelections: Int!
	selectedCourseCode: String
	status: String!
	isSatisfied: Boolean!
	notes: String
	options: [RoadmapCourse!]!
}

type StudentProgress {
	studentExternalKey: String!
	courseStatuses: [CourseStatusEntry!]!
	electiveSelections: [ElectiveSelectionEntry!]!
}

type CourseStatusEntry {
	courseCode: String!
	status: String!
}

type ElectiveSelectionEntry {
	groupCode: String!
	courseCode: String!
}

input CourseStatusInput {
	courseCode: String!
	status: String!
}

input ElectiveSelectionInput {
	groupCode: String!
	courseCode: String!
}

input ProgressInput {
	courseStatuses: [CourseStatusInput!]
	electiveSelections: [ElectiveSelectionInput!]
}

input UpdateCourseStatusInput {
	universityCode: String!
	programCode: String!
	courseCode: String!
	status: String!
	studentExternalKey: String
}

input SelectElectiveInput {
	universityCode: String!
	programCode: String!
	groupCode: String!
	courseCode: String!
	studentExternalKey: String
}

input ClearElectiveSelectionInput {
	universityCode: String!
	programCode: String!
	groupCode: String!
	studentExternalKey: String
}
`
