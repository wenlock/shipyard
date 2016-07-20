package model

type JsonData struct{
	Count 	int `json:"count"`
	Next string `json:"next"`
	Previous string `json:"previous"`
	Results []DockerHubResult `json:"results"`
}
type DockerHubResult struct{
	Name string `json:"name"`
	FullSize int `json:"full_size"`
	Id int `json:"id,omitempty"`
	Repository int `json:"repository"`
	Creator int `json:"creator"`
	LastUpdater int `json:"last_updater"`
	LastUpdated string `json:"last_updated"`
	ImageId string `json:"image_id"`
	V2 bool 	`json:"v2"`
	Platforms []int `json:"platforms"`
}
type MyTag struct {
	Name string `json:"name"`
}
type JsonResult struct {
	Results []MyTag `json:"results"`
}