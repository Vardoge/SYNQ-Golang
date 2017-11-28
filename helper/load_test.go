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
