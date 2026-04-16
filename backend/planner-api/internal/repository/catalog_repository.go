package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"planahead/planner-api/internal/model"
	"planahead/planner-api/internal/waterloo"
)

type CatalogRepository struct {
	db     *sql.DB
	client *waterloo.Client
	cache  repositoryCache
}

type repositoryCache struct {
	mu              sync.RWMutex
	catalogID       string
	catalogFetched  time.Time
	programs        map[string]cachedProgramDefinition
}

type cachedProgramDefinition struct {
	definition *model.ProgramDefinition
	fetchedAt  time.Time
}

const (
	catalogCacheTTL = 12 * time.Hour
	programCacheTTL = 12 * time.Hour
)

func NewCatalogRepository(db *sql.DB, client *waterloo.Client) *CatalogRepository {
	return &CatalogRepository{
		db:     db,
		client: client,
		cache: repositoryCache{
			programs: map[string]cachedProgramDefinition{},
		},
	}
}

func (r *CatalogRepository) SyncWaterlooPrograms(ctx context.Context) error {
	catalogID, err := r.currentCatalogID(ctx)
	if err != nil {
		return err
	}

	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO universities (code, name)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE name = VALUES(name)`,
		waterloo.UniversityCode,
		waterloo.UniversityName,
	); err != nil {
		return err
	}

	var universityID int64
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT id FROM universities WHERE code = ?`,
		waterloo.UniversityCode,
	).Scan(&universityID); err != nil {
		return err
	}

	items, err := r.client.ListPrograms(ctx, catalogID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	currentProgramCodes := make([]string, 0)
	for _, item := range items {
		if !strings.EqualFold(item.UndergraduateCredential.Name, "Major") {
			continue
		}

		summary := waterloo.ToProgramSummary(item)
		currentProgramCodes = append(currentProgramCodes, summary.Code)

		if _, err := r.db.ExecContext(
			ctx,
			`INSERT INTO programs (university_id, code, name, degree, description)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				degree = VALUES(degree),
				description = VALUES(description)`,
			universityID,
			summary.Code,
			summary.Name,
			summary.Degree,
			summary.Description,
		); err != nil {
			return err
		}
	}

	if len(currentProgramCodes) == 0 {
		return fmt.Errorf("waterloo sync returned zero major programs at %s", now.Format(time.RFC3339))
	}

	placeholders := strings.Repeat("?,", len(currentProgramCodes))
	placeholders = strings.TrimSuffix(placeholders, ",")

	args := make([]any, 0, len(currentProgramCodes)+1)
	args = append(args, universityID)
	for _, code := range currentProgramCodes {
		args = append(args, code)
	}

	query := fmt.Sprintf(
		`DELETE FROM programs
		WHERE university_id = ?
		AND code NOT IN (%s)`,
		placeholders,
	)
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
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
	cacheKey := universityCode + ":" + programCode
	if cached := r.cachedProgram(cacheKey); cached != nil {
		return cached, nil
	}

	var summary model.ProgramSummary
	err := r.db.QueryRowContext(
		ctx,
		`SELECT p.code, p.name, p.degree, p.description, u.code
		FROM programs p
		INNER JOIN universities u ON u.id = p.university_id
		WHERE u.code = ? AND p.code = ?`,
		universityCode,
		programCode,
	).Scan(&summary.Code, &summary.Name, &summary.Degree, &summary.Description, &summary.UniversityCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("program %s/%s not found", universityCode, programCode)
		}
		return nil, err
	}

	catalogID, err := r.currentCatalogID(ctx)
	if err != nil {
		return nil, err
	}

	detail, err := r.client.Program(ctx, catalogID, programCode)
	if err != nil {
		return nil, err
	}

	program, err := waterloo.BuildProgramDefinition(
		ctx,
		summary,
		detail,
		func(ctx context.Context, courseID string) (*waterloo.CourseDetail, error) {
			return r.client.CourseByID(ctx, catalogID, courseID)
		},
	)
	if err != nil {
		return nil, err
	}

	if err := r.upsertCourses(ctx, detail, program.Terms); err != nil {
		return nil, err
	}

	r.storeProgram(cacheKey, program)
	return program, nil
}

func (r *CatalogRepository) upsertCourses(ctx context.Context, _ *waterloo.ProgramDetail, terms []model.TermDefinition) error {
	var universityID int64
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT id FROM universities WHERE code = ?`,
		waterloo.UniversityCode,
	).Scan(&universityID); err != nil {
		return err
	}

	seen := map[string]model.CourseDefinition{}
	for _, term := range terms {
		for _, requirement := range term.Requirements {
			if requirement.Course != nil && !strings.HasPrefix(requirement.Course.Course.Code, "RULE-") {
				seen[requirement.Course.Course.Code] = requirement.Course.Course
			}
			if requirement.Group == nil {
				continue
			}
			for _, option := range requirement.Group.Options {
				seen[option.Course.Code] = option.Course
			}
		}
	}

	codes := make([]string, 0, len(seen))
	for code := range seen {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	for _, code := range codes {
		course := seen[code]
		subject := ""
		if course.Subject != nil {
			subject = *course.Subject
		}
		var description any
		if course.Description != nil {
			description = *course.Description
		}

		if _, err := r.db.ExecContext(
			ctx,
			`INSERT INTO courses (university_id, code, title, subject, credits, description)
			VALUES (?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				title = VALUES(title),
				subject = VALUES(subject),
				credits = VALUES(credits),
				description = VALUES(description)`,
			universityID,
			course.Code,
			course.Title,
			subject,
			course.Credits,
			description,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *CatalogRepository) currentCatalogID(ctx context.Context) (string, error) {
	r.cache.mu.RLock()
	if r.cache.catalogID != "" && time.Since(r.cache.catalogFetched) < catalogCacheTTL {
		catalogID := r.cache.catalogID
		r.cache.mu.RUnlock()
		return catalogID, nil
	}
	r.cache.mu.RUnlock()

	catalog, err := r.client.CurrentCatalog(ctx)
	if err != nil {
		return "", err
	}

	r.cache.mu.Lock()
	r.cache.catalogID = catalog.ID
	r.cache.catalogFetched = time.Now().UTC()
	r.cache.mu.Unlock()

	return catalog.ID, nil
}

func (r *CatalogRepository) cachedProgram(cacheKey string) *model.ProgramDefinition {
	r.cache.mu.RLock()
	cached, exists := r.cache.programs[cacheKey]
	r.cache.mu.RUnlock()

	if !exists || cached.definition == nil || time.Since(cached.fetchedAt) >= programCacheTTL {
		return nil
	}

	return cached.definition
}

func (r *CatalogRepository) storeProgram(cacheKey string, definition *model.ProgramDefinition) {
	r.cache.mu.Lock()
	r.cache.programs[cacheKey] = cachedProgramDefinition{
		definition: definition,
		fetchedAt:  time.Now().UTC(),
	}
	r.cache.mu.Unlock()
}
