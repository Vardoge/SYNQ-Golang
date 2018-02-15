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
	DEFAULT_V1_URL      = "https://api.synq.fm"
	SYNQ_LEGACY_VERSION = "v1"
	SYNQ_LEGACY_ROUTE   = "v1"
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
	Details map[string]interface{}
}

func (a Api) Version() string {
	return SYNQ_LEGACY_VERSION
}

func (a Api) ParseError(status int, bytes []byte) error {
	var eResp ErrorResp
	err := json.Unmarshal(bytes, &eResp)
	if err != nil {
		log.Printf("could not parse error response : %s\n", err.Error())
		return common.NewError("could not parse %d error : %s", status, string(bytes))
	}
	log.Printf("Received %v\n", eResp)
	return errors.New(eResp.Message)
}

func New(key string, timeouts ...time.Duration) Api {
	return NewV1(key, timeouts...)
}

func NewV1(key string, timeouts ...time.Duration) Api {
	base := NewBase(key, timeouts...)
	base.SetUrl(DEFAULT_V1_URL)
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
	form.Set("api_key", a.GetKey())
	urlStr := a.GetUrl() + "/" + SYNQ_LEGACY_ROUTE + "/video/" + action
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err == nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, err
}

func (a *Api) handlePost(action string, form url.Values, v interface{}) error {
	req, err := a.makeReq(action, form)
	if err != nil {
		return err
	}
	return handleReq(a, req, v)
}
