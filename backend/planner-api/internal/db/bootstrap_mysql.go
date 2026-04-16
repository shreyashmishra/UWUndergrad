package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"planahead/planner-api/internal/catalog"
	"planahead/planner-api/internal/model"
)

const mysqlSchema = `
CREATE TABLE IF NOT EXISTS students (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    external_key VARCHAR(64) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS universities (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS courses (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    university_id BIGINT NOT NULL,
    code VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    subject VARCHAR(64) NOT NULL,
    credits DOUBLE NOT NULL DEFAULT 0.5,
    description TEXT NULL,
    CONSTRAINT fk_courses_university FOREIGN KEY (university_id) REFERENCES universities(id) ON DELETE CASCADE,
    CONSTRAINT uq_course_university_code UNIQUE (university_id, code)
);

CREATE TABLE IF NOT EXISTS programs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    university_id BIGINT NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    degree VARCHAR(128) NOT NULL,
    description TEXT NOT NULL,
    CONSTRAINT fk_programs_university FOREIGN KEY (university_id) REFERENCES universities(id) ON DELETE CASCADE,
    CONSTRAINT uq_program_university_code UNIQUE (university_id, code)
);

CREATE TABLE IF NOT EXISTS program_plan_templates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    program_id BIGINT NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    version VARCHAR(32) NOT NULL,
    CONSTRAINT fk_program_plan_templates_program FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS terms (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    program_plan_template_id BIGINT NOT NULL,
    code VARCHAR(64) NOT NULL,
    label VARCHAR(255) NOT NULL,
    year INT NOT NULL,
    season ENUM('FALL', 'WINTER', 'SPRING') NOT NULL,
    sequence INT NOT NULL,
    CONSTRAINT fk_terms_template FOREIGN KEY (program_plan_template_id) REFERENCES program_plan_templates(id) ON DELETE CASCADE,
    CONSTRAINT uq_term_template_code UNIQUE (program_plan_template_id, code)
);

CREATE TABLE IF NOT EXISTS prerequisite_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    course_id BIGINT NOT NULL,
    prerequisite_course_id BIGINT NOT NULL,
    minimum_grade VARCHAR(16) NULL,
    is_corequisite BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_prerequisite_rules_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    CONSTRAINT fk_prerequisite_rules_prereq FOREIGN KEY (prerequisite_course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS requirement_groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    program_id BIGINT NOT NULL,
    term_id BIGINT NOT NULL,
    code VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(500) NOT NULL,
    kind ENUM('CORE', 'ELECTIVE', 'ONE_OF') NOT NULL,
    sequence INT NOT NULL,
    min_selections INT NOT NULL,
    max_selections INT NOT NULL,
    CONSTRAINT fk_requirement_groups_program FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE,
    CONSTRAINT fk_requirement_groups_term FOREIGN KEY (term_id) REFERENCES terms(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS elective_groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    requirement_group_id BIGINT NOT NULL UNIQUE,
    selection_label VARCHAR(255) NOT NULL,
    credits_required DOUBLE NOT NULL DEFAULT 0.5,
    CONSTRAINT fk_elective_groups_requirement_group FOREIGN KEY (requirement_group_id) REFERENCES requirement_groups(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS program_requirements (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    program_id BIGINT NOT NULL,
    term_id BIGINT NOT NULL,
    requirement_group_id BIGINT NULL,
    course_id BIGINT NOT NULL,
    requirement_type ENUM('COURSE', 'ELECTIVE_GROUP') NOT NULL,
    sequence INT NOT NULL,
    display_title VARCHAR(255) NULL,
    notes TEXT NULL,
    is_recommended BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT fk_program_requirements_program FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE,
    CONSTRAINT fk_program_requirements_term FOREIGN KEY (term_id) REFERENCES terms(id) ON DELETE CASCADE,
    CONSTRAINT fk_program_requirements_group FOREIGN KEY (requirement_group_id) REFERENCES requirement_groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_program_requirements_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS student_course_progress (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    student_id BIGINT NOT NULL,
    program_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    status ENUM('NOT_STARTED', 'PLANNED', 'IN_PROGRESS', 'COMPLETED') NOT NULL,
    updated_at DATETIME NOT NULL,
    CONSTRAINT fk_student_course_progress_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_student_course_progress_program FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE,
    CONSTRAINT fk_student_course_progress_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    CONSTRAINT uq_student_program_course_progress UNIQUE (student_id, program_id, course_id)
);

CREATE TABLE IF NOT EXISTS elective_selections (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    student_id BIGINT NOT NULL,
    program_id BIGINT NOT NULL,
    elective_group_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    updated_at DATETIME NOT NULL,
    CONSTRAINT fk_elective_selections_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    CONSTRAINT fk_elective_selections_program FOREIGN KEY (program_id) REFERENCES programs(id) ON DELETE CASCADE,
    CONSTRAINT fk_elective_selections_elective_group FOREIGN KEY (elective_group_id) REFERENCES elective_groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_elective_selections_course FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    CONSTRAINT uq_student_program_elective UNIQUE (student_id, program_id, elective_group_id)
);
`

