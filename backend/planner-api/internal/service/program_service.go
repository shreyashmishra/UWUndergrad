package service

import (
	"context"
	"sort"

	"planahead/planner-api/internal/model"
	"planahead/planner-api/internal/repository"
)

type ProgramService struct {
	catalogRepo  *repository.CatalogRepository
	studentRepo  *repository.StudentRepository
	defaultStudentKey string
}

func NewProgramService(catalogRepo *repository.CatalogRepository, studentRepo *repository.StudentRepository, defaultStudentKey string) *ProgramService {
	return &ProgramService{
		catalogRepo:  catalogRepo,
		studentRepo:  studentRepo,
		defaultStudentKey: defaultStudentKey,
	}
}

func (s *ProgramService) ListUniversities(ctx context.Context) ([]model.University, error) {
	return s.catalogRepo.ListUniversities(ctx)
}

func (s *ProgramService) ListProgramsByUniversity(ctx context.Context, universityCode string) ([]model.ProgramSummary, error) {
	return s.catalogRepo.ListProgramsByUniversity(ctx, universityCode)
}

func (s *ProgramService) RoadmapByProgram(ctx context.Context, universityCode string, programCode string, progressInput *model.ProgressSnapshot) (*model.EvaluatedRoadmap, error) {
	program, err := s.catalogRepo.GetProgramDefinition(ctx, universityCode, programCode)
	if err != nil {
		return nil, err
	}

	progress := model.ProgressSnapshot{
		CourseStatuses:     map[string]model.CourseStatus{},
		ElectiveSelections: map[string]string{},
	}
	if progressInput != nil {
		progress = *progressInput
	} else {
		progress, err = s.studentRepo.GetProgressSnapshot(ctx, universityCode, programCode, s.defaultStudentKey)
		if err != nil {
			return nil, err
		}
	}

	roadmap := evaluateProgram(*program, progress)
	return &roadmap, nil
}

func (s *ProgramService) RequirementSummary(ctx context.Context, universityCode string, programCode string, progressInput *model.ProgressSnapshot) (*model.RequirementSummary, error) {
	roadmap, err := s.RoadmapByProgram(ctx, universityCode, programCode, progressInput)
	if err != nil {
		return nil, err
	}
	return &roadmap.Summary, nil
}

type StudentService struct {
	repo              *repository.StudentRepository
	defaultStudentKey string
}

func NewStudentService(repo *repository.StudentRepository, defaultStudentKey string) *StudentService {
	return &StudentService{
		repo:              repo,
		defaultStudentKey: defaultStudentKey,
	}
}

func (s *StudentService) StudentProgress(ctx context.Context, universityCode string, programCode string, studentExternalKey *string) (*model.StudentProgress, error) {
	key := s.resolveStudentKey(studentExternalKey)
	snapshot, err := s.repo.GetProgressSnapshot(ctx, universityCode, programCode, key)
	if err != nil {
		return nil, err
	}
	return toStudentProgress(snapshot, key), nil
}

func (s *StudentService) UpdateCourseStatus(ctx context.Context, universityCode string, programCode string, courseCode string, status model.CourseStatus, studentExternalKey *string) (*model.StudentProgress, error) {
	key := s.resolveStudentKey(studentExternalKey)
	snapshot, err := s.repo.UpdateCourseStatus(ctx, universityCode, programCode, courseCode, status, key)
	if err != nil {
		return nil, err
	}
	return toStudentProgress(snapshot, key), nil
}

func (s *StudentService) SelectElective(ctx context.Context, universityCode string, programCode string, groupCode string, courseCode string, studentExternalKey *string) (*model.StudentProgress, error) {
	key := s.resolveStudentKey(studentExternalKey)
	snapshot, err := s.repo.SelectElective(ctx, universityCode, programCode, groupCode, courseCode, key)
	if err != nil {
		return nil, err
	}
	return toStudentProgress(snapshot, key), nil
}

func (s *StudentService) ClearElectiveSelection(ctx context.Context, universityCode string, programCode string, groupCode string, studentExternalKey *string) (*model.StudentProgress, error) {
	key := s.resolveStudentKey(studentExternalKey)
	snapshot, err := s.repo.ClearElectiveSelection(ctx, universityCode, programCode, groupCode, key)
	if err != nil {
		return nil, err
	}
	return toStudentProgress(snapshot, key), nil
}

