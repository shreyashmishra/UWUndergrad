from __future__ import annotations

from sqlalchemy.exc import OperationalError
from sqlalchemy.orm import Session

from app.db.repositories.catalog_repository import CatalogRepository
from app.modules.universities.common.requirement_types import ProgressSnapshot
from app.modules.universities.catalog_fallback import (
    get_program_definition as get_fallback_program_definition,
    list_available_universities as list_fallback_universities,
    list_programs_by_university as list_fallback_programs,
)
from app.modules.universities.service import get_requirement_engine


class ProgramCatalogService:
    def __init__(self, session: Session) -> None:
        self.catalog_repository = CatalogRepository(session)

    def list_available_universities(self):
        try:
            universities = self.catalog_repository.list_universities()
            return universities or list_fallback_universities()
        except OperationalError:
            return list_fallback_universities()

    def list_programs_by_university(self, university_code: str):
        try:
            programs = self.catalog_repository.list_programs_by_university(university_code)
            return programs or list_fallback_programs(university_code)
        except OperationalError:
            return list_fallback_programs(university_code)

    def get_roadmap(self, university_code: str, program_code: str, progress: ProgressSnapshot):
        try:
            program_definition = self.catalog_repository.get_program_definition(university_code, program_code)
        except (OperationalError, ValueError):
            program_definition = get_fallback_program_definition(university_code, program_code)
        engine = get_requirement_engine(university_code)
        return engine.evaluate(program_definition, progress)

    def get_requirement_summary(self, university_code: str, program_code: str, progress: ProgressSnapshot):
        return self.get_roadmap(university_code, program_code, progress).summary
