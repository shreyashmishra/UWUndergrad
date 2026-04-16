package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"planahead/planner-api/internal/model"
)

type CatalogRepository struct {
	db *sql.DB
}

func NewCatalogRepository(db *sql.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListUniversities(ctx context.Context) ([]model.University, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code, name FROM universities ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.University, 0)
	for rows.Next() {
		var item model.University
		if err := rows.Scan(&item.Code, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *CatalogRepository) ListProgramsByUniversity(ctx context.Context, universityCode string) ([]model.ProgramSummary, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT p.code, p.name, p.degree, p.description, u.code
		FROM programs p
		INNER JOIN universities u ON u.id = p.university_id
		WHERE u.code = ?
		ORDER BY p.name ASC`,
		universityCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.ProgramSummary, 0)
	for rows.Next() {
		var item model.ProgramSummary
		if err := rows.Scan(&item.Code, &item.Name, &item.Degree, &item.Description, &item.UniversityCode); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *CatalogRepository) GetProgramDefinition(ctx context.Context, universityCode string, programCode string) (*model.ProgramDefinition, error) {
	type programRow struct {
		ProgramID      int64
		TemplateID     int64
		UniversityCode string
		ProgramCode    string
		Name           string
		Degree         string
		Description    string
	}

	var program programRow
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT p.id, ppt.id, u.code, p.code, p.name, p.degree, p.description
		FROM programs p
		INNER JOIN universities u ON u.id = p.university_id
		INNER JOIN program_plan_templates ppt ON ppt.program_id = p.id
		WHERE u.code = ? AND p.code = ?`,
		universityCode,
		programCode,
	).Scan(&program.ProgramID, &program.TemplateID, &program.UniversityCode, &program.ProgramCode, &program.Name, &program.Degree, &program.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("program %s/%s not found", universityCode, programCode)
		}
		return nil, err
	}

	terms, termIDByCode, err := r.loadTerms(ctx, program.TemplateID)
	if err != nil {
		return nil, err
	}

	requirementsByTerm, err := r.loadCourseRequirements(ctx, program.ProgramID)
	if err != nil {
		return nil, err
	}

	groupsByTerm, err := r.loadGroupRequirements(ctx, program.ProgramID, termIDByCode)
	if err != nil {
		return nil, err
	}

	for termIndex := range terms {
		termCode := terms[termIndex].Code
		termRequirements := append(requirementsByTerm[termCode], groupsByTerm[termCode]...)
		sort.SliceStable(termRequirements, func(i int, j int) bool {
			return termRequirements[i].Sequence < termRequirements[j].Sequence
		})
		terms[termIndex].Requirements = termRequirements
	}

	return &model.ProgramDefinition{
		UniversityCode: program.UniversityCode,
		ProgramCode:    program.ProgramCode,
		Name:           program.Name,
		Degree:         program.Degree,
		Description:    program.Description,
		Terms:          terms,
	}, nil
}

func (r *CatalogRepository) loadTerms(ctx context.Context, templateID int64) ([]model.TermDefinition, map[int64]string, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, code, label, year, season, sequence
		FROM terms
		WHERE program_plan_template_id = ?
		ORDER BY sequence ASC`,
		templateID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	terms := make([]model.TermDefinition, 0)
	termIDByCode := map[int64]string{}
	for rows.Next() {
		var termID int64
		var term model.TermDefinition
		if err := rows.Scan(&termID, &term.Code, &term.Label, &term.Year, &term.Season, &term.Sequence); err != nil {
			return nil, nil, err
		}
		terms = append(terms, term)
		termIDByCode[termID] = term.Code
	}

	return terms, termIDByCode, rows.Err()
}

func (r *CatalogRepository) loadCourseRequirements(ctx context.Context, programID int64) (map[string][]model.TermRequirementDefinition, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT t.code, pr.sequence, pr.display_title, pr.notes, c.code, c.title, c.credits, c.description, c.subject
		FROM program_requirements pr
		INNER JOIN terms t ON t.id = pr.term_id
		INNER JOIN courses c ON c.id = pr.course_id
		WHERE pr.program_id = ? AND pr.requirement_group_id IS NULL
		ORDER BY t.sequence ASC, pr.sequence ASC`,
		programID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prerequisitesByCourse, err := r.loadPrerequisites(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string][]model.TermRequirementDefinition{}
	for rows.Next() {
		var termCode string
		var sequence int32
		var displayTitle sql.NullString
		var notes sql.NullString
		var courseCode string
		var title string
		var credits float64
		var description sql.NullString
		var subject sql.NullString

		if err := rows.Scan(&termCode, &sequence, &displayTitle, &notes, &courseCode, &title, &credits, &description, &subject); err != nil {
			return nil, err
		}

		courseReq := model.CourseRequirementDefinition{
			Code:          nullStringOr(displayTitle, courseCode),
			Sequence:      sequence,
			Notes:         nullableString(notes),
			Prerequisites: prerequisitesByCourse[courseCode],
			Kind:          model.RequirementKindCourse,
			Course: model.CourseDefinition{
				Code:        courseCode,
				Title:       title,
				Credits:     credits,
				Description: nullableString(description),
				Subject:     nullableString(subject),
			},
		}
		result[termCode] = append(result[termCode], model.TermRequirementDefinition{
			Kind:     model.RequirementKindCourse,
			Sequence: sequence,
			Course:   &courseReq,
		})
	}

	return result, rows.Err()
}