func (s *StudentService) resolveStudentKey(studentExternalKey *string) string {
	if studentExternalKey == nil || *studentExternalKey == "" {
		return s.defaultStudentKey
	}
	return *studentExternalKey
}

func toStudentProgress(snapshot model.ProgressSnapshot, studentExternalKey string) *model.StudentProgress {
	progress := &model.StudentProgress{
		StudentExternalKey: studentExternalKey,
		CourseStatuses:     make([]model.CourseStatusRecord, 0, len(snapshot.CourseStatuses)),
		ElectiveSelections: make([]model.ElectiveSelectionRecord, 0, len(snapshot.ElectiveSelections)),
	}

	for courseCode, status := range snapshot.CourseStatuses {
		progress.CourseStatuses = append(progress.CourseStatuses, model.CourseStatusRecord{
			CourseCode: courseCode,
			Status:     status,
		})
	}
	sort.Slice(progress.CourseStatuses, func(i int, j int) bool {
		return progress.CourseStatuses[i].CourseCode < progress.CourseStatuses[j].CourseCode
	})

	for groupCode, courseCode := range snapshot.ElectiveSelections {
		progress.ElectiveSelections = append(progress.ElectiveSelections, model.ElectiveSelectionRecord{
			GroupCode:  groupCode,
			CourseCode: courseCode,
		})
	}
	sort.Slice(progress.ElectiveSelections, func(i int, j int) bool {
		return progress.ElectiveSelections[i].GroupCode < progress.ElectiveSelections[j].GroupCode
	})

	return progress
}

func evaluateProgram(program model.ProgramDefinition, progress model.ProgressSnapshot) model.EvaluatedRoadmap {
	evaluatedTerms := make([]model.EvaluatedTerm, 0, len(program.Terms))
	requirementStatuses := make([]model.CourseStatus, 0)

	for _, term := range program.Terms {
		evaluatedRequirements := make([]model.EvaluatedTermRequirement, 0, len(term.Requirements))
		termStatuses := make([]model.CourseStatus, 0, len(term.Requirements))

		for _, requirement := range term.Requirements {
			if requirement.Course != nil {
				evaluatedCourse := evaluateCourse(*requirement.Course, progress, true)
				evaluatedRequirements = append(evaluatedRequirements, model.EvaluatedTermRequirement{
					Kind:     model.RequirementKindCourse,
					Sequence: requirement.Sequence,
					Course:   &evaluatedCourse,
				})
				termStatuses = append(termStatuses, evaluatedCourse.Status)
				requirementStatuses = append(requirementStatuses, evaluatedCourse.Status)
				continue
			}

			if requirement.Group != nil {
				evaluatedGroup := evaluateGroup(*requirement.Group, progress)
				evaluatedRequirements = append(evaluatedRequirements, model.EvaluatedTermRequirement{
					Kind:     model.RequirementKindElectiveGroup,
					Sequence: requirement.Sequence,
					Group:    &evaluatedGroup,
				})
				termStatuses = append(termStatuses, evaluatedGroup.Status)
				requirementStatuses = append(requirementStatuses, evaluatedGroup.Status)
			}
		}

		completedCount := int32(0)
		for _, status := range termStatuses {
			if status == model.CourseStatusCompleted {
				completedCount++
			}
		}

		evaluatedTerms = append(evaluatedTerms, model.EvaluatedTerm{
			Code:           term.Code,
			Label:          term.Label,
			Year:           term.Year,
			Season:         term.Season,
			Sequence:       term.Sequence,
			CompletedCount: completedCount,
			TotalCount:     int32(len(termStatuses)),
			Requirements:   evaluatedRequirements,
		})
	}

	summary := summarizeRequirementStatuses(requirementStatuses, int32(len(progress.ElectiveSelections)))
	return model.EvaluatedRoadmap{
		UniversityCode: program.UniversityCode,
		ProgramCode:    program.ProgramCode,
		ProgramName:    program.Name,
		Degree:         program.Degree,
		Description:    program.Description,
		Summary:        summary,
		Terms:          evaluatedTerms,
	}
}

