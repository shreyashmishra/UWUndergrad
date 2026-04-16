package waterloo

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"planahead/planner-api/internal/model"
)

var (
	parentheticalDegreePattern = regexp.MustCompile(`^(.*)\s+\(([^()]+)\)$`)
	spaceCodePattern           = regexp.MustCompile(`^([A-Z]+)(\d[A-Z0-9]*)$`)
	whitespacePattern          = regexp.MustCompile(`\s+`)
	// Matches academic term labels like "1A Term", "2B", "Term 3A", "1A - Fall"
	academicTermPattern = regexp.MustCompile(`(?i)\b([1-9])\s*([AB])\b`)
)

type CourseFetcher func(ctx context.Context, courseID string) (*CourseDetail, error)

type courseRef struct {
	ID   string
	Code string
}

type parserState struct {
	fetchCourse      CourseFetcher
	courseCache      map[string]*CourseDetail
	cacheMu          sync.RWMutex
	sequenceByLabel  map[string]int32
	groupCounter     int
	syntheticCounter int
}

func ToProgramSummary(item ProgramListItem) model.ProgramSummary {
	name, degree := splitProgramTitle(item.Title, item.UndergraduateCredential.Name)
	description := strings.TrimSpace(
		strings.Join([]string{
			item.FieldOfStudy.Name,
			item.FacultyCalendarDisplay.Name,
			"Official University of Waterloo academic calendar import.",
		}, " · "),
	)

	return model.ProgramSummary{
		Code:           item.PID,
		Name:           name,
		Degree:         degree,
		Description:    description,
		UniversityCode: UniversityCode,
	}
}

func BuildProgramDefinition(
	ctx context.Context,
	summary model.ProgramSummary,
	detail *ProgramDetail,
	fetchCourse CourseFetcher,
) (*model.ProgramDefinition, error) {
	requirementHTML := firstNonEmpty(
		detail.RequiredCoursesTermByTerm,
		detail.CourseRequirementsNoUnits,
		detail.Requirements,
		detail.CourseListsNew,
	)

	description := firstNonEmpty(
		htmlToText(detail.GraduationRequirements),
		htmlToText(detail.DegreeRequirements),
		htmlToText(detail.SpecializationDetails),
		summary.Description,
	)

	parser := &parserState{
		fetchCourse:     fetchCourse,
		courseCache:     map[string]*CourseDetail{},
		sequenceByLabel: map[string]int32{},
	}

	parser.preloadCourses(ctx, requirementHTML)

	parsedTerms, err := parser.parseRequirements(ctx, requirementHTML)
	if err != nil {
		return nil, err
	}
	terms := parsedTerms

	for index := range terms {
		term := &terms[index]
		sort.SliceStable(term.Requirements, func(i int, j int) bool {
			return term.Requirements[i].Sequence < term.Requirements[j].Sequence
		})
	}

	return &model.ProgramDefinition{
		UniversityCode: summary.UniversityCode,
		ProgramCode:    summary.Code,
		Name:           summary.Name,
		Degree:         summary.Degree,
		Description:    description,
		Terms:          terms,
	}, nil
}

func (p *parserState) parseRequirements(ctx context.Context, requirementHTML string) ([]model.TermDefinition, error) {
	if strings.TrimSpace(requirementHTML) == "" {
		return nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(requirementHTML))
	if err != nil {
		return nil, err
	}

	requirementsByLabel := map[string][]model.TermRequirementDefinition{}
	labelOrder := make([]string, 0)

	ensureLabel := func(label string) {
		if _, exists := requirementsByLabel[label]; exists {
			return
		}
		labelOrder = append(labelOrder, label)
		requirementsByLabel[label] = []model.TermRequirementDefinition{}
	}

	sections := doc.Find("section")
	if sections.Length() == 0 {
		sections = doc.Selection
	}

	sections.Each(func(_ int, section *goquery.Selection) {
		defaultLabel := normalizeWhitespace(
			section.Find("header h2 span").First().Text(),
		)
		if defaultLabel == "" {
			defaultLabel = "Official Requirements"
		}
		ensureLabel(defaultLabel)

		section.Find("li[data-test]").Each(func(_ int, rule *goquery.Selection) {
			label := nearestGroupLabel(rule, defaultLabel)
			ensureLabel(label)
			requirementsByLabel[label] = append(requirementsByLabel[label], p.parseRule(ctx, rule, label)...)
		})
	})

	terms := make([]model.TermDefinition, 0, len(labelOrder))
	for index, label := range labelOrder {
		requirements := requirementsByLabel[label]
		if len(requirements) == 0 {
			continue
		}
		year, season, shortCode := parseAcademicTerm(label)
		termLabel := label
		if shortCode != "" {
			termLabel = shortCode // use "1A", "2B", etc. as the display label
		}
		terms = append(terms, model.TermDefinition{
			Code:         sanitizeCode(label, index+1),
			Label:        termLabel,
			Year:         year,
			Season:       season,
			Sequence:     int32(index + 1),
			Requirements: requirements,
		})
	}

	return terms, nil
}

