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
	AccountId string          `json:"account_id"`
	VideoId   string          `json:"video_id"`
	Id        string          `json:"id"`
	Location  string          `json:"location"`
	State     string          `json:"state"`
	Type      string          `json:"type"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	Metadata  json.RawMessage `json:"metadata"`
	Api       ApiV2           `json:"-"`
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
