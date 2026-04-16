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
