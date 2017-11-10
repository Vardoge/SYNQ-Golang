package synq

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DEFAULT_V1_URL = "https://api.synq.fm"
)

type Api struct {
	BaseApi
}

func (a Api) version() string {
	return "v1"
}

func NewV1(key string, timeouts ...time.Duration) Api {
	base := New(key, timeouts...)
	base.Url = DEFAULT_V1_URL
	return Api{BaseApi: base}
}

// Helper function to query for videos
func (a *Api) Query(filter string) ([]Video, error) {
	var videos []Video
	form := url.Values{}
	form.Set("filter", filter)
	err := a.handlePost("query", form, &videos)
	return videos, err
}

// Calls the /v1/video/create API to create a new video object
func (a *Api) Create(userdata ...map[string]interface{}) (Video, error) {
	video := Video{}
	form := url.Values{}
	if len(userdata) > 0 {
		bytes, _ := json.Marshal(userdata[0])
		formKey := "userdata"
		form.Set(formKey, string(bytes))
	}
	err := a.handlePost("create", form, &video)
	if err != nil {
		return video, err
	}
	video.Api = a
	return video, nil
}

// Helper function to get details for a video, will create video object
func (a *Api) GetVideo(id string) (Video, error) {
	video := Video{}
	video.Id = id
	video.Api = a
	err := video.GetVideo()
	return video, err
}

// Helper function to update video
func (a *Api) Update(id string, source string) (Video, error) {
	video := Video{}
	video.Id = id
	video.Api = a
	err := video.Update(source)
	return video, err
}

func (a *Api) makeReq(action string, form url.Values) (*http.Request, error) {
	form.Set(("api_key"), a.key())
	urlStr := a.url() + "/v1/video/" + action
	return http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
}

func (a *Api) handlePost(action string, form url.Values, v interface{}) error {
	req, err := a.makeReq(action, form)
	if err != nil {
		return err
	}
	return handleReq(a, req, v)
}
