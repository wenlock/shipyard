package manager

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	"github.com/samalba/dockerclient"
	apiClient "github.com/shipyard/shipyard/client"
	"github.com/shipyard/shipyard/model"
	"time"
	"strings"
)

var (
	ErrImageWasEmpty       = errors.New("image name was empty, please pass a valid object")
	ErrDomainNoLongerValid = "ErrorDomainNoLongerValid"
)

// check if an image exists
func (m DefaultManager) VerifyIfImageExistsLocally(image model.Image) bool {
	images, err := apiClient.GetLocalImages(m.DockerClient().URL.String())

	if err != nil {
		log.Error(err)
		return false
	}

	for _, img := range images {
		imageRepoTags := img.RepoTags
		for _, imageRepoTag := range imageRepoTags {
			if imageRepoTag == image.PullableName() {
				fmt.Printf("Image %s exists locally as %s \n", image.PullableName(), imageRepoTag)
				return true
			}
		}
	}

	return false
}

// Checks to see if the given image is available locally.
// If it's not, it pulls the image (registry is defined by image.PullableName()).
// If it is, it skips the pull.
func (m DefaultManager) PullImage(image model.Image) error {

	// Check to see if the image exists locally, if not, try to pull it.
	if m.VerifyIfImageExistsLocally(image) {
		log.Debug("Image %s exists locally, will not try to pull...", image.PullableName())
		return nil
	}

	username := ""
	password := ""
	if image.RegistryId != "" {
		registry, err := m.Registry(image.RegistryId)
		if err != nil {
			log.Warnf("Could not find registry %s for image %s", image.RegistryId, image.ID)
		} else {
			username = registry.Username
			password = registry.Password
		}
	}

	auth := dockerclient.AuthConfig{username, password, ""}

	ticker := time.NewTicker(time.Second * 15)
	go func() {
		for t := range ticker.C {
			fmt.Print("Time: ", t.UTC())
			fmt.Printf(" Pulling image: %s. Please be patient while the process finishes ... \n", image.PullableName())
		}
	}()

	// TODO: stop using samalba/dockerclient, use Docker, Inc docker engine client library instead
	err := m.client.PullImage(image.PullableName(), &auth)

	if err != nil {
		fmt.Printf("Could not pull image %s ... \n %s \n", image.PullableName(), err)
		ticker.Stop()
		return err
	}
	ticker.Stop()

	return nil
}

//methods related to the Image structure
func (m DefaultManager) GetImages(projectId string) ([]*model.Image, error) {

	res, err := r.Table(tblNameImages).Filter(map[string]string{"projectId": projectId}).Run(m.session)
	defer res.Close()
	if err != nil {
		return nil, err
	}
	images := []*model.Image{}
	if err := res.All(&images); err != nil {
		return nil, err
	}

	for _, image := range images {
		m.injectRegistryInfo(image)
	}
	return images, nil
}

func (m DefaultManager) GetImage(projectId string, imageId string) (*model.Image, error) {
	var image *model.Image
	res, err := r.Table(tblNameImages).Filter(map[string]string{"id": imageId}).Run(m.session)
	defer res.Close()
	if err != nil {
		return nil, err
	}
	if res.IsNil() {
		return nil, ErrImageDoesNotExist
	}
	if err := res.One(&image); err != nil {
		return nil, err
	}

	m.injectRegistryInfo(image)
	return image, nil
}

func (m DefaultManager) injectRegistryInfo(image *model.Image) {
	if image.RegistryId != "" {
		registry, err := m.Registry(image.RegistryId)
		domain := ErrDomainNoLongerValid
		if err == nil {
			domain = formatRegistryDomain(registry.Addr)
		}
		image.RegistryDomain = domain
	}
}

func (m DefaultManager) CreateImage(projectId string, image *model.Image) error {
	var eventType string
	image.ProjectId = projectId

	m.injectRegistryInfo(image)
	response, err := r.Table(tblNameImages).Insert(image).RunWrite(m.session)
	if err != nil {

		return err
	}
	image.ID = func() string {
		if len(response.GeneratedKeys) > 0 {
			return string(response.GeneratedKeys[0])
		}
		return ""
	}()
	eventType = "add-image"

	m.logEvent(eventType, fmt.Sprintf("id=%s", image.ID), []string{"security"})
	return nil
}

func (m DefaultManager) UpdateImage(projectId string, image *model.Image) error {
	var eventType string
	// check if exists; if so, update
	rez, err := m.GetImage(projectId, image.ID)
	if err != nil && err != ErrImageDoesNotExist {
		return err
	}
	// update
	if rez == nil {
		return ErrImageDoesNotExist
	}

	m.injectRegistryInfo(image)

	// Convert struct to map and refrain from doing this manually
	updates := map[string]interface{}{
		"name":           image.Name,
		"imageId":        image.ImageId,
		"tag":            image.Tag,
		"ilmTags":        image.IlmTags,
		"description":    image.Description,
		"registryId":     image.RegistryId,
		"location":       image.Location,
		"skipImageBuild": image.SkipImageBuild,
		"projectId":      image.ProjectId,
		"registryDomain": image.RegistryDomain,
	}
	if _, err := r.Table(tblNameImages).Filter(map[string]string{"id": image.ID}).Update(updates).RunWrite(m.session); err != nil {
		return err
	}

	eventType = "update-image"

	m.logEvent(eventType, fmt.Sprintf("id=%s", image.ID), []string{"security"})
	return nil
}

// Sets the on_success|on_failure tag for the given ilm image and
// performs a `docker tag` to reflect the changes
func (m DefaultManager) UpdateImageIlmTags(projectId string, imageId string, ilmTag string) error {
	// check if exists; if so, update
	image, err := m.GetImage(projectId, imageId)
	if err != nil && err != ErrImageDoesNotExist {
		return err
	}

	if image == nil {
		return ErrImageDoesNotExist
	}

	// Set the ilm tag to the given image
	image.IlmTags = []string{ilmTag}

	// TODO: cleanup of old tags
	// Tag the image. Analogous to `docker tag`
	imagePullableFormat := image.PullableName()
	imageRepoFormat := strings.Replace(
		imagePullableFormat,
		fmt.Sprintf(":%s", image.Tag),
		"",
		1,
	)
	err = m.client.TagImage(
		imagePullableFormat,
		imageRepoFormat,
		ilmTag,
		true,
	)

	if err != nil {
		log.Debugf("Could not apply success tag (%s) to image %s", ilmTag, image.PullableName())
		return err
	}

	if _, err := r.Table(tblNameImages).Filter(map[string]string{"id": imageId}).Update(image).RunWrite(m.session); err != nil {
		return err
	}

	return nil
}

func (m DefaultManager) DeleteImage(projectId string, imageId string) error {
	res, err := r.Table(tblNameImages).Filter(map[string]string{"id": imageId}).Delete().Run(m.session)
	defer res.Close()
	if err != nil {
		return err
	}
	if res.IsNil() {
		return ErrImageDoesNotExist
	}

	m.logEvent("delete-image", fmt.Sprintf("id=%s", imageId), []string{"security"})

	return nil
}

func (m DefaultManager) DeleteAllImages() error {
	res, err := r.Table(tblNameImages).Delete().Run(m.session)
	defer res.Close()
	if err != nil {
		return err
	}

	return nil
}
