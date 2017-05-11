package synq

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

func parseResp(resp *http.Response, err error, v interface{}) error {
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
			return err
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

func (a *Api) handleReq(req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: time.Duration(a.Timeout) * time.Millisecond}
	resp, err := httpClient.Do(req)
	return parseResp(resp, err, v)
}

func (a *Api) postForm(url string, data url.Values, v interface{}) error {
	httpClient := &http.Client{Timeout: time.Duration(a.Timeout) * time.Millisecond}
	resp, err := httpClient.PostForm(url, data)
	return parseResp(resp, err, v)
}

func (a *Api) handlePost(action string, form url.Values, v interface{}) error {
	urlString := a.Url + "/v1/video/" + action
	form.Set("api_key", a.Key)
	return a.postForm(urlString, form, v)
}
