from __future__ import annotations

from sqlalchemy.exc import OperationalError
from sqlalchemy.orm import Session

from app.db.repositories.student_repository import StudentRepository
from app.modules.universities.common.requirement_types import CourseStatus
from app.modules.universities.catalog_fallback import get_demo_progress_snapshot


class StudentProgressService:
    def __init__(self, session: Session) -> None:
        self.student_repository = StudentRepository(session)

    def get_student_progress(self, university_code: str, program_code: str, student_external_key: str = StudentRepository.DEMO_STUDENT_KEY):
        try:
            return self.student_repository.get_progress_snapshot(university_code, program_code, student_external_key)
        except OperationalError:
            return get_demo_progress_snapshot(university_code, program_code)

    def update_course_status(
        self,
        university_code: str,
        program_code: str,
        course_code: str,
        status: CourseStatus,
        student_external_key: str = StudentRepository.DEMO_STUDENT_KEY,
    ):
        return self.student_repository.update_course_status(
            university_code=university_code,
            program_code=program_code,
            course_code=course_code,
            status=status,
            student_external_key=student_external_key,
        )

    def select_elective(
        self,
        university_code: str,
        program_code: str,
        group_code: str,
        course_code: str,
        student_external_key: str = StudentRepository.DEMO_STUDENT_KEY,
    ):
        return self.student_repository.select_elective(
            university_code=university_code,
            program_code=program_code,
            group_code=group_code,
            course_code=course_code,
            student_external_key=student_external_key,
        )

    def clear_elective_selection(
        self,
        university_code: str,
        program_code: str,
        group_code: str,
        student_external_key: str = StudentRepository.DEMO_STUDENT_KEY,
    ):
        return self.student_repository.clear_elective_selection(
            university_code=university_code,
            program_code=program_code,
            group_code=group_code,
            student_external_key=student_external_key,
        )
