package synq

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DEFAULT_V2_URL = "http://b9n2fsyd6jbfihx82.stoplight-proxy.io/"
)

type ApiV2 struct {
	BaseApi
}

type VideoV2 struct {
	Id        string                 `json:"id"`
	Userdata  map[string]interface{} `json:"user_data"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Api       *ApiV2                 `json:"-"`
}

func (v VideoV2) Value() (driver.Value, error) {
	json, err := json.Marshal(v)
	return json, err
}

func (v *VideoV2) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	err := json.Unmarshal(source, &v)
	if err != nil {
		return err
	}
	return nil
}

func version() string {
	return "v2"
}

func (a ApiV2) makeReq(command string, form url.Values) *http.Request {
	method := "POST"
	ret := strings.Split(command, "_")
	action := ret[0]
	type_ := "videos"
	if len(ret) > 1 {
		type_ = ret[1] + "s"
	}
	urlStr := a.url() + "/v2/" + type_
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
	req.Header.Add("Authorization", "Bearer "+a.key())
	return req
}

func (a *ApiV2) CreateAccount() string {
	return ""
}

func (a *ApiV2) Create(userdata ...map[string]interface{}) (VideoV2, error) {
	video := VideoV2{}
	form := url.Values{}
	if len(userdata) > 0 {
		bytes, _ := json.Marshal(userdata[0])
		formKey := "user_data"
		form.Set(formKey, string(bytes))
	}
	err := Request(a, "create", form, &video)
	if err != nil {
		return video, err
	}
	video.Api = a
	return video, nil
}
