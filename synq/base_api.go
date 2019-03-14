package synq

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	Version() string
	GetKey() string
	GetUrl() string
	GetTimeout(string) time.Duration
	SetTimeout(string, time.Duration)
	ParseError(int, []byte) error
	SetUrl(string)
	SetKey(string)
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
	switch resp.StatusCode {
	case 204:
		// Delete, missing does not have response body
		return nil
	case 200,
		201:
		err = json.Unmarshal(responseAsBytes, &v)
		if err != nil {
			log.Printf("could not parse response : %s\n", err.Error())
			return common.NewError("could not parse : %s", string(responseAsBytes))
		}
		return nil
	default:
		return a.ParseError(resp.StatusCode, responseAsBytes)
	}
}

func NewBase(key string, timeouts ...time.Duration) BaseApi {
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

func (b *BaseApi) GetTimeout(type_ string) time.Duration {
	if type_ == "upload" {
		return b.UploadTimeout
	} else {
		return b.Timeout
	}
}

func (b *BaseApi) SetTimeout(type_ string, dur time.Duration) {
	if type_ == "upload" {
		b.UploadTimeout = dur
	} else {
		b.Timeout = dur
	}
}

func (b *BaseApi) GetUrl() string {
	return b.Url
}

func (b *BaseApi) GetKey() string {
	return b.Key
}

func (b *BaseApi) SetUrl(url string) {
	b.Url = url
}

func (b *BaseApi) SetKey(key string) {
	b.Key = key
}

func handleReq(a ApiF, req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: a.GetTimeout("")}
	resp, err := httpClient.Do(req)
	if strings.HasSuffix(req.URL.Path, "/settings") && resp.StatusCode == http.StatusNotFound {
		fmt.Printf("%+v\n", req)
		fmt.Printf("%+v\n", resp)
	}
	return parseSynqResp(a, resp, err, v)
}

func handleUploadReq(a ApiF, req *http.Request, v interface{}) error {
	httpClient := &http.Client{Timeout: a.GetTimeout("upload")}
	resp, err := httpClient.Do(req)
	return parseAwsResp(resp, err, v)
}
