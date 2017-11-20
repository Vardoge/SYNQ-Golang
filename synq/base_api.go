package synq

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/SYNQfm/helpers/common"
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

type ApiF interface {
	key() string
	url() string
	version() string
	timeout(string) time.Duration
	ParseError([]byte) error
	SetUrl(string)
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

func parseSynqResp(a ApiF, resp *http.Response, err error, v interface{}) error {
	if err != nil {
		log.Println("could not make http request")
		return err
	}

	responseAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("could not read resp body")
		return err
	}
	if resp.StatusCode == 204 { // Delete does not have response body
		return nil
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return a.ParseError(responseAsBytes)
	}

	err = json.Unmarshal(responseAsBytes, &v)
	if err != nil {
		log.Printf("could not parse response : %s\n", err.Error())
		return common.NewError("could not parse : %s", string(responseAsBytes))
	}
	return nil
}

func New(key string, timeouts ...time.Duration) BaseApi {
	timeout := time.Duration(DEFAULT_TIMEOUT_MS) * time.Millisecond
	up_timeout := time.Duration(DEFAULT_UPLOAD_MS) * time.Millisecond
	if len(timeouts) > 1 {
		timeout = timeouts[0]
		up_timeout = timeouts[1]
	} else if len(timeouts) > 0 {
		timeout = timeouts[0]
	}
	return BaseApi{
		Key:           key,
		Timeout:       timeout,
		UploadTimeout: up_timeout,
	}
}

func (b *BaseApi) timeout(type_ string) time.Duration {
	if type_ == "upload" {
		return b.UploadTimeout
	} else {
		return b.Timeout
	}
}

func (b *BaseApi) url() string {
	return b.Url
}

func (b *BaseApi) key() string {
	return b.Key
}

func (b *BaseApi) SetUrl(url string) {
	b.Url = url
}

func handleReq(a ApiF, req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: a.timeout("")}
	resp, err := httpClient.Do(req)
	return parseSynqResp(a, resp, err, v)
}

func handleUploadReq(a ApiF, req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: a.timeout("upload")}
	resp, err := httpClient.Do(req)
	return parseAwsResp(resp, err, v)
}
