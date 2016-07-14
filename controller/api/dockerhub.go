package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
)

// dockerhub forward POC
func (a *Api) dockerhubSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: make an actual proxy using the `github.com/mailgun/oxy/forward` package (note: client cannot change host during forwarding)
	w.Header().Set("content-type", "application/json")

	query := r.URL.Query().Get("q")
	fmt.Printf("query:" + query)

	response, err := http.Get("https://index.docker.io/v1/search?q=" + query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(contents)
}

// get the tags of an image from dockerhub
func (a *Api) dockerhubTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	imageName := vars["imageName"]

	response, err := http.Get("https://hub.docker.com/v2/repositories/library/" + imageName + "/tags/")
	if err != nil {
		http.Error(w, err.Error(), response.StatusCode)
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), response.StatusCode)
		return
	}
	w.Write(contents)
}
