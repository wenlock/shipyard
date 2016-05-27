package model

import (
	"go/types"
	"time"
)

type Parameter struct {
	ParamName  string   `json:"paramName" gorethink:"paramName"`
	ParamValue []string `json:"paramValue" gorethink:"paramValue"`
}
type Test struct {
	ID               string            `json:"id,omitempty" gorethink:"id,omitempty"`
	Name             string            `json:"name" gorethink:"name"`
	Description      string            `json:"description" gorethink:"description"`
	Targets          []*TargetArtifact `json:"targets" gorethink:"targets"`
	SelectedTestType string            `json:"selectedTestType" gorethink:"selectedTestType"`
	Provider         *Provider         `json:"provider" gorethink:"provider"`
	ProviderJob      *ProviderJob      `json:"providerJob" gorethink:"providerJob"`
	Tagging          Tagging           `json:"tagging" gorethink:"tagging"`
	FromTag          string            `json:"fromTag" gorethink:"fromTag"`
	Parameters       []*Parameter      `json:"parameters" gorethink:"parameters"`
	ProjectId        string            `json:"projectId" gorethink:"projectId"`
}
type Tagging struct {
	OnSuccess string `json:"onSuccess" gorethink:"onSuccess"`
	OnFailure string `json:"onFailure" gorethink:"onFailure"`
}

func NewTest(
	name string,
	description string,
	targets []*TargetArtifact,
	selectedTestType string,
	projectId string,
	provider *Provider,
	providerjob *ProviderJob,
	parameters []*Parameter,
	successTag string,
	failTag string,
	fromTag string,
) *Test {
	test := new(Test)

	test.Name = name
	test.Description = description
	test.Targets = targets
	test.SelectedTestType = selectedTestType
	test.Tagging.OnSuccess = successTag
	test.Tagging.OnFailure = failTag
	test.FromTag = fromTag
	test.Parameters = parameters
	test.ProjectId = projectId

	return test
}

type TestResult struct {
	ImageId       string `json:"imageId" gorethink:"imageId"`
	ImageName     string `json:"imageName" gorethink:"imageName"`
	BuildId       string `json:"buildId" gorethink:"buildId"`
	DockerImageId string `json:"dockerImageId" gorethink:"dockerImageId"`
	TestId        string `json:"testId" gorethink:"testId"`
	TestName      string `json:"testName" gorethink:"testName"`
	Blocker       bool   `json:"blocker" gorethink:"blocker"`
	Status        string      `json:"status" gorethink:"status"`
	Date          time.Time   `json:"date" gorethink:"date"`
	EndDate       time.Time   `json:"endDate" gorethink:"endDate"`
	AppliedTag    []string    `json:"appliedTag" gorethink:"appliedTag"`
	Action        interface{} `json:"action" gorethink:"action"`
}

func NewTestResult(
	imageId string,
	imageName string,
	buildId string,
	dockerImageId string,
	testId string,
	testName string,
	blocker bool,
	status string,
	date time.Time,
	endDate time.Time,
	appliedTag []string,
	action *types.Object,
) *TestResult {
	testResult := new(TestResult)
	testResult.ImageId = imageId
	testResult.ImageName = imageName
	testResult.BuildId = buildId
	testResult.DockerImageId = dockerImageId
	testResult.TestId = testId
	testResult.BuildId = buildId
	testResult.TestName = testName
	testResult.Blocker = blocker
	testResult.Status = status
	testResult.Date = date
	testResult.EndDate = endDate
	testResult.AppliedTag = appliedTag
	testResult.Action = action

	return testResult
}
