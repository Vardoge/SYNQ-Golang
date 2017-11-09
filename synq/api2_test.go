package synq

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTestApiV2(key string, generic bool) ApiV2 {
	api := ApiV2{}
	api.Key = key
	setupTestServer(generic)
	api.Url = testServer.URL
	return api
}

func TestMakeReq2(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2("fake", false)
	form := make(url.Values)
	req := api.makeReq("create", form)
	assert.NotNil(req)
	assert.Equal("/v2/videos", req.URL.Path)
	assert.Equal("POST", req.Method)
	form.Add("video_id", "123")
	req = api.makeReq("details", form)
	assert.Equal("GET", req.Method)
	assert.Equal("/v2/videos/123", req.URL.Path)
	assert.Equal("Bearer "+api.Key, req.Header.Get("Authorization"))
}
