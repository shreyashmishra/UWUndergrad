package model

type CourseStatus string

const (
	CourseStatusNotStarted CourseStatus = "NOT_STARTED"
	CourseStatusPlanned    CourseStatus = "PLANNED"
	CourseStatusInProgress CourseStatus = "IN_PROGRESS"
	CourseStatusCompleted  CourseStatus = "COMPLETED"
)

type RequirementKind string

const (
	RequirementKindCourse        RequirementKind = "COURSE"
	RequirementKindElectiveGroup RequirementKind = "ELECTIVE_GROUP"
)

type RequirementGroupKind string

const (
	RequirementGroupKindCore     RequirementGroupKind = "CORE"
	RequirementGroupKindElective RequirementGroupKind = "ELECTIVE"
	RequirementGroupKindOneOf    RequirementGroupKind = "ONE_OF"
)

type TermSeason string

const (
	TermSeasonFall   TermSeason = "FALL"
	TermSeasonWinter TermSeason = "WINTER"
	TermSeasonSpring TermSeason = "SPRING"
)

type University struct {
	Code string
	Name string
}

type ProgramSummary struct {
	Code           string
	Name           string
	Degree         string
	Description    string
	UniversityCode string
}

type CourseDefinition struct {
	Code        string
	Title       string
	Credits     float64
	Description *string
	Subject     *string
}

type PrerequisiteDefinition struct {
	CourseCode    string
	MinimumGrade  *string
	IsCorequisite bool
}

type CourseRequirementDefinition struct {
	Code          string
	Course        CourseDefinition
	Sequence      int32
	Notes         *string
	Prerequisites []PrerequisiteDefinition
	Kind          RequirementKind
}

type ElectiveGroupDefinition struct {
	Code               string
	Title              string
	Description        string
	Kind               RequirementGroupKind
	MinSelections      int32
	MaxSelections      int32
	Sequence           int32
	Options            []CourseRequirementDefinition
	Notes              *string
}

type TermRequirementDefinition struct {
	Kind     RequirementKind
	Sequence int32
	Course   *CourseRequirementDefinition
	Group    *ElectiveGroupDefinition
}

type TermDefinition struct {
	Code         string
	Label        string
	Year         int32
	Season       TermSeason
	Sequence     int32
	Requirements []TermRequirementDefinition
}

type ProgramDefinition struct {
	UniversityCode string
	ProgramCode    string
	Name           string
	Degree         string
	Description    string
	Terms          []TermDefinition
}

type ProgressSnapshot struct {
	CourseStatuses     map[string]CourseStatus
	ElectiveSelections map[string]string
}

type EvaluatedCourse struct {
	Code                string
	Title               string
	Credits             float64
	Description         *string
	Subject             *string
	Status              CourseStatus
	PrerequisitesMet    bool
	PrerequisiteMessage *string
	Notes               *string
	IsSelected          bool
}

type EvaluatedGroup struct {
	Code               string
	Title              string
	Description        string
	Kind               RequirementGroupKind
	MinSelections      int32
	MaxSelections      int32
	SelectedCourseCode *string
	Status             CourseStatus
	IsSatisfied        bool
	Notes              *string
	Options            []EvaluatedCourse
}

type EvaluatedTermRequirement struct {
	Kind     RequirementKind
	Sequence int32
	Course   *EvaluatedCourse
	Group    *EvaluatedGroup
}

type RequirementSummary struct {
	TotalRequirements      int32
	CompletedRequirements  int32
	InProgressRequirements int32
	PlannedRequirements    int32
	RemainingRequirements  int32
	SelectedElectives      int32
	CompletionPercentage   float64
}

type EvaluatedTerm struct {
	Code           string
	Label          string
	Year           int32
	Season         TermSeason
	Sequence       int32
	CompletedCount int32
	TotalCount     int32
	Requirements   []EvaluatedTermRequirement
}

type EvaluatedRoadmap struct {
	UniversityCode string
	ProgramCode    string
	ProgramName    string
	Degree         string
	Description    string
	Summary        RequirementSummary
	Terms          []EvaluatedTerm
}

type StudentProgress struct {
	StudentExternalKey string
	CourseStatuses     []CourseStatusRecord
	ElectiveSelections []ElectiveSelectionRecord
}

type CourseStatusRecord struct {
	CourseCode string
	Status     CourseStatus
}

type ElectiveSelectionRecord struct {
	GroupCode  string
	CourseCode string
}
