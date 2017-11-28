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
	Id        string          `json:"id"`
	Userdata  json.RawMessage `json:"user_data"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Api       *ApiV2          `json:"-"`
	Assets    []Asset         `json:"assets"`
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

func (v *VideoV2) GetVideoAssetList() error {
	list := AssetList{}
	url := v.Api.getBaseUrl() + "/videos/" + v.Id + "/assets"
	err := v.Api.handleGet(url, &list)
	if err != nil {
		return err
	}
	v.Assets = list.Assets
	return nil
}

func (v VideoV2) GetAsset(assetId string) (Asset, error) {
	url := v.Api.getBaseUrl() + "/assets/" + assetId
	var asset Asset
	asset.Api = *v.Api
	err := asset.handleAssetReq("GET", url, nil)
	return asset, err
}

func (v *VideoV2) FindAsset(location string) (Asset, bool) {
	for _, a := range v.Assets {
		if a.Location == location && a.Id != "" {
			return a, true
		}
	}
	return Asset{}, false
}

func (v *VideoV2) CreateOrUpdateAsset(asset *Asset) error {
	// make sure the API is set
	asset.Api = *v.Api
	// check if this asset exists, if it does, just update
	a, found := v.FindAsset(asset.Location)
	if found {
		asset.Id = a.Id
		return asset.Update()
	} else {
		url := v.Api.getBaseUrl() + "/assets"
		data, _ := json.Marshal(asset)
		body := bytes.NewBuffer(data)
		err := asset.handleAssetReq("POST", url, body)
		if err == nil {
			v.Assets = append(v.Assets, *asset)
		}
		return err
	}
}

func (v *VideoV2) CreateAsset(state, fileType, location string) (Asset, error) {
	var asset Asset
	asset.VideoId = v.Id
	asset.State = state
	asset.Type = fileType
	asset.Location = location
	err := v.CreateOrUpdateAsset(&asset)
	return asset, err
}