// parseAcademicTerm extracts structured year/season metadata from a section label.
// Recognises patterns like "1A Term", "2B", "Term 3A", "4B - Winter" etc.
// Returns year=0 and an empty shortCode when the label is not a recognised academic term.
func parseAcademicTerm(label string) (year int32, season model.TermSeason, shortCode string) {
	matches := academicTermPattern.FindStringSubmatch(label)
	if len(matches) < 3 {
		return 0, model.TermSeasonFall, ""
	}
	y, _ := strconv.Atoi(matches[1])
	letter := strings.ToUpper(matches[2])
	if letter == "B" {
		return int32(y), model.TermSeasonWinter, matches[1] + letter
	}
	return int32(y), model.TermSeasonFall, matches[1] + letter
}

func (p *parserState) parseRule(ctx context.Context, rule *goquery.Selection, label string) []model.TermRequirementDefinition {
	text := normalizeWhitespace(rule.Text())
	if text == "" {
		return nil
	}

	courseRefs := extractCourseRefs(rule)
	sequence := p.nextSequence(label)

	switch {
	case len(courseRefs) == 0:
		return []model.TermRequirementDefinition{p.syntheticRequirement(text, sequence)}
	case isChoiceRule(text) && len(courseRefs) > 1:
		options := make([]model.CourseRequirementDefinition, 0, len(courseRefs))
		for _, ref := range courseRefs {
			options = append(options, p.courseRequirement(ctx, ref, 1))
		}
		groupCode := fmt.Sprintf("GROUP-%03d", p.groupCounter)
		p.groupCounter++
		description := text
		return []model.TermRequirementDefinition{
			{
				Kind:     model.RequirementKindElectiveGroup,
				Sequence: sequence,
				Group: &model.ElectiveGroupDefinition{
					Code:          groupCode,
					Title:         summarizeText(text, 72),
					Description:   description,
					Kind:          model.RequirementGroupKindOneOf,
					MinSelections: 1,
					MaxSelections: 1,
					Sequence:      sequence,
					Options:       options,
					Notes:         &description,
				},
			},
		}
	default:
		result := make([]model.TermRequirementDefinition, 0, len(courseRefs))
		for offset, ref := range courseRefs {
			courseRequirement := p.courseRequirement(ctx, ref, int32(offset)+1)
			result = append(result, model.TermRequirementDefinition{
				Kind:     model.RequirementKindCourse,
				Sequence: sequence + int32(offset),
				Course:   &courseRequirement,
			})
		}
		return result
	}
}

func (p *parserState) courseRequirement(ctx context.Context, ref courseRef, sequence int32) model.CourseRequirementDefinition {
	course := model.CourseDefinition{
		Code:    ref.Code,
		Title:   ref.Code,
		Credits: 0.5,
	}

	var noteParts []string
	detail, err := p.loadCourse(ctx, ref.ID)
	if err == nil && detail != nil {
		course.Code = formatCourseCode(firstNonEmpty(detail.CatalogCourseID, ref.Code))
		course.Title = detail.Title
		if credits, parseErr := strconv.ParseFloat(detail.Credits.Value, 64); parseErr == nil {
			course.Credits = credits
		}
		if description := normalizeWhitespace(detail.Description); description != "" {
			course.Description = stringPointer(description)
		}
		if subject := normalizeWhitespace(detail.SubjectCode.Name); subject != "" {
			course.Subject = stringPointer(subject)
		}
		if prerequisiteText := htmlToText(detail.Prerequisites); prerequisiteText != "" {
			noteParts = append(noteParts, "Official prerequisite: "+prerequisiteText)
		}
		if corequisiteText := htmlToText(detail.Corequisites); corequisiteText != "" {
			noteParts = append(noteParts, "Official corequisite: "+corequisiteText)
		}
		if antirequisiteText := htmlToText(detail.Antirequisites); antirequisiteText != "" {
			noteParts = append(noteParts, "Official antirequisite: "+antirequisiteText)
		}
	}

	var notes *string
	if len(noteParts) > 0 {
		joined := strings.Join(noteParts, "\n\n")
		notes = &joined
	}

	return model.CourseRequirementDefinition{
		Code:     course.Code,
		Course:   course,
		Sequence: sequence,
		Notes:    notes,
		Kind:     model.RequirementKindCourse,
	}
}

func (p *parserState) syntheticRequirement(text string, sequence int32) model.TermRequirementDefinition {
	p.syntheticCounter++
	code := fmt.Sprintf("RULE-%03d", p.syntheticCounter)
	title := summarizeText(text, 88)
	note := text

	return model.TermRequirementDefinition{
		Kind:     model.RequirementKindCourse,
		Sequence: sequence,
		Course: &model.CourseRequirementDefinition{
			Code: code,
			Course: model.CourseDefinition{
				Code:        code,
				Title:       title,
				Credits:     0,
				Description: stringPointer(text),
				Subject:     stringPointer("Requirement"),
			},
			Sequence: sequence,
			Notes:    &note,
			Kind:     model.RequirementKindCourse,
		},
	}
}


