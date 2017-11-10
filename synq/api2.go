package synq

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
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

func (a ApiV2) version() string {
	return "v2"
}

func NewV2(token string, timeouts ...time.Duration) ApiV2 {
	base := New(token, timeouts...)
	base.Url = DEFAULT_V2_URL
	return ApiV2{BaseApi: base}
}

func (a ApiV2) handleAuth(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+a.key())
}

func (a ApiV2) getBaseUrl() string {
	return a.url() + "/v2"
}

func (a *ApiV2) CreateAccount() string {
	return ""
}

func (a *ApiV2) Create(userdata ...map[string]interface{}) (VideoV2, error) {
	video := VideoV2{}
	url := a.getBaseUrl() + "/videos"
	body := bytes.NewBuffer([]byte{})
	if len(userdata) > 0 {
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return video, err
	}
	handleReq(a, req, video)
	video.Api = a
	return video, nil
}
