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

//func (a *Api) handleReq(req *http.Request, video *Video) error {
func (a *Api) handleReq(url string, data url.Values, v interface{}) error {
	httpClient := &http.Client{Timeout: time.Duration(a.Timeout) * time.Millisecond}
	resp, err := httpClient.PostForm(url, data)
	if err != nil {
		log.Println("could not PostForm")
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
		log.Println("could not parse video response")
		return err
	}
	return nil
}

func (a *Api) handlePost(action string, form url.Values, v interface{}) error {
	urlString := a.Url + "/v1/video/" + action
	form.Set("api_key", a.Key)
	return a.handleReq(urlString, form, v)
}
