package synq

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/require"
)

var testAssetId string
var testVideoIdV2 string
var testVideoId2V2 string
var testAuth string
var testServer *test_server.TestServer

func init() {
	testAssetId = test_server.ASSET_ID
	testVideoIdV2 = test_server.V2_VIDEO_ID
	testVideoId2V2 = test_server.V2_VIDEO_ID2
	testAuth = test_server.TEST_AUTH
}

func setupTestApiV2(key string) ApiV2 {
	api := NewV2(key)
	testServer = test_server.SetupServer(SYNQ_VERSION, DEFAULT_SAMPLE_DIR)
	url := testServer.GetUrl()
	api.SetUrl(url)
	api.UploadUrl = url
	return api
}

func TestMakeReq2(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2("fake")
	body := strings.NewReader("")
	req, err := api.makeRequest("POST", "url", body)
	assert.Nil(err)
	assert.Equal("POST", req.Method)
	req, err = api.makeRequest("GET", "url", body)
	assert.Nil(err)
	assert.Equal("GET", req.Method)
}

func TestLogin(t *testing.T) {
	assert := require.New(t)
	server := test_server.SetupServer(SYNQ_VERSION, DEFAULT_SAMPLE_DIR)
	url := server.GetUrl()
	defer server.Close()
	_, err := Login("fake", "fake", url)
	assert.NotNil(err)
	api, e := Login("user", "pass", url)
	assert.Nil(e)
	assert.Equal(test_server.TEST_AUTH, api.Key)
}

func TestCreate2(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2("fake")
	_, err := api.Create()
	assert.NotNil(err)
	api.SetKey(testAuth)
	video, err := api.Create()
	assert.Nil(err)
	assert.Equal(testVideoIdV2, video.Id)
}

func TestGet2(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(testAuth)
	_, err := api.GetVideo("")
	assert.NotNil(err)
	assert.Equal("video id '' is invalid", err.Error())
	video, err := api.GetVideo(testVideoIdV2)
	assert.Nil(err)
	assert.Equal(testVideoIdV2, video.Id)
	assert.Len(video.Assets, 1)
	assert.Equal(testAssetId, video.Assets[0].Id)
	assert.NotNil(video.Api)
}

func TestGetAssetByApi(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(testAuth)
	_, err := api.GetAsset("")
	assert.NotNil(err)
	assert.Equal("asset id '' is invalid", err.Error())
	asset, err := api.GetAsset(testAssetId)
	assert.Nil(err)
	video := asset.Video
	assert.Equal(testAssetId, asset.Id)
	assert.Equal(testVideoIdV2, video.Id)
	assert.Len(video.Assets, 1)
	assert.Equal(testAssetId, video.Assets[0].Id)
	assert.NotNil(video.Api)
}

func TestGetVideos(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(testAuth)
	videos, err := api.GetVideos("")
	assert.Nil(err)
	assert.Len(videos, 2)
	assert.Equal(testVideoIdV2, videos[0].Id)
	assert.Equal(testVideoId2V2, videos[1].Id)
}

func TestParseErrorV2(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(testAuth)
	bytes := []byte{}
	err := api.ParseError(404, bytes)
	assert.Equal("404 Item not found", err.Error())
}

func TestUpdateAssetMetadata(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(testAuth)
	metadata := json.RawMessage("{\"test\": true}")
	asset, err := api.UpdateAssetMetadata(testAssetId, metadata)
	assert.Nil(err)
	assert.Equal(string(metadata), string(asset.Metadata))
}