func (r *CatalogRepository) loadGroupRequirements(ctx context.Context, programID int64, termIDByCode map[int64]string) (map[string][]model.TermRequirementDefinition, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT rg.id, rg.term_id, rg.code, rg.title, rg.description, rg.kind, rg.sequence, rg.min_selections, rg.max_selections, eg.selection_label
		FROM requirement_groups rg
		INNER JOIN elective_groups eg ON eg.requirement_group_id = rg.id
		WHERE rg.program_id = ?
		ORDER BY rg.sequence ASC`,
		programID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prerequisitesByCourse, err := r.loadPrerequisites(ctx)
	if err != nil {
		return nil, err
	}

	result := map[string][]model.TermRequirementDefinition{}
	for rows.Next() {
		var groupID int64
		var termID int64
		var group model.ElectiveGroupDefinition
		var selectionLabel sql.NullString
		if err := rows.Scan(&groupID, &termID, &group.Code, &group.Title, &group.Description, &group.Kind, &group.Sequence, &group.MinSelections, &group.MaxSelections, &selectionLabel); err != nil {
			return nil, err
		}
		group.Notes = nullableString(selectionLabel)

		options, err := r.loadGroupOptions(ctx, groupID, prerequisitesByCourse)
		if err != nil {
			return nil, err
		}
		group.Options = options

		termCode := termIDByCode[termID]
		result[termCode] = append(result[termCode], model.TermRequirementDefinition{
			Kind:     model.RequirementKindElectiveGroup,
			Sequence: group.Sequence,
			Group:    &group,
		})
	}

	return result, rows.Err()
}

func (r *CatalogRepository) loadGroupOptions(ctx context.Context, groupID int64, prerequisitesByCourse map[string][]model.PrerequisiteDefinition) ([]model.CourseRequirementDefinition, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT pr.sequence, pr.display_title, pr.notes, c.code, c.title, c.credits, c.description, c.subject
		FROM program_requirements pr
		INNER JOIN courses c ON c.id = pr.course_id
		WHERE pr.requirement_group_id = ?
		ORDER BY pr.sequence ASC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.CourseRequirementDefinition, 0)
	for rows.Next() {
		var option model.CourseRequirementDefinition
		var displayTitle sql.NullString
		var notes sql.NullString
		var description sql.NullString
		var subject sql.NullString
		if err := rows.Scan(&option.Sequence, &displayTitle, &notes, &option.Course.Code, &option.Course.Title, &option.Course.Credits, &description, &subject); err != nil {
			return nil, err
		}
		option.Code = nullStringOr(displayTitle, option.Course.Code)
		option.Notes = nullableString(notes)
		option.Kind = model.RequirementKindElectiveGroup
		option.Course.Description = nullableString(description)
		option.Course.Subject = nullableString(subject)
		option.Prerequisites = prerequisitesByCourse[option.Course.Code]
		result = append(result, option)
	}

	return result, rows.Err()
}

func (r *CatalogRepository) loadPrerequisites(ctx context.Context) (map[string][]model.PrerequisiteDefinition, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT c.code, p.code, pr.minimum_grade, pr.is_corequisite
		FROM prerequisite_rules pr
		INNER JOIN courses c ON c.id = pr.course_id
		INNER JOIN courses p ON p.id = pr.prerequisite_course_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string][]model.PrerequisiteDefinition{}
	for rows.Next() {
		var courseCode string
		var prerequisite model.PrerequisiteDefinition
		var minimumGrade sql.NullString
		if err := rows.Scan(&courseCode, &prerequisite.CourseCode, &minimumGrade, &prerequisite.IsCorequisite); err != nil {
			return nil, err
		}
		prerequisite.MinimumGrade = nullableString(minimumGrade)
		result[courseCode] = append(result[courseCode], prerequisite)
	}

	return result, rows.Err()
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func nullStringOr(value sql.NullString, fallback string) string {
	if !value.Valid || value.String == "" {
		return fallback
	}
	return value.String
}
