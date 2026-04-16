package waterloo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	BaseURL        = "https://uwaterloocm.kuali.co/api/v1/catalog"
	UniversityCode = "WATERLOO"
	UniversityName = "University of Waterloo"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Catalog struct {
	ID    string `json:"_id"`
	Title string `json:"title"`
}

type CredentialType struct {
	Name string `json:"name"`
}

type FieldOfStudy struct {
	Name string `json:"name"`
}

type Faculty struct {
	Name string `json:"name"`
}

type ProgramListItem struct {
	ID                      string         `json:"id"`
	PID                     string         `json:"pid"`
	Code                    string         `json:"code"`
	Title                   string         `json:"title"`
	UndergraduateCredential CredentialType `json:"undergraduateCredentialType"`
	FieldOfStudy            FieldOfStudy   `json:"fieldOfStudy"`
	FacultyCalendarDisplay  Faculty        `json:"facultyCalendarDisplay"`
}

type ProgramDetail struct {
	ID                                   string         `json:"id"`
	PID                                  string         `json:"pid"`
	Code                                 string         `json:"code"`
	Title                                string         `json:"title"`
	UndergraduateCredential              CredentialType `json:"undergraduateCredentialType"`
	FieldOfStudy                         FieldOfStudy   `json:"fieldOfStudy"`
	FacultyCalendarDisplay               Faculty        `json:"facultyCalendarDisplay"`
	RequiredCoursesTermByTerm            string         `json:"requiredCoursesTermByTerm"`
	CourseRequirementsNoUnits            string         `json:"courseRequirementsNoUnits"`
	Requirements                         string         `json:"requirements"`
	CourseListsNew                       string         `json:"courseListsNew"`
	GraduationRequirements               string         `json:"graduationRequirements"`
	DegreeRequirements                   string         `json:"degreeRequirements"`
	AdmissionRequirements                string         `json:"admissionRequirements"`
	DeclarationRequirements              string         `json:"declarationRequirements"`
	AdditionalConstraints                string         `json:"additionalConstraints"`
	CoOperativeRequirementsUndergraduate string         `json:"coOperativeRequirementsUndergraduate"`
	SpecializationDetails                string         `json:"specializationDetails"`
}

type CourseCredits struct {
	Value string `json:"value"`
}

type CourseDetail struct {
	ID              string        `json:"id"`
	PID             string        `json:"pid"`
	CatalogCourseID string        `json:"__catalogCourseId"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Credits         CourseCredits `json:"credits"`
	SubjectCode     struct {
		Name string `json:"name"`
	} `json:"subjectCode"`
	Prerequisites  string `json:"prerequisites"`
	Corequisites   string `json:"corequisites"`
	Antirequisites string `json:"antirequisites"`
}

func (c *Client) CurrentCatalog(ctx context.Context) (*Catalog, error) {
	var catalog Catalog
	if err := c.getJSON(ctx, "/public/catalogs/current", &catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

func (c *Client) ListPrograms(ctx context.Context, catalogID string) ([]ProgramListItem, error) {
	var items []ProgramListItem
	if err := c.getJSON(ctx, fmt.Sprintf("/programs/%s", catalogID), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) Program(ctx context.Context, catalogID string, pid string) (*ProgramDetail, error) {
	var detail ProgramDetail
	if err := c.getJSON(ctx, fmt.Sprintf("/program/%s/%s", catalogID, pid), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (c *Client) CourseByID(ctx context.Context, catalogID string, courseID string) (*CourseDetail, error) {
	var detail CourseDetail
	if err := c.getJSON(ctx, fmt.Sprintf("/course/byId/%s/%s", catalogID, courseID), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

func (c *Client) getJSON(ctx context.Context, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("waterloo api %s returned %s", path, res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return err
	}

	return nil
}
