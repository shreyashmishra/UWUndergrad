package graph

import (
	"context"

	"planahead/planner-api/internal/model"
	"planahead/planner-api/internal/service"
)

type RootResolver struct {
	programService *service.ProgramService
	studentService *service.StudentService
	authService    *service.AuthService
}

func NewRootResolver(programService *service.ProgramService, studentService *service.StudentService, authService *service.AuthService) *RootResolver {
	return &RootResolver{
		programService: programService,
		studentService: studentService,
		authService:    authService,
	}
}

func (r *RootResolver) AvailableUniversities(ctx context.Context) ([]*universityResolver, error) {
	items, err := r.programService.ListUniversities(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*universityResolver, 0, len(items))
	for _, item := range items {
		result = append(result, &universityResolver{item: item})
	}
	return result, nil
}

func (r *RootResolver) ProgramsByUniversity(ctx context.Context, args struct{ UniversityCode string }) ([]*programResolver, error) {
	items, err := r.programService.ListProgramsByUniversity(ctx, args.UniversityCode)
	if err != nil {
		return nil, err
	}

	result := make([]*programResolver, 0, len(items))
	for _, item := range items {
		result = append(result, &programResolver{item: item})
	}
	return result, nil
}

func (r *RootResolver) RoadmapByProgram(ctx context.Context, args struct {
	UniversityCode string
	ProgramCode    string
	ProgressInput  *progressInputResolverArgs
}) (*roadmapResolver, error) {
	roadmap, err := r.programService.RoadmapByProgram(ctx, args.UniversityCode, args.ProgramCode, args.ProgressInput.toSnapshot())
	if err != nil {
		return nil, err
	}
	return &roadmapResolver{item: *roadmap}, nil
}

func (r *RootResolver) StudentProgress(ctx context.Context, args struct {
	UniversityCode     string
	ProgramCode        string
	StudentExternalKey *string
}) (*studentProgressResolver, error) {
	progress, err := r.studentService.StudentProgress(ctx, args.UniversityCode, args.ProgramCode, args.StudentExternalKey)
	if err != nil {
		return nil, err
	}
	return &studentProgressResolver{item: *progress}, nil
}

func (r *RootResolver) RequirementSummary(ctx context.Context, args struct {
	UniversityCode string
	ProgramCode    string
	ProgressInput  *progressInputResolverArgs
}) (*requirementSummaryResolver, error) {
	summary, err := r.programService.RequirementSummary(ctx, args.UniversityCode, args.ProgramCode, args.ProgressInput.toSnapshot())
	if err != nil {
		return nil, err
	}
	return &requirementSummaryResolver{item: *summary}, nil
}

func (r *RootResolver) UpdateCourseStatus(ctx context.Context, args struct {
	Input struct {
		UniversityCode     string
		ProgramCode        string
		CourseCode         string
		Status             string
		StudentExternalKey *string
	}
}) (*studentProgressResolver, error) {
	progress, err := r.studentService.UpdateCourseStatus(ctx, args.Input.UniversityCode, args.Input.ProgramCode, args.Input.CourseCode, model.CourseStatus(args.Input.Status), args.Input.StudentExternalKey)
	if err != nil {
		return nil, err
	}
	return &studentProgressResolver{item: *progress}, nil
}

func (r *RootResolver) SelectElective(ctx context.Context, args struct {
	Input struct {
		UniversityCode     string
		ProgramCode        string
		GroupCode          string
		CourseCode         string
		StudentExternalKey *string
	}
}) (*studentProgressResolver, error) {
	progress, err := r.studentService.SelectElective(ctx, args.Input.UniversityCode, args.Input.ProgramCode, args.Input.GroupCode, args.Input.CourseCode, args.Input.StudentExternalKey)
	if err != nil {
		return nil, err
	}
	return &studentProgressResolver{item: *progress}, nil
}

func (r *RootResolver) ClearElectiveSelection(ctx context.Context, args struct {
	Input struct {
		UniversityCode     string
		ProgramCode        string
		GroupCode          string
		StudentExternalKey *string
	}
}) (*studentProgressResolver, error) {
	progress, err := r.studentService.ClearElectiveSelection(ctx, args.Input.UniversityCode, args.Input.ProgramCode, args.Input.GroupCode, args.Input.StudentExternalKey)
	if err != nil {
		return nil, err
	}
	return &studentProgressResolver{item: *progress}, nil
}

