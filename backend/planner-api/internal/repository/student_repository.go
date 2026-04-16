package repository

import (
	"context"
	"database/sql"
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
	courseStatuses := map[string]model.CourseStatus{}
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT course_code, status
		FROM planner_student_course_progress
		WHERE student_external_key = ? AND university_code = ? AND program_code = ?`,
		studentExternalKey,
		universityCode,
		programCode,
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
		`SELECT group_code, course_code
		FROM planner_elective_selections
		WHERE student_external_key = ? AND university_code = ? AND program_code = ?`,
		studentExternalKey,
		universityCode,
		programCode,
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
	if status == model.CourseStatusNotStarted {
		if _, err := r.db.ExecContext(
			ctx,
			`DELETE FROM planner_student_course_progress
			WHERE student_external_key = ? AND university_code = ? AND program_code = ? AND course_code = ?`,
			studentExternalKey,
			universityCode,
			programCode,
			courseCode,
		); err != nil {
			return model.ProgressSnapshot{}, err
		}
		return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
	}

	now := time.Now().UTC()
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO planner_student_course_progress
			(student_external_key, university_code, program_code, course_code, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			updated_at = VALUES(updated_at)`,
		studentExternalKey,
		universityCode,
		programCode,
		courseCode,
		status,
		now,
	); err != nil {
		return model.ProgressSnapshot{}, err
	}

	return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
}

func (r *StudentRepository) SelectElective(ctx context.Context, universityCode string, programCode string, groupCode string, courseCode string, studentExternalKey string) (model.ProgressSnapshot, error) {
	now := time.Now().UTC()
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO planner_elective_selections
			(student_external_key, university_code, program_code, group_code, course_code, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			course_code = VALUES(course_code),
			updated_at = VALUES(updated_at)`,
		studentExternalKey,
		universityCode,
		programCode,
		groupCode,
		courseCode,
		now,
	); err != nil {
		return model.ProgressSnapshot{}, err
	}

	return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
}

func (r *StudentRepository) ClearElectiveSelection(ctx context.Context, universityCode string, programCode string, groupCode string, studentExternalKey string) (model.ProgressSnapshot, error) {
	if _, err := r.db.ExecContext(
		ctx,
		`DELETE FROM planner_elective_selections
		WHERE student_external_key = ? AND university_code = ? AND program_code = ? AND group_code = ?`,
		studentExternalKey,
		universityCode,
		programCode,
		groupCode,
	); err != nil {
		return model.ProgressSnapshot{}, err
	}

	return r.GetProgressSnapshot(ctx, universityCode, programCode, studentExternalKey)
}
