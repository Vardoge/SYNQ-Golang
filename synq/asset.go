package synq

import (
	"bytes"
	"encoding/json"
	"io"
)

type AssetResponse struct {
	Asset *Asset `json:"data"`
}

type AssetList struct {
	Assets []Asset `json:"data"`
}

type Asset struct {
	AccountId string                 `json:"account_id"`
	VideoId   string                 `json:"video_id"`
	Id        string                 `json:"id"`
	Location  string                 `json:"location"`
	State     string                 `json:"state"`
	Type      string                 `json:"type"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
	Api       ApiV2                  `json:"-"`
}

func (a *ApiV2) GetAssetList() ([]Asset, error) {
	list := AssetList{}
	url := a.getBaseUrl() + "/assets"
	req, err := a.makeRequest("GET", url, nil)
	if err != nil {
		return list.Assets, err
	}
	err = handleReq(a, req, &list)
	if err != nil {
		return list.Assets, err
	}
	return list.Assets, nil
}

func (v *VideoV2) GetVideoAssetList() error {
	list := AssetList{}
	url := v.Api.getBaseUrl() + "/videos/" + v.Id + "/assets"
	req, err := v.Api.makeRequest("GET", url, nil)
	if err != nil {
		return err
	}
	err = handleReq(v.Api, req, &list)
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

func (v VideoV2) CreateAsset(state, fileType, location string) (Asset, error) {
	var asset Asset
	asset.Api = *v.Api
	asset.VideoId = v.Id
	asset.State = state
	asset.Type = fileType
	asset.Location = location

	url := v.Api.getBaseUrl() + "/assets"
	data, _ := json.Marshal(asset)
	body := bytes.NewBuffer(data)
	err := asset.handleAssetReq("POST", url, body)
	return asset, err
}

func (a *Asset) Update() error {
	url := a.Api.getBaseUrl() + "/assets/" + a.Id
	data, _ := json.Marshal(a)
	body := bytes.NewBuffer(data)
	return a.handleAssetReq("PUT", url, body)
}

func (a *Asset) Delete() error {
	url := a.Api.getBaseUrl() + "/assets/" + a.Id
	data, _ := json.Marshal(a)
	body := bytes.NewBuffer(data)
	return a.handleAssetReq("DELETE", url, body)
}

func (a *Asset) handleAssetReq(method, url string, body io.Reader) error {
	resp := AssetResponse{Asset: a}
	req, err := a.Api.makeRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")

	err = handleReq(a.Api, req, &resp)
	if err != nil {
		return err
	}

	return nil
}