func (r *RootResolver) Register(ctx context.Context, args struct {
	Email    string
	FullName string
	Password string
}) (*authPayloadResolver, error) {
	payload, err := r.authService.Register(ctx, args.Email, args.FullName, args.Password)
	if err != nil {
		return nil, err
	}
	return &authPayloadResolver{item: payload}, nil
}

func (r *RootResolver) Login(ctx context.Context, args struct {
	Email    string
	Password string
}) (*authPayloadResolver, error) {
	payload, err := r.authService.Login(ctx, args.Email, args.Password)
	if err != nil {
		return nil, err
	}
	return &authPayloadResolver{item: payload}, nil
}

type authPayloadResolver struct {
	item *service.AuthPayload
}

func (r *authPayloadResolver) Token() string       { return r.item.Token }
func (r *authPayloadResolver) StudentName() string { return r.item.StudentName }
func (r *authPayloadResolver) ExternalKey() string { return r.item.ExternalKey }

type progressInputResolverArgs struct {
	CourseStatuses     *[]struct {
		CourseCode string
		Status     string
	}
	ElectiveSelections *[]struct {
		GroupCode  string
		CourseCode string
	}
}

func (p *progressInputResolverArgs) toSnapshot() *model.ProgressSnapshot {
	if p == nil {
		return nil
	}

	snapshot := &model.ProgressSnapshot{
		CourseStatuses:     map[string]model.CourseStatus{},
		ElectiveSelections: map[string]string{},
	}

	if p.CourseStatuses != nil {
		for _, item := range *p.CourseStatuses {
			snapshot.CourseStatuses[item.CourseCode] = model.CourseStatus(item.Status)
		}
	}

	if p.ElectiveSelections != nil {
		for _, item := range *p.ElectiveSelections {
			snapshot.ElectiveSelections[item.GroupCode] = item.CourseCode
		}
	}

	return snapshot
}

type universityResolver struct{ item model.University }

func (r *universityResolver) Code() string { return r.item.Code }
func (r *universityResolver) Name() string { return r.item.Name }

type programResolver struct{ item model.ProgramSummary }

func (r *programResolver) Code() string           { return r.item.Code }
func (r *programResolver) Name() string           { return r.item.Name }
func (r *programResolver) Degree() string         { return r.item.Degree }
func (r *programResolver) Description() string    { return r.item.Description }
func (r *programResolver) UniversityCode() string { return r.item.UniversityCode }

type roadmapResolver struct{ item model.EvaluatedRoadmap }

func (r *roadmapResolver) UniversityCode() string             { return r.item.UniversityCode }
func (r *roadmapResolver) ProgramCode() string                { return r.item.ProgramCode }
func (r *roadmapResolver) ProgramName() string                { return r.item.ProgramName }
func (r *roadmapResolver) Degree() string                     { return r.item.Degree }
func (r *roadmapResolver) Description() string                { return r.item.Description }
func (r *roadmapResolver) Summary() *requirementSummaryResolver { return &requirementSummaryResolver{item: r.item.Summary} }
func (r *roadmapResolver) Terms() []*termRoadmapResolver {
	result := make([]*termRoadmapResolver, 0, len(r.item.Terms))
	for _, term := range r.item.Terms {
		result = append(result, &termRoadmapResolver{item: term})
	}
	return result
}

type requirementSummaryResolver struct{ item model.RequirementSummary }

func (r *requirementSummaryResolver) TotalRequirements() int32      { return r.item.TotalRequirements }
func (r *requirementSummaryResolver) CompletedRequirements() int32  { return r.item.CompletedRequirements }
func (r *requirementSummaryResolver) InProgressRequirements() int32 { return r.item.InProgressRequirements }
func (r *requirementSummaryResolver) PlannedRequirements() int32    { return r.item.PlannedRequirements }
func (r *requirementSummaryResolver) RemainingRequirements() int32  { return r.item.RemainingRequirements }
func (r *requirementSummaryResolver) SelectedElectives() int32      { return r.item.SelectedElectives }
func (r *requirementSummaryResolver) CompletionPercentage() float64 { return r.item.CompletionPercentage }

type termRoadmapResolver struct{ item model.EvaluatedTerm }

func (r *termRoadmapResolver) Code() string           { return r.item.Code }
func (r *termRoadmapResolver) Label() string          { return r.item.Label }
func (r *termRoadmapResolver) Year() int32            { return r.item.Year }
func (r *termRoadmapResolver) Season() string         { return string(r.item.Season) }
func (r *termRoadmapResolver) Sequence() int32        { return r.item.Sequence }
func (r *termRoadmapResolver) CompletedCount() int32  { return r.item.CompletedCount }
func (r *termRoadmapResolver) TotalCount() int32      { return r.item.TotalCount }
func (r *termRoadmapResolver) Requirements() []*termRequirementResolver {
	result := make([]*termRequirementResolver, 0, len(r.item.Requirements))
	for _, requirement := range r.item.Requirements {
		result = append(result, &termRequirementResolver{item: requirement})
	}
	return result
}

