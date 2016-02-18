package api

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/shipyard/shipyard/model"
)

func (a *Api) images(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	images, err := a.manager.Images()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(images); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) saveImage(w http.ResponseWriter, r *http.Request) {
	var image *model.Image
	if err := json.NewDecoder(r.Body).Decode(&image); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := a.manager.SaveImage(image); err != nil {
		log.Errorf("error saving image: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debugf("saved image: name=%s", image.Name)
	w.WriteHeader(http.StatusCreated)
}
func (a *Api) updateImage(w http.ResponseWriter, r *http.Request) {
	var image *model.Image
	if err := json.NewDecoder(r.Body).Decode(&image); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.manager.UpdateImage(image); err != nil {
		log.Errorf("error updating image: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debugf("updated image: name=%s", image.Name)
	w.WriteHeader(http.StatusNoContent)
}
func (a *Api) image(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	image, err := a.manager.Image(name)
	if err != nil {
		log.Errorf("error deleting image: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(image); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) imagesByProjectId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectId := vars["projectId"]

	image, err := a.manager.ImagesByProjectId(projectId)
	if err != nil {
		log.Errorf("error deleting image: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(image); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (a *Api) deleteImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	image, err := a.manager.Image(name)
	if err != nil {
		log.Errorf("error deleting image: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := a.manager.DeleteImage(image); err != nil {
		log.Errorf("error deleting image: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("deleted image: id=%s name=%s", image.ID, image.Name)
	w.WriteHeader(http.StatusNoContent)
}

