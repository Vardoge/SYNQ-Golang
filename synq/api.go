package synq

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	DEFAULT_URL        = "https://api.synq.fm"
	DEFAULT_TIMEOUT_MS = 5000
)

type Api struct {
	Key     string
	Url     string
	Timeout int
}

type ErrorResp struct {
	//"url": "http://docs.synq.fm/api/v1/error/some_error_code",
	//"name": "Some error occurred.",
	//"message": "A lengthy, human-readable description of the error that has occurred."
	Url     string
	Name    string
	Message string
}

func New(key string) Api {
	api := Api{Key: key}
	api.Url = DEFAULT_URL
	api.Timeout = DEFAULT_TIMEOUT_MS
	return api
}

func (a *Api) addHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func (a *Api) handleReq(req *http.Request) (video Video, eResp ErrorResp) {
	httpClient := &http.Client{Timeout: time.Duration(a.Timeout) * time.Millisecond}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("could not DO request")
		return video, ErrorResp{Message: err.Error()}
	}
	responseAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("could not parse resp body")
		return video, ErrorResp{Message: err.Error()}
	}

	if resp.StatusCode != 200 {
		err = json.Unmarshal(responseAsBytes, &eResp)
		if err != nil {
			log.Println("could not parse error response")
			return video, ErrorResp{Message: err.Error()}
		}
		return video, eResp
	}
	err = json.Unmarshal(responseAsBytes, &video)
	if err != nil {
		log.Println("could not parse video response")
		return video, ErrorResp{Message: err.Error()}
	}
	return video, eResp
}

func (a *Api) handlePost(action string, form url.Values) (video Video, err error) {
	urlString := a.Url + "/v1/video/" + action
	form.Set("api_key", a.Key)
	req, err := http.NewRequest("POST", urlString, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println("error creating the new request")
		return video, err
	}
	v, e := a.handleReq(req)
	if e.Message != "" {
		return v, errors.New(e.Message)
	}
	return v, nil
}
