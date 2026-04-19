package db

import (
	"context"
	"database/sql"
	"strings"
)

const mysqlSchema = `
CREATE TABLE IF NOT EXISTS planner_students (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    external_key VARCHAR(64) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS universities (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS programs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    university_id BIGINT NOT NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    degree VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    CONSTRAINT fk_programs_university FOREIGN KEY (university_id) REFERENCES universities(id) ON DELETE CASCADE,
    CONSTRAINT uq_program_university_code UNIQUE (university_id, code)
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

CREATE TABLE IF NOT EXISTS planner_student_course_progress (
    student_external_key VARCHAR(64) NOT NULL,
    university_code VARCHAR(64) NOT NULL,
    program_code VARCHAR(64) NOT NULL,
    course_code VARCHAR(64) NOT NULL,
    status ENUM('NOT_STARTED', 'PLANNED', 'IN_PROGRESS', 'COMPLETED') NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (student_external_key, program_code, course_code)
);

CREATE TABLE IF NOT EXISTS planner_elective_selections (
    student_external_key VARCHAR(64) NOT NULL,
    university_code VARCHAR(64) NOT NULL,
    program_code VARCHAR(64) NOT NULL,
    group_code VARCHAR(128) NOT NULL,
    course_code VARCHAR(64) NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (student_external_key, program_code, group_code)
);
`

func BootstrapMySQL(ctx context.Context, db *sql.DB) error {
	return execStatements(ctx, db, mysqlSchema)
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
