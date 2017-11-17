package synq

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type VideoResp struct {
	Video VideoV2 `json:"data"`
}

type VideoV2 struct {
	Id        string                 `json:"id"`
	Userdata  map[string]interface{} `json:"user_data"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Api       *ApiV2                 `json:"-"`
	Assets    []Asset
}

type Asset struct {
	Id       string        `json:"id"`
	Type     string        `json:"type"`
	VideoId  string        `json:"video_id"`
	Location string        `json:"location"`
	State    string        `json:"state"`
	Account  string        `json:"account_id"`
	Metadata VideoMetadata `json:"metadata"`
}

type VideoMetadata struct {
	JobId    string `json:"job_id"`
	JobState string `json:"job_state"`
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

func (v *VideoV2) AddAsset(asset Asset) error {
	url := v.Api.getBaseUrl() + "/assets"
	b, _ := json.Marshal(asset)
	body := bytes.NewBuffer(b)
	req, err := v.Api.makeRequest("POST", url, body)
	if err != nil {
		return err
	}
	a := Asset{}
	err = handleReq(v.Api, req, &a)
	if err != nil {
		return err
	}
	v.Assets = append(v.Assets, a)
	return nil
}

func (a *Asset) Update() error {
	return nil
}