func (p *parserState) loadCourse(ctx context.Context, courseID string) (*CourseDetail, error) {
	if courseID == "" {
		return nil, nil
	}

	p.cacheMu.RLock()
	cached, ok := p.courseCache[courseID]
	p.cacheMu.RUnlock()
	if ok {
		return cached, nil
	}

	detail, err := p.fetchCourse(ctx, courseID)
	p.cacheMu.Lock()
	p.courseCache[courseID] = detail
	p.cacheMu.Unlock()
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (p *parserState) nextSequence(label string) int32 {
	p.sequenceByLabel[label]++
	return p.sequenceByLabel[label]
}

func extractCourseRefs(rule *goquery.Selection) []courseRef {
	seen := map[string]bool{}
	refs := make([]courseRef, 0)
	rule.Find(`a[href*="#/courses/view/"]`).Each(func(_ int, anchor *goquery.Selection) {
		href, exists := anchor.Attr("href")
		if !exists {
			return
		}
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		courseID := parts[len(parts)-1]
		code := formatCourseCode(normalizeWhitespace(anchor.Text()))
		if courseID == "" || code == "" || seen[courseID] {
			return
		}
		seen[courseID] = true
		refs = append(refs, courseRef{
			ID:   courseID,
			Code: code,
		})
	})
	return refs
}

func nearestGroupLabel(rule *goquery.Selection, fallback string) string {
	label := fallback
	rule.ParentsFiltered("div").EachWithBreak(func(_ int, parent *goquery.Selection) bool {
		text := normalizeWhitespace(parent.ChildrenFiltered("span.rules_groupHeader_37").First().Text())
		if text == "" {
			return true
		}
		label = text
		return false
	})
	return label
}

func isChoiceRule(text string) bool {
	normalized := strings.ToLower(text)
	return strings.Contains(normalized, "complete 1 of the following") ||
		strings.Contains(normalized, "complete 1 course from the following") ||
		strings.Contains(normalized, "choose 1")
}

func splitProgramTitle(title string, fallbackDegree string) (string, string) {
	matches := parentheticalDegreePattern.FindStringSubmatch(title)
	if len(matches) != 3 {
		return title, fallbackDegree
	}
	return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
}

func formatCourseCode(code string) string {
	code = strings.TrimSpace(strings.ReplaceAll(code, " ", ""))
	matches := spaceCodePattern.FindStringSubmatch(code)
	if len(matches) != 3 {
		return code
	}
	return matches[1] + " " + matches[2]
}

func htmlToText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return normalizeWhitespace(raw)
	}
	return normalizeWhitespace(doc.Text())
}

func normalizeWhitespace(value string) string {
	return strings.TrimSpace(whitespacePattern.ReplaceAllString(value, " "))
}

func sanitizeCode(label string, fallback int) string {
	label = strings.ToUpper(label)
	label = strings.ReplaceAll(label, " ", "_")
	label = strings.ReplaceAll(label, "-", "_")
	label = strings.ReplaceAll(label, "&", "AND")
	label = regexp.MustCompile(`[^A-Z0-9_]+`).ReplaceAllString(label, "")
	if label == "" {
		return fmt.Sprintf("SECTION_%02d", fallback)
	}
	return label
}

func summarizeText(text string, max int) string {
	text = normalizeWhitespace(text)
	if len(text) <= max {
		return text
	}
	return strings.TrimSpace(text[:max-1]) + "…"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func stringPointer(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	result := value
	return &result
}

func (p *parserState) preloadCourses(ctx context.Context, requirementHTML string) {
	if strings.TrimSpace(requirementHTML) == "" {
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(requirementHTML))
	if err != nil {
		return
	}

	uniqueRefs := map[string]struct{}{}
	doc.Find(`a[href*="#/courses/view/"]`).Each(func(_ int, anchor *goquery.Selection) {
		href, exists := anchor.Attr("href")
		if !exists {
			return
		}
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		courseID := parts[len(parts)-1]
		if courseID == "" {
			return
		}
		uniqueRefs[courseID] = struct{}{}
	})

	if len(uniqueRefs) == 0 {
		return
	}

	jobs := make(chan string)
	workerCount := 8
	if len(uniqueRefs) < workerCount {
		workerCount = len(uniqueRefs)
	}

	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for courseID := range jobs {
				fetchCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
				detail, err := p.fetchCourse(fetchCtx, courseID)
				cancel()

				p.cacheMu.Lock()
				if err == nil {
					p.courseCache[courseID] = detail
				} else if _, exists := p.courseCache[courseID]; !exists {
					p.courseCache[courseID] = nil
				}
				p.cacheMu.Unlock()
			}
		}()
	}

	for courseID := range uniqueRefs {
		jobs <- courseID
	}
	close(jobs)
	workers.Wait()
}
