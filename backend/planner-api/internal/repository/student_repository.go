package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"planahead/planner-api/internal/model"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) GetProgressSnapshot(ctx context.Context, universityCode string, programCode string, studentExternalKey string) (model.ProgressSnapshot, error) {
	studentID, err := r.getOrCreateStudent(ctx, studentExternalKey)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	programID, err := r.getProgramID(ctx, universityCode, programCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}

	courseStatuses := map[string]model.CourseStatus{}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT c.code, scp.status
		FROM student_course_progress scp
		INNER JOIN courses c ON c.id = scp.course_id
		WHERE scp.student_id = ? AND scp.program_id = ?`,
		studentID,
		programID,
	)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var courseCode string
		var status model.CourseStatus
		if err := rows.Scan(&courseCode, &status); err != nil {
			return model.ProgressSnapshot{}, err
		}
		courseStatuses[courseCode] = status
	}
	if err := rows.Err(); err != nil {
		return model.ProgressSnapshot{}, err
	}

	electiveSelections := map[string]string{}
	selectionRows, err := r.db.QueryContext(
		ctx,
		`SELECT rg.code, c.code
		FROM elective_selections es
		INNER JOIN elective_groups eg ON eg.id = es.elective_group_id
		INNER JOIN requirement_groups rg ON rg.id = eg.requirement_group_id
		INNER JOIN courses c ON c.id = es.course_id
		WHERE es.student_id = ? AND es.program_id = ?`,
		studentID,
		programID,
	)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	defer selectionRows.Close()

	for selectionRows.Next() {
		var groupCode string
		var courseCode string
		if err := selectionRows.Scan(&groupCode, &courseCode); err != nil {
			return model.ProgressSnapshot{}, err
		}
		electiveSelections[groupCode] = courseCode
	}

	return model.ProgressSnapshot{
		CourseStatuses:     courseStatuses,
		ElectiveSelections: electiveSelections,
	}, selectionRows.Err()
}

func (r *StudentRepository) UpdateCourseStatus(ctx context.Context, universityCode string, programCode string, courseCode string, status model.CourseStatus, studentExternalKey string) (model.ProgressSnapshot, error) {
	studentID, err := r.getOrCreateStudent(ctx, studentExternalKey)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	programID, err := r.getProgramID(ctx, universityCode, programCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	courseID, err := r.getCourseID(ctx, universityCode, courseCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}

	if status == model.CourseStatusNotStarted {
		if _, err := r.db.ExecContext(
			ctx,
			`DELETE FROM student_course_progress WHERE student_id = ? AND program_id = ? AND course_id = ?`,
			studentID,
			programID,
			courseID,
		); err != nil {
			return model.ProgressSnapshot{}, err
		}
		return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
	}

	now := time.Now().UTC()
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO student_course_progress (student_id, program_id, course_id, status, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE status = VALUES(status), updated_at = VALUES(updated_at)`,
		studentID,
		programID,
		courseID,
		status,
		now,
	); err != nil {
		return model.ProgressSnapshot{}, err
	}

	return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
}

func (r *StudentRepository) SelectElective(ctx context.Context, universityCode string, programCode string, groupCode string, courseCode string, studentExternalKey string) (model.ProgressSnapshot, error) {
	studentID, err := r.getOrCreateStudent(ctx, studentExternalKey)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	programID, err := r.getProgramID(ctx, universityCode, programCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	electiveGroupID, err := r.getElectiveGroupID(ctx, programID, groupCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	courseID, err := r.getCourseID(ctx, universityCode, courseCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	if err := r.validateElectiveOption(ctx, electiveGroupID, courseID); err != nil {
		return model.ProgressSnapshot{}, err
	}

	now := time.Now().UTC()
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO elective_selections (student_id, program_id, elective_group_id, course_id, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE course_id = VALUES(course_id), updated_at = VALUES(updated_at)`,
		studentID,
		programID,
		electiveGroupID,
		courseID,
		now,
	); err != nil {
		return model.ProgressSnapshot{}, err
	}

	return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
}

func (r *StudentRepository) ClearElectiveSelection(ctx context.Context, universityCode string, programCode string, groupCode string, studentExternalKey string) (model.ProgressSnapshot, error) {
	studentID, err := r.getOrCreateStudent(ctx, studentExternalKey)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	programID, err := r.getProgramID(ctx, universityCode, programCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}
	electiveGroupID, err := r.getElectiveGroupID(ctx, programID, groupCode)
	if err != nil {
		return model.ProgressSnapshot{}, err
	}

	if _, err := r.db.ExecContext(
		ctx,
		`DELETE FROM elective_selections WHERE student_id = ? AND program_id = ? AND elective_group_id = ?`,
		studentID,
		programID,
		electiveGroupID,
	); err != nil {
		return model.ProgressSnapshot{}, err
	}

	return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
}

func (r *StudentRepository) getOrCreateStudent(ctx context.Context, externalKey string) (int64, error) {
	var studentID int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM students WHERE external_key = ?`, externalKey).Scan(&studentID)
	if err == nil {
		return studentID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO students (external_key, full_name, email) VALUES (?, ?, ?)`,
		externalKey,
		"Demo Student",
		fmt.Sprintf("%s@planahead.local", externalKey),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *StudentRepository) getProgramID(ctx context.Context, universityCode string, programCode string) (int64, error) {
	var programID int64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT p.id
		FROM programs p
		INNER JOIN universities u ON u.id = p.university_id
		WHERE u.code = ? AND p.code = ?`,
		universityCode,
		programCode,
	).Scan(&programID)
	if err != nil {
		return 0, fmt.Errorf("program not found: %w", err)
	}
	return programID, nil
}

func (r *StudentRepository) getCourseID(ctx context.Context, universityCode string, courseCode string) (int64, error) {
	var courseID int64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT c.id
		FROM courses c
		INNER JOIN universities u ON u.id = c.university_id
		WHERE u.code = ? AND c.code = ?`,
		universityCode,
		courseCode,
	).Scan(&courseID)
	if err != nil {
		return 0, fmt.Errorf("course not found: %w", err)
	}
	return courseID, nil
}

func (r *StudentRepository) getElectiveGroupID(ctx context.Context, programID int64, groupCode string) (int64, error) {
	var electiveGroupID int64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT eg.id
		FROM elective_groups eg
		INNER JOIN requirement_groups rg ON rg.id = eg.requirement_group_id
		WHERE rg.program_id = ? AND rg.code = ?`,
		programID,
		groupCode,
	).Scan(&electiveGroupID)
	if err != nil {
		return 0, fmt.Errorf("elective group not found: %w", err)
	}
	return electiveGroupID, nil
}

func (r *StudentRepository) validateElectiveOption(ctx context.Context, electiveGroupID int64, courseID int64) error {
	var count int
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM program_requirements pr
		INNER JOIN elective_groups eg ON eg.requirement_group_id = pr.requirement_group_id
		WHERE eg.id = ? AND pr.course_id = ?`,
		electiveGroupID,
		courseID,
	).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("course is not a valid option for that elective group")
	}
	return nil
}