func BootstrapMySQL(ctx context.Context, db *sql.DB, studentExternalKey string) error {
	if err := execStatements(ctx, db, mysqlSchema); err != nil {
		return err
	}

	if err := ensureDemoStudentProgress(ctx, db, studentExternalKey); err != nil {
		return err
	}

	var universityCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM universities`).Scan(&universityCount); err != nil {
		return err
	}

	if universityCount == 0 {
		if err := seedCatalog(ctx, db, studentExternalKey); err != nil {
			return err
		}
	}

	return nil
}

func seedCatalog(ctx context.Context, db *sql.DB, studentExternalKey string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `INSERT INTO universities (code, name) VALUES (?, ?)`, catalog.WaterlooCode, "University of Waterloo")
	if err != nil {
		return err
	}
	universityID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	courseIDs := map[string]int64{}

	for _, program := range catalog.ProgramDefinitions() {
		if err := ensureCoursesForProgram(ctx, tx, universityID, courseIDs, program); err != nil {
			return err
		}

		programResult, err := tx.ExecContext(
			ctx,
			`INSERT INTO programs (university_id, code, name, degree, description) VALUES (?, ?, ?, ?, ?)`,
			universityID,
			program.ProgramCode,
			program.Name,
			program.Degree,
			program.Description,
		)
		if err != nil {
			return err
		}
		programID, err := programResult.LastInsertId()
		if err != nil {
			return err
		}

		templateResult, err := tx.ExecContext(
			ctx,
			`INSERT INTO program_plan_templates (program_id, title, version) VALUES (?, ?, ?)`,
			programID,
			fmt.Sprintf("%s Roadmap", program.Name),
			"2026.1",
		)
		if err != nil {
			return err
		}
		templateID, err := templateResult.LastInsertId()
		if err != nil {
			return err
		}

		groupIDs := map[string]int64{}
		electiveGroupIDs := map[string]int64{}

		for _, term := range program.Terms {
			termResult, err := tx.ExecContext(
				ctx,
				`INSERT INTO terms (program_plan_template_id, code, label, year, season, sequence) VALUES (?, ?, ?, ?, ?, ?)`,
				templateID,
				term.Code,
				term.Label,
				term.Year,
				term.Season,
				term.Sequence,
			)
			if err != nil {
				return err
			}
			termID, err := termResult.LastInsertId()
			if err != nil {
				return err
			}

			for _, requirement := range term.Requirements {
				if requirement.Course != nil {
					if _, err := tx.ExecContext(
						ctx,
						`INSERT INTO program_requirements (program_id, term_id, requirement_group_id, course_id, requirement_type, sequence, display_title, notes, is_recommended)
						VALUES (?, ?, NULL, ?, ?, ?, ?, ?, TRUE)`,
						programID,
						termID,
						courseIDs[requirement.Course.Course.Code],
						requirement.Kind,
						requirement.Sequence,
						requirement.Course.Code,
						requirement.Course.Notes,
					); err != nil {
						return err
					}
					continue
				}

				if requirement.Group == nil {
					continue
				}

				groupResult, err := tx.ExecContext(
					ctx,
					`INSERT INTO requirement_groups (program_id, term_id, code, title, description, kind, sequence, min_selections, max_selections)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					programID,
					termID,
					requirement.Group.Code,
					requirement.Group.Title,
					requirement.Group.Description,
					requirement.Group.Kind,
					requirement.Group.Sequence,
					requirement.Group.MinSelections,
					requirement.Group.MaxSelections,
				)
				if err != nil {
					return err
				}
				groupID, err := groupResult.LastInsertId()
				if err != nil {
					return err
				}
				groupIDs[requirement.Group.Code] = groupID

				electiveResult, err := tx.ExecContext(
					ctx,
					`INSERT INTO elective_groups (requirement_group_id, selection_label, credits_required) VALUES (?, ?, ?)`,
					groupID,
					requirement.Group.Description,
					0.5,
				)
				if err != nil {
					return err
				}
				electiveGroupID, err := electiveResult.LastInsertId()
				if err != nil {
					return err
				}
				electiveGroupIDs[requirement.Group.Code] = electiveGroupID

				for _, option := range requirement.Group.Options {
					if _, err := tx.ExecContext(
						ctx,
						`INSERT INTO program_requirements (program_id, term_id, requirement_group_id, course_id, requirement_type, sequence, display_title, notes, is_recommended)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, TRUE)`,
						programID,
						termID,
						groupID,
						courseIDs[option.Course.Code],
						model.RequirementKindElectiveGroup,
						option.Sequence,
						option.Course.Code,
						option.Notes,
					); err != nil {
						return err
					}
				}
			}
		}

		if err := insertPrerequisites(ctx, tx, courseIDs, program); err != nil {
			return err
		}

		if err := insertDemoProgress(ctx, tx, studentExternalKey, programID, courseIDs, electiveGroupIDs, catalog.DemoProgress()[program.ProgramCode]); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ensureCoursesForProgram(ctx context.Context, tx *sql.Tx, universityID int64, courseIDs map[string]int64, program model.ProgramDefinition) error {
	for _, courseDef := range iterCourses(program) {
		if _, exists := courseIDs[courseDef.Course.Code]; exists {
			continue
		}

		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO courses (university_id, code, title, subject, credits, description) VALUES (?, ?, ?, ?, ?, ?)`,
			universityID,
			courseDef.Course.Code,
			courseDef.Course.Title,
			valueOrDefault(courseDef.Course.Subject, strings.Split(courseDef.Course.Code, " ")[0]),
			courseDef.Course.Credits,
			courseDef.Course.Description,
		)
		if err != nil {
			return err
		}
		courseID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		courseIDs[courseDef.Course.Code] = courseID
	}

	return nil
}

func insertPrerequisites(ctx context.Context, tx *sql.Tx, courseIDs map[string]int64, program model.ProgramDefinition) error {
	seen := map[string]bool{}
	for _, courseDef := range iterCourses(program) {
		for _, prerequisite := range courseDef.Prerequisites {
			key := fmt.Sprintf("%s->%s", courseDef.Course.Code, prerequisite.CourseCode)
			if seen[key] {
				continue
			}
			seen[key] = true

			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO prerequisite_rules (course_id, prerequisite_course_id, minimum_grade, is_corequisite) VALUES (?, ?, ?, ?)`,
				courseIDs[courseDef.Course.Code],
				courseIDs[prerequisite.CourseCode],
				prerequisite.MinimumGrade,
				prerequisite.IsCorequisite,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureDemoStudentProgress(ctx context.Context, db *sql.DB, studentExternalKey string) error {
	var studentCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM students WHERE external_key = ?`, studentExternalKey).Scan(&studentCount); err != nil {
		return err
	}
	if studentCount > 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO students (external_key, full_name, email) VALUES (?, ?, ?)`,
		studentExternalKey,
		"Demo Student",
		"demo.student@planahead.local",
	); err != nil {
		return err
	}

	return tx.Commit()
}

func insertDemoProgress(
	ctx context.Context,
	tx *sql.Tx,
	studentExternalKey string,
	programID int64,
	courseIDs map[string]int64,
	electiveGroupIDs map[string]int64,
	snapshot model.ProgressSnapshot,
) error {
	if len(snapshot.CourseStatuses) == 0 && len(snapshot.ElectiveSelections) == 0 {
		return nil
	}

	var studentID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM students WHERE external_key = ?`, studentExternalKey).Scan(&studentID); err != nil {
		return err
	}

	now := time.Now().UTC()
	for courseCode, status := range snapshot.CourseStatuses {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO student_course_progress (student_id, program_id, course_id, status, updated_at) VALUES (?, ?, ?, ?, ?)`,
			studentID,
			programID,
			courseIDs[courseCode],
			status,
			now,
		); err != nil {
			return err
		}
	}

	for groupCode, courseCode := range snapshot.ElectiveSelections {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO elective_selections (student_id, program_id, elective_group_id, course_id, updated_at) VALUES (?, ?, ?, ?, ?)`,
			studentID,
			programID,
			electiveGroupIDs[groupCode],
			courseIDs[courseCode],
			now,
		); err != nil {
			return err
		}
	}

	return nil
}

func iterCourses(program model.ProgramDefinition) []model.CourseRequirementDefinition {
	result := make([]model.CourseRequirementDefinition, 0)
	seen := map[string]bool{}

	for _, term := range program.Terms {
		for _, requirement := range term.Requirements {
			if requirement.Course != nil && !seen[requirement.Course.Course.Code] {
				seen[requirement.Course.Course.Code] = true
				result = append(result, *requirement.Course)
			}

			if requirement.Group == nil {
				continue
			}

			for _, option := range requirement.Group.Options {
				if seen[option.Course.Code] {
					continue
				}
				seen[option.Course.Code] = true
				result = append(result, option)
			}
		}
	}

	return result
}

func valueOrDefault(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}

func execStatements(ctx context.Context, db *sql.DB, script string) error {
	for _, statement := range strings.Split(script, ";") {
		trimmed := strings.TrimSpace(statement)
		if trimmed == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, trimmed); err != nil {
			return err
		}
	}

	return nil
}
