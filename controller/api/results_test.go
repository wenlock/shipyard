package api

import (
	"fmt"
	"github.com/gorilla/context"
	apiClient "github.com/shipyard/shipyard/client"
	"github.com/shipyard/shipyard/model"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	RESULT1_DESC    = "the first result"
	RESULT2_DESC    = "the second result"
	RESULT1_BUILDID = "build 1"
	RESULT2_BUILDID = "build 2"
	RESULT1_AUTHOR  = "author 1"
	RESULT2_AUTHOR  = "author 2"
	RESULT1_VERSION = "version 1"
	RESULT2_VERSION = "version 2"
	RESULT1_TAG     = "tag 1"
	RESULT2_TAG     = "tag 2"
	RESULT1_UPDATER = "updater 1"
	RESULT2_UPDATER = "updater 2"
	// for updating a result
	RESULT1_DESC2    = "the first result v2"
	RESULT1_BUILDID2 = "build 1 v2"
	RESULT1_AUTHOR2  = "author 1 v2"
	RESULT1_VERSION2 = "version 1 v2"
	RESULT1_TAG2     = "tag 1 v2"
	RESULT1_UPDATER2 = "updater 1 v2"
)

var (
	PROJECT_WITH_RESULTS_1_SAVED_ID string
	result1Object                   *model.Result
	result2Object                   *model.Result
)

func init() {
	dockerEndpoint := os.Getenv("SHIPYARD_DOCKER_URI")

	// Default docker endpoint
	if dockerEndpoint == "" {
		dockerEndpoint = "tcp://127.0.0.1:2375"
	}

	rethinkDbEndpoint := os.Getenv("SHIPYARD_RETHINKDB_URI")

	// Default rethinkdb endpoint
	if rethinkDbEndpoint == "" {
		rethinkDbEndpoint = "rethinkdb:28015"
	}

	localApi, localMux, err := InitServer(&ShipyardServerConfig{
		RethinkdbAddr:          rethinkDbEndpoint,
		RethinkdbAuthKey:       "",
		RethinkdbDatabase:      "shipyard_test",
		DisableUsageInfo:       true,
		ListenAddr:             "",
		AuthWhitelist:          []string{},
		EnableCors:             true,
		LdapServer:             "",
		LdapPort:               389,
		LdapBaseDn:             "",
		LdapAutocreateUsers:    true,
		LdapDefaultAccessLevel: "containers:ro",
		DockerUrl:              dockerEndpoint,
		TlsCaCert:              "",
		TlsCert:                "",
		TlsKey:                 "",
		AllowInsecure:          true,
		ShipyardTlsCert:        "",
		ShipyardTlsKey:         "",
		ShipyardTlsCACert:      "",
	})

	if err != nil {
		panic(fmt.Sprintf("Test init() for results_test.go failed %s", err))
	}

	api = localApi
	globalMux = localMux

	cleanupResults()

	// Instantiate test server with Gorilla Mux Router enabled.
	// If you don't wrap the mux with the context.ClearHandler(),
	// then the server request cycle won't go through GorillaMux routing.
	ts = httptest.NewServer(context.ClearHandler(globalMux))

}

// TODO: this is snot cleaning up the tokens
func cleanupResults() error {

	if err := api.manager.DeleteAllResults(); err != nil {
		return err
	}

	if err := api.manager.DeleteAllProjects(); err != nil {
		return err
	}

	return nil
}

func TestResultsGetAuthToken(t *testing.T) {

	Convey("Given a valid set of credentials", t, func() {
		Convey("When we make a successful request for an auth token", func() {
			header, err := apiClient.GetAuthToken(ts.URL, SYUSER, SYPASS)
			So(err, ShouldBeNil)

			Convey("Then we get a valid authentication header\n", func() {
				SY_AUTHTOKEN = header
				So(header, ShouldNotBeEmpty)
				numberOfParts := 2
				authToken := strings.SplitN(header, ":", numberOfParts)
				So(len(authToken), ShouldEqual, numberOfParts)
				So(authToken[0], ShouldEqual, SYUSER)
			})
		})

	})
}

func TestCreateNewResult(t *testing.T) {
	Convey("Given that we have a valid token and a valid project", t, func() {
		So(SY_AUTHTOKEN, ShouldNotBeNil)
		So(SY_AUTHTOKEN, ShouldNotBeEmpty)
		id, code, err := apiClient.CreateProject(SY_AUTHTOKEN, ts.URL, PROJECT1_NAME, PROJECT1_DESC, PROJECT1_STATUS, nil, nil, false)

		So(err, ShouldBeNil)
		So(code, ShouldEqual, http.StatusCreated)
		So(id, ShouldNotBeEmpty)
		PROJECT_WITH_RESULTS_1_SAVED_ID = id
		Convey("When we make a request to create a new result", func() {
			code, err := apiClient.CreateResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, RESULT1_DESC, RESULT1_BUILDID, RESULT1_AUTHOR, RESULT1_VERSION, RESULT1_TAG, RESULT1_UPDATER)
			Convey("Then we get back a successful response", func() {
				So(err, ShouldBeNil)
				So(code, ShouldEqual, http.StatusCreated)
			})
		})
	})
}

