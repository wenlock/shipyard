package health_checker

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	"github.com/samalba/dockerclient"
	"github.com/shipyard/shipyard/model"
	"net/http"
	"time"
)

var (
	ErrNoDbListAvailable = errors.New("Could not list databases in health checker")
)

const (
	WaitDuration = 5 * time.Second
	MaxAttempts  = 5
// TODO: expose these constants in the manager package to avoid duplicates
	tblNameProviders = "providers"
// TODO: improve these statuses
	HealthCheckStatusNotOk = "NOT OK"
)

// Health Checker object
type HealthChecker struct {
	dbAddress    string
	databaseName string
	dbAuthKey    string
	delay        time.Duration
	session      *r.Session
	attempts     int
	dockerClient *dockerclient.DockerClient
}

// Instantiate a new HealthChecker object
func NewHealthChecker(
dbAddress string,
databaseName string,
dbAuthKey string,
delay time.Duration,
dockerClient *dockerclient.DockerClient,
) (*HealthChecker, error) {

	hc := &HealthChecker{
		dbAddress:    dbAddress,
		databaseName: databaseName,
		dbAuthKey:    dbAuthKey,
		delay:        delay,
		dockerClient: dockerClient,
		attempts:     0,
	}
	return hc, nil
}

// An object that will allow us to bind the health response from a provider
type Health struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
}

// Start the health checker in the background
func (hc *HealthChecker) Start() {
	go start(hc)
}

// Performs health checks for all providers and services
func (hc *HealthChecker) checkAll() {
	log.Infof("Performing all health checks...")

	hc.checkDockerDaemon()

	hc.checkProviders()
}

// Performs health checks on all providers registered in the ILM
func (hc *HealthChecker) checkProviders() {
	response, err := r.Table(tblNameProviders).Run(hc.session)

	if err != nil {
		log.Warn("Error getting all providers from the db")
		return
	}

	providers := []*model.Provider{}

	if err := response.All(&providers); err != nil {
		log.Warn("Could not get provider list from db")
		return
	}

	for _, provider := range providers {
		providerString := fmt.Sprintf("name = %s id = %s, health check url = %s", provider.Name, provider.ID, provider.Health.Url)
		log.Infof("Executing healthcheck for provider %s", providerString)
		if provider.Health.Url == "" {
			log.Warnln("No health check url provided, skipping...")
			continue
		}

		healthResponse, err := http.Get(provider.Health.Url)

		if err != nil {
			log.Warnf("Received an error in health check request %s", err.Error())
			continue
		}

		prefix := fmt.Sprintf("Provider %s", providerString)
		if healthResponse.StatusCode == http.StatusOK {
			var health *Health
			err := json.NewDecoder(healthResponse.Body).Decode(&health)

			if err != nil {
				log.Warnln("Could not unmarshall response")
				continue
			}
			log.Infoln(prefix, "health check OK :)")
			log.Infof("Response = %s with code = %d", health.Status, health.StatusCode)

			// Mark health as OK
			provider.Health.Status = health.Status
		} else {
			// Mark health as NOT OK
			provider.Health.Status = HealthCheckStatusNotOk
			log.Warnln(prefix, "health check FAILRED :(")
		}

		// Update health time stamp
		provider.Health.LastUpdate = time.Now()

		// Update the provider health with previous results
		r.Table(tblNameProviders).Filter(
			map[string]string{"id": provider.ID},
		).Update(provider).Run(hc.session)
	}
}

func (hc *HealthChecker) checkDockerDaemon() {
	_, err := hc.dockerClient.ListContainers(true, false, "")

	msg := "Health check = docker daemon SUCCESS :)"
	if err != nil {
		msg = "Health check = docker daemon FAILED :("
	}
	log.Infof(msg)
}

// Starts a health check routine for the passed HealthChecker object
func start(hc *HealthChecker) {
	session, err := r.Connect(r.ConnectOpts{
		Address:  hc.dbAddress,
		Database: hc.databaseName,
		AuthKey:  hc.dbAuthKey,
		MaxIdle:  10,
		MaxOpen: 20,
	})
	if err != nil {
		return
	}
	log.Info("checking database")
	hc.session = session

	// Get the list of all databases
	found := false
	for {
		if hc.attempts == MaxAttempts {
			break
		}
		hc.attempts += hc.attempts
		log.Infof("getting database list from db server at %s", hc.dbAddress)
		resp, err := r.DBList().Run(hc.session)

		if err == nil && !resp.IsNil() {
			databaseList := []interface{}{}
			resp.All(&databaseList)

			// Look for database name
			for _, db := range databaseList {
				if db == hc.databaseName {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		log.Infof("Will try again to ping the database...waiting for %s", WaitDuration)
		time.Sleep(WaitDuration)
	}

	// Perform health checks
	if found {
		for {
			log.Infof("Performing health checks...")
			hc.checkAll()
			time.Sleep(hc.delay)
		}
	} else {
		log.Errorf("Could not find database %s", hc.databaseName)
	}
}