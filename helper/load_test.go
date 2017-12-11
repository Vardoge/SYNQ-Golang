package helper

import (
	"os"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/require"
)

const (
	sampleDir = "../sample"
	cacheDir  = "cache_dir"
)

var testAuth string

type Cache struct {
	Dir string
}

func init() {
	test_server.SetSampleDir(sampleDir)
	testAuth = test_server.TEST_AUTH
}

func setup() (synq.ApiV2, Cache) {
	os.RemoveAll(cacheDir)
	os.MkdirAll(cacheDir, 0755)
	cache := Cache{Dir: cacheDir}
	api := synq.NewV2(testAuth)
	url := test_server.SetupServer("v2")
	api.SetUrl(url)
	api.UploadUrl = url
	return api, cache
}

func (c Cache) GetCacheFile(name string) string {
	return c.Dir + "/" + name + ".json"
}

func TestLoadVideo(t *testing.T) {
	assert := require.New(t)
	api, cache := setup()
	v, e := LoadVideoV2(test_server.V2_VIDEO_ID, cache, api)
	assert.Nil(e)
	assert.Equal(test_server.V2_VIDEO_ID, v.Id)
	cacheFile := cache.GetCacheFile(v.Id)
	_, err := os.Stat(cacheFile)
	assert.Nil(err)
	// should avoid the call
	api2 := synq.ApiV2{}
	v2, e2 := LoadVideoV2(test_server.V2_VIDEO_ID, cache, api2)
	assert.Nil(e2)
	assert.Equal(v.Id, v2.Id)
	assert.NotEmpty(v2.Api)
}

func TestLoadAsset(t *testing.T) {
	assert := require.New(t)
	api, cache := setup()
	a, e := LoadAsset(test_server.ASSET_ID, cache, api)
	assert.Nil(e)
	assert.Equal(test_server.ASSET_ID, a.Id)
	assert.Equal(test_server.V2_VIDEO_ID, a.Video.Id)
	cacheFile := cache.GetCacheFile(a.Id)
	_, err := os.Stat(cacheFile)
	assert.Nil(err)
	// should avoid the call
	api2 := synq.ApiV2{}
	a2, e2 := LoadAsset(test_server.ASSET_ID, cache, api2)
	assert.Nil(e2)
	assert.Equal(a.Id, a2.Id)
	assert.NotEmpty(a2.Api)
}

func TestLoadUp(t *testing.T) {
	assert := require.New(t)
	api, cache := setup()
	params := synq.UnicornParam{
		Ctype: "video/mp4",
	}
	up, err := LoadUploadParameters(test_server.V2_VIDEO_ID, params, cache, api)
	assert.Nil(err)
	assert.Equal("uploads/9e/9d/9e9dc8c8f70541db88dab3034894deb9/01823629bcf24c34b714ae21e1a4647f.mp4", up.Key)
	cacheFile := cache.GetCacheFile(test_server.V2_VIDEO_ID + "_up")
	_, err = os.Stat(cacheFile)
	assert.Nil(err)
}
