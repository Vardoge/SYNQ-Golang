package synq

import (
	"strings"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/require"
)

var testAssetId string
var testVideoIdV2 string
var testAuth string

func init() {
	testAssetId = test_server.ASSET_ID
	testVideoIdV2 = test_server.V2_VIDEO_ID
	testAuth = test_server.TEST_AUTH
}

func setupTestApiV2(key string) ApiV2 {
	api := NewV2(key)
	url := test_server.SetupServer("v2")
	api.SetUrl(url)
	return api
}

func TestMakeReq2(t *testing.T) {
	assert := require.New(t)
	api := test_server.SetupServerAndApiV2("fake")
	body := strings.NewReader("")
	req, err := api.makeRequest("POST", "url", body)
	assert.Nil(err)
	assert.Equal("POST", req.Method)
	req, err = api.makeRequest("GET", "url", body)
	assert.Nil(err)
	assert.Equal("GET", req.Method)
}

func TestCreate2(t *testing.T) {
	assert := require.New(t)
	api := test_server.SetupServerAndApiV2("fake")
	_, err := api.Create()
	assert.NotNil(err)
	api.SetKey(testAuth)
	video, err := api.Create()
	assert.Nil(err)
	assert.Equal(testVideoIdV2, video.Id)
}

func TestGet2(t *testing.T) {
	assert := require.New(t)
	api := test_server.SetupServerAndApiV2(testAuth)
	_, err := api.GetVideo("")
	assert.NotNil(err)
	video, err := api.GetVideo(testVideoIdV2)
	assert.Nil(err)
	assert.Equal(testVideoIdV2, video.Id)
	assert.Len(video.Assets, 1)
	assert.Equal(testAssetId, video.Assets[0].Id)
	assert.NotNil(video.Api)
}

func TestParseErrorV2(t *testing.T) {
	assert := require.New(t)
	api := test_server.SetupServerAndApiV2(testAuth)
	bytes := []byte{}
	err := api.ParseError(404, bytes)
	assert.Equal("404 Item not found", err.Error())
}
