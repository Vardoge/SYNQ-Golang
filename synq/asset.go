package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/SYNQfm/SYNQ-Golang/upload"
)

type AssetResponse struct {
	Asset *Asset `json:"data"`
}

type AssetList struct {
	Assets []Asset `json:"data"`
}

type Asset struct {
	AccountId        string                  `json:"account_id"`
	VideoId          string                  `json:"video_id"`
	Id               string                  `json:"id"`
	Location         string                  `json:"location"`
	State            string                  `json:"state"`
	Type             string                  `json:"type"`
	CreatedAt        string                  `json:"created_at"`
	UpdatedAt        string                  `json:"updated_at"`
	Metadata         json.RawMessage         `json:"metadata"`
	Api              ApiV2                   `json:"-"`
	Video            VideoV2                 `json:"-"`
	UploadParameters upload.UploadParameters `json:"-"`
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

func (a *Asset) UploadFile(fileName string) error {
	if a.Api.UploadUrl == "" {
		return errors.New("invalid upload url, can not upload file")
	}
	if a.UploadParameters.Key == "" {
		// if the location exists, get the upload parameters again
		if a.Location != "" && a.Type != "" {
			up, err := a.Video.GetUploadParams(a.Type, a.Id)
			if err != nil {
				return err
			}
			a.UploadParameters = up
		} else {
			return errors.New("upload parameters is invalid")
		}
	}
	f, err := os.Open(fileName)
	defer f.Close()
	if os.IsNotExist(err) {
		return errors.New("file '" + fileName + "' does not exist")
	}

	params := a.UploadParameters
	if !strings.Contains(params.SignatureUrl, "http") {
		sigUrl := a.Api.UploadUrl + params.SignatureUrl
		log.Printf("Updating sig url to include host '%s'\n", a.Api.UploadUrl)
		params.SignatureUrl = sigUrl
	}
	aws, err := upload.CreatorFn(params)
	if err != nil {
		return err
	}
	_, err = aws.Upload(f)
	return err
}