func TestGetAllResults(t *testing.T) {
	Convey("Given that we have created an additional project", t, func() {
		So(SY_AUTHTOKEN, ShouldNotBeNil)
		So(SY_AUTHTOKEN, ShouldNotBeEmpty)
		code, err := apiClient.CreateResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, RESULT2_DESC, RESULT2_BUILDID, RESULT2_AUTHOR, RESULT2_VERSION, RESULT2_TAG, RESULT2_UPDATER)
		So(err, ShouldBeNil)
		So(code, ShouldEqual, http.StatusCreated)
		Convey("When we make a request to retrieve all results", func() {
			results, code, err := apiClient.GetResults(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID)
			Convey("Then the request should return some objects", func() {
				So(err, ShouldBeNil)
				So(code, ShouldEqual, http.StatusOK)
				So(results, ShouldNotBeNil)
				So(len(results), ShouldEqual, 2)
				Convey("And the objects return should have the expected structure and expected values.", func() {
					descriptions := []string{}
					buildids := []string{}
					authors := []string{}
					updaters := []string{}
					versions := []string{}
					tags := []string{}
					ids := []string{}

					for _, result := range results {
						descriptions = append(descriptions, result.Description)
						buildids = append(buildids, result.BuildId)
						authors = append(authors, result.Author)
						updaters = append(updaters, result.Updater)
						versions = append(versions, result.ProjectVersion)
						tags = append(tags, result.LastTagApplied)
						ids = append(ids, result.ID)
						So(result.ID, ShouldNotBeNil)
						So(result.ID, ShouldNotBeEmpty)
						So(result.CreateDate, ShouldNotBeEmpty)
						So(result.LastUpdate, ShouldNotBeEmpty)
						So(result.RunDate, ShouldNotBeEmpty)
						So(result.EndDate, ShouldNotBeEmpty)
					}

					So(RESULT1_DESC, ShouldBeIn, descriptions)
					So(RESULT2_DESC, ShouldBeIn, descriptions)
					So(RESULT1_BUILDID, ShouldBeIn, buildids)
					So(RESULT2_BUILDID, ShouldBeIn, buildids)
					So(RESULT1_AUTHOR, ShouldBeIn, authors)
					So(RESULT2_AUTHOR, ShouldBeIn, authors)
					So(RESULT1_UPDATER, ShouldBeIn, updaters)
					So(RESULT2_UPDATER, ShouldBeIn, updaters)
					So(RESULT1_VERSION, ShouldBeIn, versions)
					So(RESULT2_VERSION, ShouldBeIn, versions)
					So(RESULT1_TAG, ShouldBeIn, tags)
					So(RESULT2_TAG, ShouldBeIn, tags)

					result1Object = &results[0]
					result2Object = &results[1]
				})
			})

		})

	})
}

//pull the test and make sure it is exactly as it was ordered to be created
func TestGetResult(t *testing.T) {
	Convey("Given that we have a valid result and a valid token", t, func() {
		So(SY_AUTHTOKEN, ShouldNotBeNil)
		So(SY_AUTHTOKEN, ShouldNotBeEmpty)
		So(result1Object, ShouldNotBeNil)
		So(result1Object.ID, ShouldNotBeEmpty)

		Convey("When we make a request to retrieve it using its id", func() {
			result, code, err := apiClient.GetResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID)
			Convey("Then the server should return OK", func() {
				So(err, ShouldBeNil)
				So(code, ShouldEqual, http.StatusOK)
				Convey("Then the result project should have the expected values", func() {
					So(result.ID, ShouldEqual, result1Object.ID)
					So(result.Description, ShouldEqual, result1Object.Description)
					So(result.BuildId, ShouldEqual, result1Object.BuildId)
					So(result.Author, ShouldEqual, result1Object.Author)
					So(result.ProjectVersion, ShouldEqual, result1Object.ProjectVersion)
					So(result.LastTagApplied, ShouldEqual, result1Object.LastTagApplied)
					So(result.Updater, ShouldEqual, result1Object.Updater)
				})
			})

		})
	})
}

func TestUpdateResult(t *testing.T) {
	Convey("Given that we have a result created already.", t, func() {
		Convey("When we request to update that result.", func() {
			code, err := apiClient.UpdateResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID, RESULT1_DESC2, RESULT1_BUILDID2, RESULT1_AUTHOR2, RESULT1_VERSION2, RESULT1_TAG2, RESULT1_UPDATER2)
			Convey("Then we get back a successful response", func() {
				Convey("Then we get an appropriate response back", func() {
					So(err, ShouldBeNil)
					So(code, ShouldEqual, http.StatusNoContent)
					Convey("And when we retrieve the result again, it has the modified values.", func() {
						result, code, err := apiClient.GetResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID)
						So(err, ShouldBeNil)
						So(code, ShouldEqual, http.StatusOK)
						So(result.ID, ShouldEqual, result1Object.ID)
						So(result.Description, ShouldEqual, RESULT1_DESC2)
						So(result.BuildId, ShouldEqual, RESULT1_BUILDID2)
						So(result.Author, ShouldEqual, RESULT1_AUTHOR2)
						So(result.ProjectVersion, ShouldEqual, RESULT1_VERSION2)
						So(result.LastTagApplied, ShouldEqual, RESULT1_TAG2)
						So(result.Updater, ShouldEqual, RESULT1_UPDATER2)
					})
				})
			})
		})
	})
}

