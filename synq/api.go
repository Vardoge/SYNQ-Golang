package synq

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	DEFAULT_URL        = "https://api.synq.fm"
	DEFAULT_V2_URL     = "http://b9n2fsyd6jbfihx82.stoplight-proxy.io/"
	DEFAULT_TIMEOUT_MS = 5000   // 5 seconds
	DEFAULT_UPLOAD_MS  = 600000 // 5 minutes
)

type Api struct {
	Key           string
	Url           string
	Timeout       time.Duration
	UploadTimeout time.Duration
	Version       string
}

type ErrorResp struct {
	//"url": "http://docs.synq.fm/api/v1/error/some_error_code",
	//"name": "Some error occurred.",
	//"message": "A lengthy, human-readable description of the error that has occurred."
	Url     string
	Name    string
	Message string
}

type AwsError struct {
	Code      string
	Message   string
	Condition string
	RequestId string
	HostId    string
}

func New(key string, timeouts ...time.Duration) Api {
	timeout := time.Duration(DEFAULT_TIMEOUT_MS) * time.Millisecond
	up_timeout := time.Duration(DEFAULT_UPLOAD_MS) * time.Millisecond
	if len(timeouts) > 1 {
		timeout = timeouts[0]
		up_timeout = timeouts[1]
	} else if len(timeouts) > 0 {
		timeout = timeouts[0]
	}
	return Api{
		Key:           key,
		Url:           DEFAULT_URL,
		Timeout:       timeout,
		UploadTimeout: up_timeout,
	}
}

func parseAwsResp(resp *http.Response, err error, v interface{}) error {
	if err != nil {
		log.Println("could not make http request")
		return err
	}
	// nothing to process, return success
	if resp.StatusCode == 204 {
		return nil
	}

	var xmlErr AwsError
	// handle this differently
	responseAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("could not parse resp body")
		return err
	}
	err = xml.Unmarshal(responseAsBytes, &xmlErr)
	if err != nil {
		log.Println("could not parse xml", err)
		return err
	}
	return errors.New(xmlErr.Message)
}

func parseSynqResp(resp *http.Response, err error, v interface{}) error {
	if err != nil {
		log.Println("could not make http request")
		return err
	}

	responseAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("could not parse resp body")
		return err
	}
	if resp.StatusCode != 200 {
		var eResp ErrorResp
		err = json.Unmarshal(responseAsBytes, &eResp)
		if err != nil {
			log.Println("could not parse error response")
			return errors.New(fmt.Sprintf("could not parse : %s", string(responseAsBytes)))
		}
		log.Printf("Received %v\n", eResp)
		return errors.New(eResp.Message)

	}
	err = json.Unmarshal(responseAsBytes, &v)
	if err != nil {
		log.Println("could not parse response")
		return errors.New(fmt.Sprintf("could not parse : %s", string(responseAsBytes)))
	}
	return nil
}

func (a *Api) isV2() bool {
	return a.Version == "v2"
}

func (a *Api) handleUploadReq(req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: a.UploadTimeout}
	resp, err := httpClient.Do(req)
	return parseAwsResp(resp, err, v)
}

func (a *Api) sendReq(req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: a.Timeout}
	resp, err := httpClient.Do(req)
	return parseSynqResp(resp, err, v)
}

func (a *Api) makeReq(action string, form url.Values) *http.Request {
	var urlStr string
	method := "POST"
	if a.isV2() {
		urlStr = a.Url + "/v2/videos"
		switch action {
		case "update":
			method = "PUT"
		}
	} else {
		form.Set(("api_key"), a.Key)
		urlStr = a.Url + "/v1/video/" + action
	}
	req, _ := http.NewRequest(method, urlStr, strings.NewReader(form.Encode()))
	return req
}

func (a *Api) request(action string, form url.Values, v interface{}) error {
	req := a.makeReq(action, form)
	return a.sendReq(req, v)
}

// Helper function to query for videos
func (a *Api) Query(filter string) ([]Video, error) {
	var videos []Video
	form := url.Values{}
	form.Set("filter", filter)
	err := a.request("query", form, &videos)
	return videos, err
}

// Calls the /v1/video/create API to create a new video object
func (a *Api) Create(userdata ...map[string]interface{}) (Video, error) {
	video := Video{}
	form := url.Values{}
	if len(userdata) > 0 {
		bytes, _ := json.Marshal(userdata[0])
		formKey := "userdata"
		if a.isV2() {
			formKey = "user_data"
		}
		form.Set(formKey, string(bytes))
	}
	err := a.request("create", form, &video)
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
