package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/SYNQfm/helpers/common"
)

const (
	DEFAULT_V2_URL       = "https://b9n2fsyd6jbfihx82.stoplight-proxy.io"
	DEFAULT_UPLOADER_URL = "https://s6krcbatzuuhmspse.stoplight-proxy.io"
	DEFAULT_PAGE_SIZE    = 100
	SYNQ_VERSION         = "v2"
	SYNQ_ROUTE           = "v1"
)

type ApiV2 struct {
	*BaseApi
	User        string
	Password    string
	UploadUrl   string
	TokenExpiry time.Time
	PageSize    int
}

type AccountResp struct {
	Account Account `json:"data"`
}

type Account struct {
	Id              string           `json:"id"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	Status          string           `json:"status"`
	Domain          string           `json:"domain"`
	Contact         string           `json:"contact_person"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	PartneredOn     string           `json:"partnered_on"`
	AccountSettings []AccountSetting `json:"account_settings"`
	Distributors    []Distributor    `json:"distributor_accounts"`
}

type AccountSetting struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Distributor struct {
	DistributorId string `json:"distributor_id"`
}

type VideoList struct {
	Videos     []json.RawMessage `json:"data"`
	PageSize   int               `json:"page_size"`
	PageNumber int               `json:"page_number"`
}

type ErrorRespV2 struct {
	Message string `json:"message"`
}

type LoginResp struct {
	Token       string    `json:"jwt"`
	TokenExpiry time.Time `json:"exp"`
}

func (a ApiV2) Version() string {
	return SYNQ_VERSION
}

func NewV2(token string, timeouts ...time.Duration) ApiV2 {
	base := NewBase(token, timeouts...)
	base.SetUrl(DEFAULT_V2_URL)
	api := ApiV2{BaseApi: &base}
	api.PageSize = DEFAULT_PAGE_SIZE
	return api
}

func (a *ApiV2) handleAuth(req *http.Request) {
	if strings.HasPrefix(a.GetKey(), "Bearer ") {
		req.Header.Add("Authorization", a.GetKey())
	} else {
		req.Header.Add("Authorization", "Bearer "+a.GetKey())
	}
}

func (a ApiV2) getBaseUrl() string {
	return a.GetUrl() + "/" + SYNQ_ROUTE
}

func (a *ApiV2) CreateAccount(name string, type_ string) string {
	return ""
}

func (a *ApiV2) GetAccount(id string) (account Account, err error) {
	var resp AccountResp
	url := a.getBaseUrl() + "/accounts/" + id
	req, err := a.makeRequest("GET", url, nil)
	if err != nil {
		return account, err
	}
	err = handleReq(a, req, &resp)
	if err != nil {
		return account, err
	}
	return resp.Account, nil
}

func (a *ApiV2) makeRequest(method string, url string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}
	if method == "POST" || method == "PUT" {
		if strings.Contains(url, "/login") {
			req.Header.Add("content-type", "application/x-www-form-urlencoded")
		} else {
			req.Header.Add("content-type", "application/json")
		}
	}
	a.handleAuth(req)
	return req, nil
}

func (a ApiV2) ParseError(status int, bytes []byte) error {
	if status == 404 {
		return errors.New("404 Item not found")
	}
	type Resp struct {
		Message string `json:"message"`
	}
	resp := Resp{}
	err := json.Unmarshal(bytes, &resp)
	if err != nil {
		return common.NewError("could not parse error %d : %s", status, string(bytes))
	}
	msg := resp.Message
	if msg == "" {
		msg = fmt.Sprintf("Failed with status %d", status)
	}
	return errors.New(msg)
}

func (a *ApiV2) handleGet(url string, v interface{}) error {
	body := bytes.NewBufferString("")
	req, err := a.makeRequest("GET", url, body)
	if err != nil {
		return err
	}
	return handleReq(a, req, v)
}

func Login(user, password string, serverUrl ...string) (ApiV2, error) {
	var api ApiV2
	resp, err := login(user, password, serverUrl...)
	if err != nil {
		return api, err
	}
	api = NewV2(resp.Token)
	api.TokenExpiry = resp.TokenExpiry
	api.User = user
	api.Password = password
	if len(serverUrl) > 0 {
		api.SetUrl(serverUrl[0])
	}
	return api, nil
}

func login(user, password string, serverUrl ...string) (LoginResp, error) {
	var r LoginResp
	var u string
	if len(serverUrl) > 0 {
		u = serverUrl[0]
	} else {
		u = DEFAULT_V2_URL
	}
	u = u + "/" + SYNQ_ROUTE + "/login"
	form := url.Values{}
	form.Add("email", user)
	form.Add("password", password)
	resp, e := http.PostForm(u, form)
	if e != nil {
		return r, e
	}
	if resp.StatusCode != 200 {
		return r, common.NewError("error getting login %d", resp.StatusCode)
	}
	bytes, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(bytes, &r)
	return r, nil
}