func TestDeleteResult(t *testing.T) {
	Convey("Given that we have a result created already.", t, func() {
		Convey("When we request to delete the result", func() {
			//delete the second project
			code, err := apiClient.DeleteResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID)
			Convey("Then we get confirmation that it was deleted.", func() {
				So(err, ShouldBeNil)
				So(code, ShouldEqual, http.StatusNoContent)
				Convey("And if we try to retrieve the project again by its id it should fail.", func() {
					//try to get the second project and make sure the server sends an error
					_, code, err = apiClient.GetResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID)
					So(err, ShouldBeNil)
					So(code, ShouldEqual, http.StatusNotFound)
					Convey("And if we get all projects, it should not be in the collection.", func() {
						results, code, err := apiClient.GetResults(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID)
						So(err, ShouldBeNil)
						So(code, ShouldEqual, http.StatusOK)
						So(results, ShouldNotBeNil)
						So(len(results), ShouldEqual, 1)
						descriptions := []string{}
						buildids := []string{}
						authors := []string{}
						updaters := []string{}
						versions := []string{}
						tags := []string{}
						ids := []string{}

						for _, result := range results {
							descriptions = append(descriptions, result.Description)
							buildids = append(buildids, result.BuildId)
							authors = append(authors, result.Author)
							updaters = append(updaters, result.Updater)
							versions = append(versions, result.ProjectVersion)
							tags = append(tags, result.LastTagApplied)
							ids = append(ids, result.ID)
							So(result.ID, ShouldNotBeNil)
							So(result.ID, ShouldNotBeEmpty)
							So(result.CreateDate, ShouldNotBeEmpty)
							So(result.LastUpdate, ShouldNotBeEmpty)
							So(result.RunDate, ShouldNotBeEmpty)
							So(result.EndDate, ShouldNotBeEmpty)
						}

						So(result1Object.ID, ShouldNotBeIn, ids)
						So(result2Object.ID, ShouldBeIn, ids)

					})
				})
			})
		})
	})

}

func TestResultNotFoundScenarios(t *testing.T) {
	cleanupResults()
	Convey("Given that a result with a given id does not exist", t, func() {
		Convey("When we try to retrieve that result by its id", func() {
			result, code, err := apiClient.GetResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID)
			Convey("Then we should get a not found error", func() {
				So(result, ShouldBeNil)
				So(code, ShouldEqual, http.StatusNotFound)
				So(err, ShouldBeNil)
			})
		})
		Convey("When we try to delete that result by its id", func() {
			code, err := apiClient.DeleteResult(SY_AUTHTOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, result1Object.ID)
			Convey("Then we should get a not found error", func() {
				So(code, ShouldEqual, http.StatusNotFound)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGetAuthTokenWithInvalidCredentialsForResults(t *testing.T) {
	Convey("Given that we have invalid credentials", t, func() {
		Convey("When we try to request an auth token", func() {
			token, err := apiClient.GetAuthToken(ts.URL, INVALID_USERNAME, INVALID_PASSWORD)
			Convey("Then we should get an error", func() {
				So(err, ShouldNotBeNil)
				Convey("And response should not contain any token", func() {
					So(token, ShouldBeBlank)
				})
			})
		})
	})
}

func TestUnauthorizedResultRequests(t *testing.T) {
	Convey("Given that we don't have a valid token", t, func() {
		Convey("When we try to get all projects", func() {
			results, code, err := apiClient.GetResults(INVALID_AUTH_TOKEN, ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID)
			Convey("Then we should be denied access", func() {
				So(code, ShouldEqual, http.StatusUnauthorized)
				So(err, ShouldNotBeNil)
				Convey("And we should not get anything in return", func() {
					So(results, ShouldBeNil)
				})
			})
		})
	})
	Convey("Given that we have an empty token", t, func() {
		Convey("When we request to create a new result", func() {
			code, _ := apiClient.CreateResult("", ts.URL, PROJECT_WITH_RESULTS_1_SAVED_ID, RESULT2_DESC, RESULT2_BUILDID, RESULT2_AUTHOR, RESULT2_VERSION, RESULT2_TAG, RESULT2_UPDATER)
			Convey("Then we should be denied access", func() {
				So(code, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}

// TODO: Add functionality for manager to close the database session
func TestCleanupResultsTests(t *testing.T) {
	// Cleanup all the state in the database
	Convey("Given that we have finished our project test suite", t, func() {
		Convey("Then we can cleanup", func() {
			err := cleanupResults()
			So(err, ShouldBeNil)
		})
	})

}
