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
	"time"
)

var (
	DEFAULT_URL        = "https://api.synq.fm"
	DEFAULT_TIMEOUT_MS = 5000   // 5 seconds
	DEFAULT_UPLOAD_MS  = 600000 // 5 minutes
)

type Api struct {
	Key string
	Url string
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

func New(key string) Api {
	api := Api{Key: key}
	api.Url = DEFAULT_URL
	return api
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
		return err
	}
	return nil
}

func (a *Api) handleUploadReq(req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: time.Duration(DEFAULT_UPLOAD_MS) * time.Millisecond}
	resp, err := httpClient.Do(req)
	return parseAwsResp(resp, err, v)
}

func (a *Api) postForm(url string, data url.Values, v interface{}) error {
	httpClient := &http.Client{Timeout: time.Duration(DEFAULT_TIMEOUT_MS) * time.Millisecond}
	resp, err := httpClient.PostForm(url, data)
	return parseSynqResp(resp, err, v)
}

func (a *Api) handlePost(action string, form url.Values, v interface{}) error {
	urlString := a.Url + "/v1/video/" + action
	form.Set("api_key", a.Key)
	return a.postForm(urlString, form, v)
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

// Calls the /v1/video/create API to create a new video object
func (a *Api) Create() (Video, error) {
	video := Video{}
	form := url.Values{}
	err := a.handlePost("create", form, &video)
	if err != nil {
		return video, err
	}
	video.Api = a
	return video, nil
}
