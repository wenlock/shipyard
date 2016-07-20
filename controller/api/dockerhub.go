package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/shipyard/shipyard/model"
	"strings"
)
// code to get all tags from dockerhub



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

	allTags := []model.MyTag{}
	globalContents := []model.JsonData{}

	repo := r.URL.Query().Get("r")
	fmt.Printf("Got image name: %s\n", repo)
	if !strings.Contains(repo,"/"){
		repo = "library/" + repo
	}
	i := 1
	for {
		jsonData := model.JsonData{}
		response, err := http.Get("https://hub.docker.com/v2/repositories/" + repo + "/tags/?page=" + strconv.Itoa(i))
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
		err = json.Unmarshal(contents, &jsonData)
		if err != nil {
			fmt.Printf("Error unmarshaling data!\n")
		}
		globalContents = append(globalContents, jsonData)
		if (jsonData.Next == ""){
			break
		}
		i ++
	}
	for _, data := range globalContents{
		for _, res := range data.Results{

			tag := model.MyTag{}
			tag.Name = res.Name
			allTags = append(allTags, tag)
		}
	}
	finalResult := model.JsonResult{Results: allTags}
	myContents, err := json.Marshal(finalResult)
	if err != nil {
		fmt.Printf("Error marshaling data!\n")
	}
	w.Write(myContents)
}
