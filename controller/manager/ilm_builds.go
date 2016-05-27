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

//type executeBuildTasksResults struct {
//	buildStatus string
//	buildResult *model.BuildResult
//}

//methods related to the Build structure
func (m DefaultManager) GetBuildsByTestId(testId string) ([]*model.Build, error) {
	res, err := r.Table(tblNameBuilds).Filter(map[string]string{"testId": testId}).Run(m.session)
	defer res.Close()
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
	log.Info("looking for build in db")
	res, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Run(m.session)
	defer res.Close()
	log.Info("done looking in db")
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
	log.Info("return build")
	return build, nil
}

func (m DefaultManager) GetBuildStatus(buildId string) (string, error) {
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
func (m DefaultManager) ExecuteBuild(testId string, buildAction *model.BuildAction) (string, error) {

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

	// Get the Test and its TargetArtifacts
	test, err := m.GetTest(testId)
	if err != nil && err != ErrTestDoesNotExist {
		return "", err
	}
	targetArtifacts := test.Targets

	// we get the ids for the targets we want to test
	targetIds := []string{}
	for _, target := range targetArtifacts {
		targetIds = append(targetIds, target.ID)

	}

	// Get the project related to the Test / Build
	// TODO: reconsider whether we need this or not
	project, err := m.Project(test.ProjectId)
	if err != nil && err != ErrProjectDoesNotExist {
		return "", err
	}

	// Retrieve the images from the projectId
	// TODO: Investigate if we can query db for the images matching the Ids in the TargetArtifacts
	projectImages, err := m.GetImages(test.ProjectId)
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

	log.Printf("Processing %d image(s)", len(imagesToBuild))

	tasks := []*model.BuildArtifact{}

	for _, image := range imagesToBuild {
		log.Printf("Processing image=%s", image.PullableName())

		registry, err := m.Registry(image.RegistryId)
		if err != nil {
			log.Warnf("Could not find registry %s for image %s", image.RegistryId, image.ID)
		}

		// Each image build will be a task in the scope of a Test Build
		tasks = append(tasks, model.NewBuildArtifact(
			image,
			registry,
		))
	}

	// Create a Build that will be send to the appropriate Provider
	providerBuild := model.NewProviderBuild(
		test.ProviderJob,
		project,
		test,
		build,
		tasks,
	)

	// Start a goroutine that will execute the build non-blocking
	go func(providerBuild *model.ProviderBuild, provider *model.Provider) {
		// Send Build request to provider
		provider.SendBuild(providerBuild)

		//// TODO: this is a mock to form a NewBuildResult
		//err := m.UpdateBuildResults(
		//	build.ID,
		//	model.NewBuildResult(
		//		build.ID,
		//		targetArtifacts[0],
		//		[]string{"one result entry", "two result entries"},
		//		true,
		//	),
		//)
		//
		//if err != nil {
		//	log.Errorf("There was an error updating build results %s", err.Error())
		//}
	}(providerBuild, provider)

	// Run build for each image (each image represents a build task)
	//go m.executeBuildTasks(project, test, build, imagesToBuild)

	// TODO: all these event types should be refactored as constants
	eventType = "add-build"
	m.logEvent(eventType, fmt.Sprintf("id=%s", build.ID), []string{"security"})
	return build.ID, nil
}

//func (m DefaultManager) executeBuildTasks(
//	project *model.Project,
//	test *model.Test,
//	build *model.Build,
//	imagesToBuild []*model.Image,
//) {
//	var taskReceivingChannels []<-chan executeBuildTasksResults
//	// For each image that we target in the test, try to run a build / verification
//	// TODO: Add "running", "start", etc... as constants somewhere
//	m.UpdateBuildStatus(build.ID, "running")
//	for _, image := range imagesToBuild {
//		log.Printf("Processing image=%s", image.PullableName())
//		channel := m.executeBuildTask(*project, *test, *build, *image)
//		taskReceivingChannels = append(taskReceivingChannels, channel)
//	}
//
//	var statuses []string
//	for _, taskReceivingChannel := range taskReceivingChannels {
//		results := <-taskReceivingChannel
//		// Fetch the build statuses and store them for later use
//		statuses = append(statuses, results.buildStatus)
//		// Update build results
//		m.UpdateBuildResults(build.ID, results.buildResult)
//	}
//	log.Info("all channels received")
//	// Check to see if build was successful by comparing the statuses for all build tasks
//	buildStatus := "finished_success"
//	for _, status := range statuses {
//		if status != "finished_success" {
//			buildStatus = "finished_failed"
//			break
//		}
//	}
//	// Update status for our build model
//	if err := m.UpdateBuildStatus(build.ID, buildStatus); err != nil {
//		log.Error(err)
//	}
//}
//
//// Generator that executes a BuilTask in the background as part of a wait group.
//// TODO: We should probably remove the `name` param as we already have the corresponding image object
//// TODO: Remove all *updates* from this method as they may cause the go routines to share address space
//func (m DefaultManager) executeBuildTask(
//	project model.Project,
//	test model.Test,
//	build model.Build,
//	image model.Image,
//) <-chan executeBuildTasksResults {
//	channel := make(chan executeBuildTasksResults)
//
//	go func() {
//		// TODO: need to revisit API spec, there are just too many redundant "Result" types stored,
//		// TODO: these should probably just be views of the BuildResults
//		// TODO: need to set the author to the real user
//		// TODO: use model.NewResult() instead
//		//result := &model.Result{
//		//	BuildId:     build.ID,
//		//	Author:      "author",
//		//	ProjectId:   project.ID,
//		//	Description: project.Description,
//		//	Updater:     "author",
//		//	CreateDate:  time.Now(),
//		//}
//
//		// TODO: use model.NewTestResult() instead
//		testResult := model.TestResult{}
//		testResult.Date = time.Now()
//		testResult.TestId = test.ID
//		testResult.BuildId = build.ID
//		testResult.TestName = test.Name
//		testResult.ImageName = image.PullableName()
//		testResult.BuildId = build.ID
//
//		// Check to see if the image exists locally, if not, try to pull it.
//		if !m.VerifyIfImageExistsLocally(image) {
//			log.Printf("Image %s not available locally, will try to pull...", image.PullableName())
//			if err := m.PullImage(image); err != nil {
//				log.Errorf("Error pulling image %s", image.PullableName())
//				return
//			}
//		}
//
//		// Get all local images
//		localImages, _ := apiClient.GetLocalImages(m.DockerClient().URL.String())
//
//		// TODO: Refactor this into its own func
//		// get the docker image id and append it to the test results
//		for _, localImage := range localImages {
//			imageRepoTags := localImage.RepoTags
//			for _, imageRepoTag := range imageRepoTags {
//				if imageRepoTag == image.PullableName() {
//					//image.DockerImageId = localImage.ID
//					testResult.DockerImageId = localImage.ID
//					image.ImageId = localImage.ID
//				}
//			}
//		}
//
//		//existingResult, _ := m.GetResults(project.ID)
//
//		// Once the image is available, try to test it with Clair
//		//log.Printf("Will attempt to test image %s with Clair...", image.PullableName())
//		//resultsSlice, isSafe, err := c.CheckImage(&image)
//		//
//		//targetArtifact := model.NewTargetArtifact(
//		//	image.ID,
//		//	model.TargetArtifactImageType,
//		//	image,
//		//)
//		//buildResult := model.NewBuildResult(build.ID, targetArtifact, resultsSlice)
//		//buildResult.TimeStamp = time.Now()
//		//
//		//finishLabel := "finished_failed"
//		//
//		//appliedTag := ""
//		//if isSafe && err == nil {
//		//	// if we don't get an error and we get the isSafe flag == true
//		//	// we mark the test for the image as successful
//		//	finishLabel = "finished_success"
//		//	// if the test is successful, we update the images' ilm tags with the test tags we defined in the case of a success
//		//	appliedTag = test.Tagging.OnSuccess
//		//	log.Infof("Image %s is safe! :)", image.PullableName())
//		//} else {
//		//	// if the test is failed, we update the images' ilm tags with the test tags we defined in the case of a failure
//		//	appliedTag = test.Tagging.OnFailure
//		//	log.Errorf("Image %s is NOT safe :(", image.PullableName())
//		//}
//		//if appliedTag != "" {
//		//	m.UpdateImageIlmTags(project.ID, image.ID, appliedTag)
//		//}
//		//
//		//
//		//testResult.SimpleResult.Status = finishLabel
//		//testResult.EndDate = time.Now()
//		//testResult.Blocker = false
//		//testResult.AppliedTag = append(testResult.AppliedTag, appliedTag)
//		//result.TestResults = append(result.TestResults, &testResult)
//		//result.LastUpdate = time.Now()
//		//result.LastTagApplied = appliedTag
//		//
//		//if existingResult != nil {
//		//	m.UpdateResult(project.ID, result)
//		//} else {
//		//	m.CreateResult(project.ID, result)
//		//}
//		//
//		//log.Info("sending channel")
//		//channel <- executeBuildTasksResults{
//		//	buildStatus: finishLabel,
//		//	buildResult: buildResult,
//		//}
//		//log.Info("sent channel")
//	}()
//
//	return channel
//}

func (m DefaultManager) UpdateBuild(buildId string, buildAction *model.BuildAction) error {
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

func (m DefaultManager) UpdateBuildResults(buildId string, result *model.BuildResult) error {
	var eventType string

	_, err := r.Table(tblNameBuilds).Get(buildId).Update(
		map[string]interface{}{
			"results": r.Row.Field("results").Append(result),
		}).RunWrite(m.session)

	if err != nil {
		return err
	}

	eventType = "update-build-results"

	m.logEvent(eventType, fmt.Sprintf("id=%s", buildId), []string{"security"})

	return nil
}
func (m DefaultManager) UpdateBuildStatus(build *model.Build, buildStatus *model.BuildStatus) error {

	status := buildStatus.Status
	switch status {
	case model.BuildStatusFinishedFailed:
	case model.BuildStatusFinishedSuccess:
	case model.BuildStatusRunningLabel:
	case model.BuildStatusStoppedLabel:
	case model.BuildStatusNewLabel:
		msg := fmt.Sprintf("Received acceptable status %s",status)
		log.Info(msg)
	default:
		msg := fmt.Sprintf("Received a status that is not acceptable %s",status)
		log.Error(msg)
		return errors.New(msg)
	}
	_, err := r.Table(tblNameBuilds).Get(build.ID).Update(
		map[string]interface{}{
			"status": buildStatus,
			},
		).RunWrite(m.session)

	if err != nil {
		log.Error("Could not set the build status go %s for build with id %s",status, build.ID)
			return err
		}

	eventType := "update-build-status"

	m.logEvent(eventType, fmt.Sprintf("id=%s", build), []string{"security"})

	return nil
}
func (m DefaultManager) DeleteBuild(buildId string) error {
	build, err := r.Table(tblNameBuilds).Filter(map[string]string{"id": buildId}).Delete().Run(m.session)
	defer build.Close()
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
	res, err := r.Table(tblNameBuilds).Delete().Run(m.session)

	defer res.Close()
	if err != nil {
		return err
	}

	return nil
}
