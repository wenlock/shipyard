package model

import "time"

// Please note that Images is not stored in the database as a nested collection (i.e. gorethink:"-")
// Thus, if you want to pull images for a give project, you must assign those directly as a slice of Image structures
type Project struct {
	ID           string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Name         string    `json:"name" gorethink:"name"`
	Description  string    `json:"description" gorethink:"description"`
	Status       string    `json:"status" gorethink:"status"`
	Images       []*Image  `json:"images,omitempty" gorethink:"-"`
	Tests        []*Test   `json:"tests,omitempty" gorethink:"-"`
	NeedsBuild   bool      `json:"needsBuild" gorethink:"needsBuild"`
	CreationTime time.Time `json:"creationTime" gorethink:"creationTime"`
	UpdateTime   time.Time `json:"updateTime" gorethink:"updateTime"`
	LastRunTime  time.Time `json:"lastRunTime" gorethink:"lastRunTime"`
	Author       string    `json:"author" gorethink:"author"`
	UpdatedBy    string    `json:"updatedBy" gorethink:"updatedBy"`
}

func (p *Project) NewProject(name string, description string, status string, images []*Image, tests []*Test, needsBuild bool, creationTime time.Time, updateTime time.Time, lastRunTime time.Time, author string, updatedBy string) *Project {

	return &Project{
		Name:         name,
		Description:  description,
		Status:       status,
		Images:       images,
		Tests:        tests,
		NeedsBuild:   needsBuild,
		CreationTime: creationTime,
		UpdateTime:   updateTime,
		LastRunTime:  lastRunTime,
		Author:       author,
		UpdatedBy:    updatedBy,
	}
}

type ProjectResults struct {
	ProjectId      string        `json:"projectId" gorethink:"projectId"`
	Description    string        `json:"description" gorethink:"description"`
	BuildId        string        `json:"buildId" gorethink:"buildId"`
	RunDate        time.Time     `json:"runDate" gorethink:"runDate"`
	EndDate        time.Time     `json:"endDate" gorethink:"endDate"`
	CreateDate     time.Time     `json:"createDate" gorethink:"createDate"`
	Author         string        `json:"author" gorethink:"author"`
	ProjectVersion string        `json:"projectVersion" gorethink:"lastRunTime"`
	LastTagApplied string        `json:"lastTagapplied" gorethink:"lastTagApplied"`
	LastUpdate     time.Time     `json:"lastUpdate" gorethink:"lastUpdate"`
	Updater        string        `json:"updater" gorethink:"updater"`
	TestResults    []*TestResult `json:"testResults" gorethink:"testResults"`
}

func NewProjectResults(
	projectId string,
	description string,
	buildId string,
	runDate time.Time,
	endDate time.Time,
	createDate time.Time,
	author string,
	projectVersion string,
	lastTagApplied string,
	lastUpdate time.Time,
	updater string,
	testResults []*TestResult,
) *ProjectResults {

	return &ProjectResults{
		ProjectId:      projectId,
		Description:    description,
		BuildId:        buildId,
		RunDate:        runDate,
		EndDate:        endDate,
		CreateDate:     createDate,
		Author:         author,
		ProjectVersion: projectVersion,
		LastTagApplied: lastTagApplied,
		LastUpdate:     lastUpdate,
		Updater:        updater,
		TestResults:    testResults,
	}
}