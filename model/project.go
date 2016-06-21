package model

import "time"

const (
	ProjectNewActionLabel = "new"
	ProjectInProgressActionLabel = "in_progress"
	ProjectStoppedActionLabel = "stopped"
	ProjectFinishedActionLabel = "finished"


)

// Please note that Images is not stored in the database as a nested collection (i.e. gorethink:"-")
// Thus, if you want to pull images for a give project, you must assign those directly as a slice of Image structures
type Project struct {
	ID           string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Name         string    `json:"name" gorethink:"name"`
	Description  string    `json:"description" gorethink:"description"`
	Status       string    `json:"status" gorethink:"status"`
	ActionStatus string    `json:"actionStatus" gorethink:"actionStatus"`
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
