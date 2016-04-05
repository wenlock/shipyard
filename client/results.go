package ilm_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shipyard/shipyard/model"
	"io/ioutil"
	"net/http"
	"time"
)

func GetResults(authHeader, url string, projectId string) ([]model.Result, int, error) {
	var results []model.Result
	resp, err := sendRequest(authHeader, "GET", fmt.Sprintf("%s/api/projects/%s/results", url, projectId), "")
	if err != nil {
		return nil, resp.StatusCode, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal([]byte(body), &results)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return results, resp.StatusCode, nil
}

func CreateResult(authHeader string, url string, projectId string, description string, buildId string, author string, projectVersion string, lastTagApplied string, updater string) (int, error) {
	var result *model.Result
	timestamp := time.Now()
	result = result.NewResult(projectId, description, buildId, timestamp, timestamp, timestamp, author, projectVersion, lastTagApplied, timestamp, updater, nil)

	data, err := json.Marshal(result)
	if err != nil {
		return 0, err
	}
	resp, err := sendRequest(authHeader, "POST", fmt.Sprintf("%s/api/projects/%s/results", url, projectId), string(data))

	return resp.StatusCode, err

}

func GetResult(authHeader, url, projectId string, resultId string) (*model.Result, int, error) {
	var result *model.Result
	resp, err := sendRequest(authHeader, "GET", fmt.Sprintf("%s/api/projects/%s/results/%s", url, projectId, resultId), "")
	if err != nil {
		return result, resp.StatusCode, err
	}

	// If we get an error status code we should not try to unmarshall body, since it will come empty from server.
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, resp.StatusCode, err
	}

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return result, resp.StatusCode, errors.New("Error, could not unmarshall result body")
	}

	return result, resp.StatusCode, nil
}

func UpdateResult(authHeader string, url string, projectId string, resultId, description string, buildId string, author string, projectVersion string, lastTagApplied string, updater string) (int, error) {

	//create the project
	var result *model.Result
	var never time.Time //empty time stamp
	result = result.NewResult(projectId, description, buildId, never, never, never, author, projectVersion, lastTagApplied, never, updater, nil)
	result.ID = resultId
	data, err := json.Marshal(result)
	if err != nil {
		return 0, err
	}
	resp, err := sendRequest(authHeader, "PUT", fmt.Sprintf("%s/api/projects/%s/results/%s", url, projectId, resultId), string(data))
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

func DeleteResult(authHeader, url, projectId string, resultId string) (int, error) {
	resp, err := sendRequest(authHeader, "DELETE", fmt.Sprintf("%s/api/projects/%s/results/%s", url, projectId, resultId), "")
	return resp.StatusCode, err
}