func evaluateCourse(requirement model.CourseRequirementDefinition, progress model.ProgressSnapshot, isSelected bool) model.EvaluatedCourse {
	status := progress.CourseStatuses[requirement.Course.Code]
	if status == "" {
		status = model.CourseStatusNotStarted
	}

	prerequisitesMet, prerequisiteMessage := evaluatePrerequisites(requirement.Prerequisites, progress.CourseStatuses)
	return model.EvaluatedCourse{
		Code:                requirement.Course.Code,
		Title:               requirement.Course.Title,
		Credits:             requirement.Course.Credits,
		Description:         requirement.Course.Description,
		Subject:             requirement.Course.Subject,
		Status:              status,
		PrerequisitesMet:    prerequisitesMet,
		PrerequisiteMessage: prerequisiteMessage,
		Notes:               requirement.Notes,
		IsSelected:          isSelected,
	}
}

func evaluateGroup(group model.ElectiveGroupDefinition, progress model.ProgressSnapshot) model.EvaluatedGroup {
	selectedCourseCode := progress.ElectiveSelections[group.Code]
	evaluatedOptions := make([]model.EvaluatedCourse, 0, len(group.Options))
	selectedStatus := model.CourseStatusNotStarted

	for _, option := range group.Options {
		isSelected := selectedCourseCode == option.Course.Code
		evaluatedOption := evaluateCourse(option, progress, isSelected)
		if isSelected {
			selectedStatus = evaluatedOption.Status
		}
		evaluatedOptions = append(evaluatedOptions, evaluatedOption)
	}

	var selectedCourseCodePtr *string
	if selectedCourseCode != "" {
		selectedCourseCodePtr = &selectedCourseCode
	}

	return model.EvaluatedGroup{
		Code:               group.Code,
		Title:              group.Title,
		Description:        group.Description,
		Kind:               group.Kind,
		MinSelections:      group.MinSelections,
		MaxSelections:      group.MaxSelections,
		SelectedCourseCode: selectedCourseCodePtr,
		Status:             selectedStatus,
		IsSatisfied:        selectedCourseCode != "" && selectedStatus != model.CourseStatusNotStarted,
		Notes:              group.Notes,
		Options:            evaluatedOptions,
	}
}

func evaluatePrerequisites(prerequisites []model.PrerequisiteDefinition, courseStatuses map[string]model.CourseStatus) (bool, *string) {
	if len(prerequisites) == 0 {
		return true, nil
	}

	unmet := make([]string, 0)
	for _, prerequisite := range prerequisites {
		status := courseStatuses[prerequisite.CourseCode]
		if status == "" {
			status = model.CourseStatusNotStarted
		}
		if prerequisite.IsCorequisite {
			if status == model.CourseStatusPlanned || status == model.CourseStatusInProgress || status == model.CourseStatusCompleted {
				continue
			}
		} else if status == model.CourseStatusInProgress || status == model.CourseStatusCompleted {
			continue
		}
		unmet = append(unmet, prerequisite.CourseCode)
	}

	if len(unmet) == 0 {
		message := "Prerequisites satisfied."
		return true, &message
	}

	if len(unmet) == 1 {
		message := "Needs " + unmet[0] + " before this course."
		return false, &message
	}

	message := "Needs "
	for index, courseCode := range unmet {
		if index > 0 {
			if index == len(unmet)-1 {
				message += ", and "
			} else {
				message += ", "
			}
		}
		message += courseCode
	}
	message += " before this course."
	return false, &message
}

func summarizeRequirementStatuses(statuses []model.CourseStatus, selectedElectives int32) model.RequirementSummary {
	summary := model.RequirementSummary{
		TotalRequirements: len32(statuses),
		SelectedElectives: selectedElectives,
	}

	for _, status := range statuses {
		switch status {
		case model.CourseStatusCompleted:
			summary.CompletedRequirements++
		case model.CourseStatusInProgress:
			summary.InProgressRequirements++
		case model.CourseStatusPlanned:
			summary.PlannedRequirements++
		}
	}

	summary.RemainingRequirements = summary.TotalRequirements - summary.CompletedRequirements - summary.InProgressRequirements - summary.PlannedRequirements
	if summary.TotalRequirements > 0 {
		summary.CompletionPercentage = float64(summary.CompletedRequirements) / float64(summary.TotalRequirements) * 100
	}
	return summary
}

func len32[T any](items []T) int32 {
	return int32(len(items))
}
