package synq

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	Api       *Api                   `json:"-"`
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