type termRequirementResolver struct{ item model.EvaluatedTermRequirement }

func (r *termRequirementResolver) Kind() string      { return string(r.item.Kind) }
func (r *termRequirementResolver) Sequence() int32   { return r.item.Sequence }
func (r *termRequirementResolver) Course() *roadmapCourseResolver {
	if r.item.Course == nil {
		return nil
	}
	return &roadmapCourseResolver{item: *r.item.Course}
}
func (r *termRequirementResolver) Group() *electiveGroupResolver {
	if r.item.Group == nil {
		return nil
	}
	return &electiveGroupResolver{item: *r.item.Group}
}

type roadmapCourseResolver struct{ item model.EvaluatedCourse }

func (r *roadmapCourseResolver) Code() string                 { return r.item.Code }
func (r *roadmapCourseResolver) Title() string                { return r.item.Title }
func (r *roadmapCourseResolver) Credits() float64             { return r.item.Credits }
func (r *roadmapCourseResolver) Description() *string         { return r.item.Description }
func (r *roadmapCourseResolver) Subject() *string             { return r.item.Subject }
func (r *roadmapCourseResolver) Status() string               { return string(r.item.Status) }
func (r *roadmapCourseResolver) PrerequisitesMet() bool       { return r.item.PrerequisitesMet }
func (r *roadmapCourseResolver) PrerequisiteMessage() *string { return r.item.PrerequisiteMessage }
func (r *roadmapCourseResolver) Notes() *string               { return r.item.Notes }
func (r *roadmapCourseResolver) IsSelected() bool             { return r.item.IsSelected }

type electiveGroupResolver struct{ item model.EvaluatedGroup }

func (r *electiveGroupResolver) Code() string               { return r.item.Code }
func (r *electiveGroupResolver) Title() string              { return r.item.Title }
func (r *electiveGroupResolver) Description() string        { return r.item.Description }
func (r *electiveGroupResolver) Kind() string               { return string(r.item.Kind) }
func (r *electiveGroupResolver) MinSelections() int32       { return r.item.MinSelections }
func (r *electiveGroupResolver) MaxSelections() int32       { return r.item.MaxSelections }
func (r *electiveGroupResolver) SelectedCourseCode() *string { return r.item.SelectedCourseCode }
func (r *electiveGroupResolver) Status() string             { return string(r.item.Status) }
func (r *electiveGroupResolver) IsSatisfied() bool          { return r.item.IsSatisfied }
func (r *electiveGroupResolver) Notes() *string             { return r.item.Notes }
func (r *electiveGroupResolver) Options() []*roadmapCourseResolver {
	result := make([]*roadmapCourseResolver, 0, len(r.item.Options))
	for _, option := range r.item.Options {
		result = append(result, &roadmapCourseResolver{item: option})
	}
	return result
}

type studentProgressResolver struct{ item model.StudentProgress }

func (r *studentProgressResolver) StudentExternalKey() string { return r.item.StudentExternalKey }
func (r *studentProgressResolver) CourseStatuses() []*courseStatusEntryResolver {
	result := make([]*courseStatusEntryResolver, 0, len(r.item.CourseStatuses))
	for _, item := range r.item.CourseStatuses {
		result = append(result, &courseStatusEntryResolver{item: item})
	}
	return result
}
func (r *studentProgressResolver) ElectiveSelections() []*electiveSelectionEntryResolver {
	result := make([]*electiveSelectionEntryResolver, 0, len(r.item.ElectiveSelections))
	for _, item := range r.item.ElectiveSelections {
		result = append(result, &electiveSelectionEntryResolver{item: item})
	}
	return result
}

type courseStatusEntryResolver struct{ item model.CourseStatusRecord }

func (r *courseStatusEntryResolver) CourseCode() string { return r.item.CourseCode }
func (r *courseStatusEntryResolver) Status() string     { return string(r.item.Status) }

type electiveSelectionEntryResolver struct{ item model.ElectiveSelectionRecord }

func (r *electiveSelectionEntryResolver) GroupCode() string  { return r.item.GroupCode }
func (r *electiveSelectionEntryResolver) CourseCode() string { return r.item.CourseCode }
