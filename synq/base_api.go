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

const (
	DEFAULT_TIMEOUT_MS = 5000   // 5 seconds
	DEFAULT_UPLOAD_MS  = 600000 // 5 minutes
)

type BaseApi struct {
	Key           string
	Url           string
	Timeout       time.Duration
	UploadTimeout time.Duration
	Version       string
}

type api interface {
	key() string
	url() string
	timeout(string) time.Duration
	makeReq(string, url.Values) *http.Request
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

func New(key string, timeouts ...time.Duration) api {
	timeout := time.Duration(DEFAULT_TIMEOUT_MS) * time.Millisecond
	up_timeout := time.Duration(DEFAULT_UPLOAD_MS) * time.Millisecond
	if len(timeouts) > 1 {
		timeout = timeouts[0]
		up_timeout = timeouts[1]
	} else if len(timeouts) > 0 {
		timeout = timeouts[0]
	}
	var url string
	if strings.Contains(key, ".") {
		url = DEFAULT_V2_URL
	} else {
		url = DEFAULT_V1_URL
	}
	base := BaseApi{
		Key:           key,
		Url:           url,
		Timeout:       timeout,
		UploadTimeout: up_timeout,
	}

	if strings.Contains(key, ".") {
		return ApiV2{BaseApi: base}
	} else {
		return Api{BaseApi: base}
	}
}

func (b BaseApi) isV2() bool {
	return strings.Contains(b.Key, ".")
}

func (b BaseApi) timeout(type_ string) time.Duration {
	if type_ == "upload" {
		return b.UploadTimeout
	} else {
		return b.Timeout
	}
}

func (b BaseApi) url() string {
	return b.Url
}

func (b BaseApi) key() string {
	return b.Key
}

func (b BaseApi) makeReq(action string, form url.Values) *http.Request {
	if b.isV2() {
		method := "POST"
		urlStr := b.url() + "/v2/videos"
		switch action {
		case "details":
			// pull out the video id from the form
			video_id := form.Get("video_id")
			method = "GET"
			urlStr = urlStr + "/" + video_id
		case "update":
			method = "PUT"
		}
		req, _ := http.NewRequest(method, urlStr, strings.NewReader(form.Encode()))
		req.Header.Add("Authorization", "Bearer "+b.key())
		return req
	} else {
		form.Set(("api_key"), b.key())
		urlStr := b.url() + "/v1/video/" + action
		req, _ := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
		return req
	}
}

func (b BaseApi) request(action string, form url.Values, v interface{}) error {
	req := b.makeReq(action, form)
	return b.handleReq(req, v)
}

func (b BaseApi) handleUploadReq(req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: b.timeout("upload")}
	resp, err := httpClient.Do(req)
	return parseAwsResp(resp, err, v)
}

func (b BaseApi) handleReq(req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: b.timeout("")}
	resp, err := httpClient.Do(req)
	return parseSynqResp(resp, err, v)
}
