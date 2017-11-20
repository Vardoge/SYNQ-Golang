package synq

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"time"

	"github.com/SYNQfm/helpers/common"
)

const (
	DEFAULT_V1_URL = "https://api.synq.fm"
)

type Api struct {
	*BaseApi
}

type ErrorResp struct {
	//"url": "http://docs.synq.fm/api/v1/error/some_error_code",
	//"name": "Some error occurred.",
	//"message": "A lengthy, human-readable description of the error that has occurred."
	Url     string
	Name    string
	Message string
}

func (a Api) Version() string {
	return "v1"
}

func NewV1(key string, timeouts ...time.Duration) Api {
	base := New(key, timeouts...)
	base.Url = DEFAULT_V1_URL
	return Api{BaseApi: &base}
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

func (a Api) ParseError(bytes []byte) error {
	var eResp ErrorResp
	err := json.Unmarshal(bytes, &eResp)
	if err != nil {
		log.Printf("could not parse error response : %s\n", err.Error())
		return common.NewError("could not parse : %s", string(bytes))
	}
	log.Printf("Received %v\n", eResp)
	return errors.New(eResp.Message)
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
	form.Set(("api_key"), a.GetKey())
	urlStr := a.GetUrl() + "/v1/video/" + action
	return http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
}

func (a *Api) handlePost(action string, form url.Values, v interface{}) error {
	req, err := a.makeReq(action, form)
	if err != nil {
		return err
	}
	return handleReq(a, req, v)
}
