package manager

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/shipyard/shipyard/model"
	"time"
)

const (
	ErrProjectResultsPreffixMsg = "Error in getting project results. "
)

// Methods related to the results structure
func (m DefaultManager) GetProjectResults(projectId string) (*model.ProjectResults, error) {

	project, err := m.Project(projectId)

	if err != nil {
		return nil, errors.New(ErrProjectResultsPreffixMsg + "Project does not exist")
	}

	tests, err := m.GetTests(project.ID)

	if err != nil {
		return nil, errors.New(ErrProjectResultsPreffixMsg + "Could not retrieve project tests")
	}

	testResults := []*model.TestResult{}

	for _, test := range tests {
		builds, err := m.GetBuildsByTestId(test.ID)

		if err != nil {
			log.Warnf("Could not get builds for test %s", test.ID)
			continue
		}

		for _, build := range builds {
			for _, buildResult := range build.Results {
				status := "failed"
				if buildResult.Successful {
					status = "success"
				}
				image := buildResult.TargetArtifact.Artifact.(model.Image)
				testResult := &model.TestResult{
					ImageId:       image.ID,
					ImageName:     image.Name,
					BuildId:       buildResult.BuildId,
					DockerImageId: "TODO: insert docker image id",
					TestId:        test.ID,
					TestName:      test.Name,
					// TODO: block value should be extracted from Test, but no value in it.
					Blocker: false,
					Status:  status,
					// Reusing the timestamps from the build and buildResult objects
					Date:       build.StartTime,
					EndDate:    buildResult.TimeStamp,
					AppliedTag: image.IlmTags,
					// TODO: what should be an adequate action in this case? the name of the test in the Provider?
					Action: test.ProviderJob.Name,
				}
				testResults = append(testResults, testResult)
			}
		}
	}
	STUB_TIME_CHANGE_ME := time.Now()
	// TODO: address issues of outdated or unncessary fields in the ProjectResult model.
	projectResults := &model.ProjectResults{
		ProjectId:      project.ID,
		Description:    project.Description,
		BuildId:        "TODO: remove this field",
		RunDate:        STUB_TIME_CHANGE_ME,
		EndDate:        STUB_TIME_CHANGE_ME,
		CreateDate:     STUB_TIME_CHANGE_ME,
		Author:         project.Author,
		ProjectVersion: "TODO: remove this field, no project version in project model anyways.",
		LastTagApplied: "TODO: remove this field, tags are managed in builds and images, not project level",
		LastUpdate:     STUB_TIME_CHANGE_ME,
		Updater:        "TODO: remove this field, what does it mean? is it the author?",
		TestResults:    testResults,
	}

	return projectResults, nil
}
