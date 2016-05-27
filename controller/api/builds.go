package api

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/shipyard/shipyard/model"
	"net/http"
)

func (a *Api) getBuildsByTestId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	testId := vars["testId"]
	builds, err := a.manager.GetBuildsByTestId(testId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(builds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) getBuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildId := vars["buildId"]

	build, err := a.manager.GetBuild(buildId)
	if err != nil {
		log.Errorf("error retrieving build: %s", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(build); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (a *Api) getBuildStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildId := vars["buildId"]

	buildStatus, err := a.manager.GetBuildStatus(buildId)
	if err != nil {
		log.Errorf("error retrieving build status: %s", err)
		return
	}

	if err := json.NewEncoder(w).Encode(buildStatus); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (a *Api) getBuildResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildId := vars["buildId"]

	buildResults, err := a.manager.GetBuildResults(buildId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info("Was able to get results in controller, now trying json marshalling")
	if err := json.NewEncoder(w).Encode(&buildResults); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (a *Api) executeBuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testId := vars["testId"]
	var buildId string

	var buildAction *model.BuildAction
	if err := json.NewDecoder(r.Body).Decode(&buildAction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	buildId, err := a.manager.ExecuteBuild(testId, buildAction)

	if err != nil {
		log.Errorf("error creating build: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debugf("saved build: id=%s", buildId)
	tempResponse := map[string]string{
		"id": buildId,
	}

	jsonResponse, err := json.Marshal(tempResponse)

	if err != nil {
		log.Errorf("error marshalling response for create build")
		http.Error(w, err.Error(), http.StatusNoContent)
	}

	log.Infof("started build execution: id=%s", buildId)

	jsonResponse = jsonResponse
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
	return
}

func (a *Api) updateBuildExecution(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	buildId := vars["buildId"]

	build, err := a.manager.GetBuild(buildId)
	if err != nil {
		log.Errorf("error updating build: %s", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var buildAction *model.BuildAction
	if err := json.NewDecoder(r.Body).Decode(&buildAction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&build); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.manager.UpdateBuild(buildId, buildAction); err != nil {
		log.Errorf("error updating build: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("updated build execution: id=%s", build.ID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) addBuildResult(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	buildId := vars["buildId"]

	var buildResult *model.BuildResult
	if err := json.NewDecoder(r.Body).Decode(&buildResult); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	build, err := a.manager.GetBuild(buildId)
	if err != nil {
		log.Errorf("error updating build: %s", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := a.manager.UpdateBuildResults(build.ID,buildResult); err != nil {
		log.Errorf("error updating build: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	log.Infof("added build result: id=%s", build.ID)
	w.WriteHeader(http.StatusCreated)
}

func (a *Api) deleteBuild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildId := vars["buildId"]

	build, err := a.manager.GetBuild(buildId)
	if err != nil {
		log.Errorf("error deleting build: %s", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := a.manager.DeleteBuild(buildId); err != nil {
		log.Errorf("error deleting build: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("deleted build: id=%s", build.ID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) updateBuildStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildId := vars["buildId"]

	build, err := a.manager.GetBuild(buildId)
	if err != nil {
		log.Errorf("error deleting build: %s", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var buildStatus *model.BuildStatus
	if err := json.NewDecoder(r.Body).Decode(&buildStatus); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.manager.UpdateBuildStatus(build, buildStatus); err != nil {
		log.Errorf("error deleting build: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("updated build status: id=%s", build.ID)
	w.WriteHeader(http.StatusNoContent)
}