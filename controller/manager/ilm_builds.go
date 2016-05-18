package manager

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	//c "github.com/shipyard/shipyard/checker"
	"github.com/shipyard/shipyard/model"
	"time"
)

//methods related to the Build structure
func (m DefaultManager) GetBuildsByTestId(testId string) ([]*model.Build, error) {
	res, err := r.Table(tblNameBuilds).Filter(map[string]string{"testId": testId}).Run(m.session)
	if err != nil {
		return nil, err
	}
	builds := []*model.Build{}
	if err := res.All(&builds); err != nil {
		return nil, err
	}
	return builds, nil
}

func (m DefaultManager) GetBuild(buildId string) (*model.Build, error) {
	res, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Run(m.session)
	defer res.Close()
	if err != nil {
		return nil, err
	}
	if res.IsNil() {
		return nil, ErrBuildDoesNotExist
	}
	var build *model.Build
	if err := res.One(&build); err != nil {
		return nil, err
	}
	return build, nil
}

func (m DefaultManager) GetBuildStatus(projectId string, testId string, buildId string) (string, error) {
	build, err := m.GetBuild(buildId)

	if err != nil {
		log.Errorf("Could not get build status for build= %s err= %s", buildId, err.Error())
		return "", err
	}

	if build.Status == nil {
		errMsg := fmt.Sprintf("Could not get build status for build= %s, status is nil", buildId)
		log.Errorf(errMsg)
		return "", errors.New(errMsg)
	}
	return build.Status.Status, nil
}

func (m DefaultManager) GetBuildResults(buildId string) ([]*model.BuildResult, error) {
	build, err := m.GetBuild(buildId)

	if err != nil {
		log.Errorf("Could not get results for build= %s err= %", buildId, err.Error())
		return nil, err
	}

	return build.Results, nil
}

func (m DefaultManager) CreateBuild(projectId string, testId string, buildAction *model.BuildAction) (string, error) {

	var eventType string
	eventType = eventType
	var build *model.Build

	// In order to create a build we should get a start action
	if buildAction.Action != model.BuildStartActionLabel {
		log.Errorf("Build action should be %s, but received %s, error = %s",
			model.BuildStartActionLabel,
			buildAction.Action,
			ErrBuildActionNotSupported.Error(),
		)
		return "", ErrBuildActionNotSupported
	}

	// Instantiate a new Build object and fill out some fields
	build = &model.Build{}
	build.TestId = testId
	build.StartTime = time.Now()

	// we change the build's buildStatus to submitted
	build.Status = &model.BuildStatus{Status: model.BuildStatusNewLabel}

	// Get the project related to the Test / Build
	project, err := m.Project(projectId)
	if err != nil && err != ErrProjectDoesNotExist {
		return "", err
	}

	// Get the Test and its TargetArtifacts
	test, err := m.GetTest(projectId, testId)
	if err != nil && err != ErrTestDoesNotExist {
		return "", err
	}
	targetArtifacts := test.Targets

	// we get the ids for the targets we want to test
	targetIds := []string{}
	for _, target := range targetArtifacts {
		targetIds = append(targetIds, target.ID)

	}
	// Retrieve the images from the projectId
	// TODO: Investigate if we can query db for the images matching the Ids in the TargetArtifacts
	projectImages, err := m.GetImages(projectId)
	if err != nil && err != ErrProjectImagesProblem {
		return "", err
	}

	// Collect the images that are TargetArtifacts
	// by comparing the ImageID with the ArtifactId
	imagesToBuild := []*model.Image{}
	for _, image := range projectImages {
		for _, artifactId := range targetIds {
			if image.ID == artifactId {
				imagesToBuild = append(imagesToBuild, image)
			}
		}
	}

	// Store the Build in the table in rethink db
	response, err := r.Table(tblNameBuilds).Insert(build).RunWrite(m.session)
	if err != nil {
		return "", err
	}

	build.ID = func() string {
		if len(response.GeneratedKeys) > 0 {
			return string(response.GeneratedKeys[0])
		}
		return ""
	}()

	// Verify that the provider is valid
	provider, err := m.GetProvider(test.Provider.ID)

	if err != nil {
		log.Errorf("Error finding provider %s with id %s for test %s with id %s",
			test.Provider.Name, test.Provider.ID, test.Name, test.ID)
		return build.ID, err
	}

	found := false

	// Verify that the ProviderJob chosen is valid
	for _, providerJob := range provider.ProviderJobs {
		if providerJob.Name == test.ProviderJob.Name &&
			providerJob.Url == test.ProviderJob.Url {
			found = true
			break
		}
	}

	if !found {
		msg := fmt.Sprintf("Error matching provider job %s with url %s for provider %s with id %s for test %s with id %s",
			test.ProviderJob.Name, test.ProviderJob.Url, test.Provider.Name, test.Provider.ID, test.Name, test.ID)
		log.Error(msg)

		return build.ID, errors.New(msg)
	}

	tasks := []*model.ProviderTask{}

	for _, image := range imagesToBuild {
		log.Printf("Processing image=%s", image.PullableName())

		registry, err := m.Registry(image.RegistryId)
		if err != nil {
			log.Warnf("Could not find registry %s for image %s", image.RegistryId, image.ID)
		}

		// Each image build will be a task in the scope of a Test Build
		tasks = append(tasks, model.NewProviderTask(
			project,
			test,
			build,
			image,
			registry,
		))
	}

	// Create a Build that will be send to the appropriate Provider
	providerBuild := model.NewProviderBuild(test.ProviderJob, tasks)

	// Start a goroutine that will execute the build non-blocking
	go func(providerBuild *model.ProviderBuild, provider *model.Provider) {
		// Send Build request to provider
		provider.SendBuild(providerBuild)
	}(providerBuild, provider)

	// TODO: all these event types should be refactored as constants
	eventType = "add-build"
	m.logEvent(eventType, fmt.Sprintf("id=%s", build.ID), []string{"security"})
	return build.ID, nil
}