func (a *ApiV2) Create(body ...[]byte) (VideoV2, error) {
	resp := VideoResp{}
	video := VideoV2{}
	url := a.getBaseUrl() + "/videos"
	buf := bytes.NewBufferString("")
	if len(body) > 0 {
		buf.Write(body[0])
	}
	req, err := a.makeRequest("POST", url, buf)
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

// This will return the "raw" json, using pagination, and the calling function is expected
// to turn it into whats needed
func (a *ApiV2) getVideos(accountId string) (videos []json.RawMessage, err error) {
	path := "/videos"
	if accountId != "" {
		path = "/accounts/" + accountId + path
	}
	// iterate through the requests until we get no more videos
	base_url := a.getBaseUrl() + path
	pageNumber := 1
	pageSize := a.PageSize
	for {
		var obj VideoList
		url := base_url + fmt.Sprintf("?page_number=%d&page_size=%d", pageNumber, pageSize)
		req, err := a.makeRequest("GET", url, nil)
		if err != nil {
			return videos, err
		}
		err = handleReq(a, req, &obj)
		if err != nil {
			return videos, err
		}
		if len(obj.Videos) == 0 {
			return videos, nil
		}
		videos = append(videos, obj.Videos...)
		pageNumber++
	}
	return videos, nil
}

func (a *ApiV2) GetVideos(accountId string) ([]VideoV2, error) {
	var videos []VideoV2
	raw, err := a.getVideos(accountId)
	if err != nil {
		return videos, err
	}
	for _, v := range raw {
		var video VideoV2
		json.Unmarshal(v, &video)
		video.Api = a
		videos = append(videos, video)
	}
	return videos, nil
}

func (a *ApiV2) GetRawVideos(accountId string) ([]json.RawMessage, error) {
	return a.getVideos(accountId)
}

// this sets the api object properly on the Video object and the assets
func (a *ApiV2) SetApi(video *VideoV2) {
	video.Api = a
	assets := []Asset{}
	for _, a := range video.Assets {
		a.Video = *video
		a.Api = *video.Api
		assets = append(assets, a)
	}
	video.Assets = assets
}

// Helper function to get details for a video, will create video object
func (a *ApiV2) GetVideo(id string) (video VideoV2, err error) {
	var resp VideoResp
	if !common.ValidUUID(id) {
		return video, common.NewError("video id '%s' is invalid", id)
	}
	uuid := common.ConvertToUUIDFormat(id)
	url := a.getBaseUrl() + "/videos/" + uuid
	req, err := a.makeRequest("GET", url, nil)
	if err != nil {
		return video, err
	}
	err = handleReq(a, req, &resp)
	if err != nil {
		return video, err
	}
	video = resp.Video
	a.SetApi(&video)
	return video, nil
}

// Helper function to get an Asset
func (a *ApiV2) GetAsset(id string) (asset Asset, err error) {
	var resp AssetResponse
	if !common.ValidUUID(id) {
		return asset, common.NewError("asset id '%s' is invalid", id)
	}
	uuid := common.ConvertToUUIDFormat(id)
	url := a.getBaseUrl() + "/assets/" + uuid
	req, err := a.makeRequest("GET", url, nil)
	if err != nil {
		return asset, err
	}
	err = handleReq(a, req, &resp)
	if err != nil {
		return asset, err
	}
	asset = *resp.Asset
	// now get the video
	video, err := a.GetVideo(asset.VideoId)
	if err != nil {
		return asset, err
	}
	asset.Video = video
	asset.Api = *a
	return asset, nil
}

func (a *ApiV2) GetAssetList() ([]Asset, error) {
	list := AssetList{}
	url := a.getBaseUrl() + "/assets"
	err := a.handleGet(url, &list)
	return list.Assets, err
}

func (a *ApiV2) GetUploadParams(vid string, params upload.UploadRequest) (up upload.UploadParameters, err error) {
	if a.UploadUrl == "" {
		return up, errors.New("UploadUrl is blank")
	}
	url := a.UploadUrl + "/" + SYNQ_ROUTE + "/videos/" + vid + "/upload"
	data, _ := json.Marshal(params)
	body := bytes.NewBuffer(data)

	req, err := a.makeRequest("POST", url, body)
	if err != nil {
		return up, err
	}
	err = handleReq(a, req, &up)
	return up, err
}

func (a *ApiV2) UpdateAssetMetadata(id string, metadata json.RawMessage) (asset Asset, err error) {
	var resp AssetResponse
	if !common.ValidUUID(id) {
		return asset, common.NewError("asset id '%s' is invalid", id)
	}
	uuid := common.ConvertToUUIDFormat(id)
	url := a.getBaseUrl() + "/assets/" + uuid
	req, err := a.makeRequest("PUT", url, strings.NewReader("{\"metadata\": "+string(metadata)+"}"))
	if err != nil {
		return asset, err
	}
	err = handleReq(a, req, &resp)
	if err != nil {
		return asset, err
	}
	asset = *resp.Asset
	asset.Api = *a
	return asset, nil
}

func (a *ApiV2) CreateAssetSettings(assetId string, settingIds []string) error {
	url := fmt.Sprintf("%s/assets/%s/settings", a.getBaseUrl(), assetId)
	data, _ := json.Marshal(map[string][]string{"settings_ids": settingIds})
	body := bytes.NewBuffer(data)
	req, err := a.makeRequest("POST", url, body)
	if err != nil {
		return err
	}
	return handleReq(a, req, new(map[string]interface{}))
}
