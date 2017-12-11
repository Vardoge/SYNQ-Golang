package synq

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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

func (v *VideoV2) Update() error {
	url := v.Api.getBaseUrl() + "/videos/" + v.Id
	type Update struct {
		Metadata json.RawMessage `json:"metadata"`
		Userdata json.RawMessage `json:"user_data"`
	}
	update := Update{Metadata: v.Metadata, Userdata: v.Userdata}
	b, _ := json.Marshal(update)
	body := bytes.NewBuffer(b)
	req, err := v.Api.makeRequest("PUT", url, body)
	if err != nil {
		return err
	}
	resp := VideoResp{}
	err = handleReq(v.Api, req, &resp)
	if err != nil {
		return err
	}
	v.Metadata = resp.Video.Metadata
	v.Userdata = resp.Video.Userdata
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
		if (a.Location == location || a.Id == location) && a.Id != "" {
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

// This will get the upload params for a sepcific video, if assetId is passed in
// it will be used instead (assuming it exists)
func (v *VideoV2) GetUploadParams(ctype string, assetId ...string) (up UploadParameters, err error) {
	api := v.Api
	params := UnicornParam{
		Ctype: ctype,
	}
	if len(assetId) > 0 {
		params.AssetId = assetId[0]
	}
	return api.GetUploadParams(v.Id, params)
}

// This will call Unicorn's /v2/video/<id>/upload API, which will
// create an asset and create a signed S3 location to upload to, including
// the signature url for multipart uploads
func (v *VideoV2) CreateAssetForUpload(ctype string) (asset Asset, err error) {
	up, err := v.GetUploadParams(ctype)
	if err != nil {
		return asset, err
	}
	// now load the asset
	asset, err = v.GetAsset(up.AssetId)
	if err != nil {
		return asset, err
	}
	asset.UploadParameters = up
	v.Assets = append(v.Assets, asset)
	return asset, nil
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

// Helper function to display information about a file
func (v *VideoV2) Display() (str string) {
	if v.Id == "" {
		str = fmt.Sprintf("Empty Video\n")
	} else {
		str = fmt.Sprintf("Video %s\n\tAssets : %d\n", v.Id, len(v.Assets))
	}
	return str
}
