package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/SYNQfm/helpers/common"
)

const (
	DEFAULT_V2_URL = "http://b9n2fsyd6jbfihx82.stoplight-proxy.io"
)

type ApiV2 struct {
	*BaseApi
}

func (a ApiV2) Version() string {
	return "v2"
}

func NewV2(token string, timeouts ...time.Duration) ApiV2 {
	base := New(token, timeouts...)
	base.Url = DEFAULT_V2_URL
	return ApiV2{BaseApi: &base}
}

func (a *ApiV2) handleAuth(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+a.GetKey())
}

func (a ApiV2) getBaseUrl() string {
	return a.GetUrl() + "/v2"
}

func (a *ApiV2) CreateAccount(name string, type_ string) string {
	return ""
}

func (a *ApiV2) makeRequest(method string, url string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	a.handleAuth(req)
	return req, nil
}

func (a ApiV2) ParseError(bytes []byte) error {
	type Resp struct {
		Message string `json:"message"`
	}
	resp := Resp{}
	err := json.Unmarshal(bytes, &resp)
	if err != nil {
		return common.NewError("could not parse : %s", string(bytes))
	}
	return errors.New(resp.Message)
}

func (a *ApiV2) handleGet(url string, v interface{}) error {
	body := bytes.NewBufferString("")
	req, err := a.makeRequest("GET", url, body)
	if err != nil {
		return err
	}
	return handleReq(a, req, v)
}

func (a *ApiV2) Create(userdata ...map[string]interface{}) (VideoV2, error) {
	resp := VideoResp{}
	video := VideoV2{}
	url := a.getBaseUrl() + "/videos"
	body := bytes.NewBuffer([]byte("{"))
	if len(userdata) > 0 {
		body.WriteString(`"user_data":`)
		b, err := json.Marshal(userdata[0])
		if err != nil {
			return video, err
		}
		body.Write(b)
	}
	body.WriteString("}")
	req, err := a.makeRequest("POST", url, body)
	if err != nil {
		return video, err
	}
	err = handleReq(a, req, &resp)
	if err != nil {
		return video, err
	}
	video = resp.Video
	video.Api = a
	return video, nil
}

// Helper function to get details for a video, will create video object
func (a *ApiV2) GetVideo(id string) (video VideoV2, err error) {
	var resp VideoResp
	url := a.getBaseUrl() + "/videos/" + id
	req, err := a.makeRequest("GET", url, nil)
	if err != nil {
		return video, err
	}
	err = handleReq(a, req, &resp)
	if err != nil {
		return video, err
	}
	video = resp.Video
	video.Api = a
	return video, nil
}

func (a *ApiV2) GetAssetList() ([]Asset, error) {
	list := AssetList{}
	url := a.getBaseUrl() + "/assets"
	err := a.handleGet(url, &list)
	return list.Assets, err
}