func (m DefaultManager) UpdateBuild(projectId string, testId string, buildId string, buildAction *model.BuildAction) error {
	var eventType string

	// check if exists; if so, update
	tmpBuild, err := m.GetBuild(buildId)
	if err != nil && err != ErrBuildDoesNotExist {
		return err
	}
	// update
	if tmpBuild != nil {
		if buildAction.Action == "stop" {
			tmpBuild.Status.Status = "stopped"
			tmpBuild.EndTime = time.Now()
			// go StopCurrentBuildFromClair
		}
		if buildAction.Action == "restart" {
			tmpBuild.Status.Status = "restarted"
			tmpBuild.EndTime = time.Now()
			// go RestartCurrentBuildFromClair

		}

		if _, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Update(tmpBuild).RunWrite(m.session); err != nil {
			return err
		}

		eventType = "update-build"
	}

	m.logEvent(eventType, fmt.Sprintf("id=%s", buildId), []string{"security"})

	return nil

}

func (m DefaultManager) UpdateBuildResults(buildId string, result model.BuildResult) error {
	var eventType string
	build, err := m.GetBuild(buildId)
	if err != nil {
		return err
	}
	build.Results = append(build.Results, &result)

	if _, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Update(build).RunWrite(m.session); err != nil {
		return err
	}

	eventType = "update-build-results"

	m.logEvent(eventType, fmt.Sprintf("id=%s", buildId), []string{"security"})

	return nil
}
func (m DefaultManager) UpdateBuildStatus(buildId string, status string) error {
	var eventType string
	build, err := m.GetBuild(buildId)
	if err != nil {
		return err
	}
	build.Status.Status = status

	if _, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Update(build).RunWrite(m.session); err != nil {
		return err
	}

	eventType = "update-build-status"

	m.logEvent(eventType, fmt.Sprintf("id=%s", buildId), []string{"security"})

	return nil
}
func (m DefaultManager) DeleteBuild(projectId string, testId string, buildId string) error {
	build, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Delete().Run(m.session)
	if err != nil {
		return err
	}

	if build.IsNil() {
		return ErrBuildDoesNotExist
	}

	m.logEvent("delete-build", fmt.Sprintf("id=%s", buildId), []string{"security"})

	return nil
}

func (m DefaultManager) DeleteAllBuilds() error {
	_, err := r.Table(tblNameBuilds).Delete().Run(m.session)

	if err != nil {
		return err
	}

	return nil
}
