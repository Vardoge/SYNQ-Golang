package helper

import (
	"testing"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TEST_CRED_FILE = "../sample/cred.json"
)

func TestGetSetup(t *testing.T) {
	assert := assert.New(t)
	setup := GetSetupByEnv("")
	assert.Equal("", setup.Version)
	setup = GetSetupByEnv(SYNQ_VERSION)
	assert.Equal(SYNQ_VERSION, setup.Version)
}

func TestSetupSynq(t *testing.T) {
	assert := assert.New(t)
	api := SetupSynq()
	assert.Equal(SYNQ_LEGACY_VERSION, api.Version())
}

func TestSetupSynqV2(t *testing.T) {
	api2 := SetupSynqV2()
	assert.Equal(t, SYNQ_VERSION, api2.Version())
}

func TestConfigure(t *testing.T) {
	assert := require.New(t)
	setting := ApiSetting{Url: "url", Timeout: 111, UploadTimeout: 555}
	api := synq.NewV1("123")
	setting.Configure(api)
	assert.Equal(setting.Url, api.GetUrl())
	assert.Equal(time.Duration(setting.Timeout)*time.Second, api.GetTimeout(""))
	assert.Equal(time.Duration(setting.UploadTimeout)*time.Second, api.GetTimeout("upload"))
}

func TestSetup(t *testing.T) {
	assert := require.New(t)
	setting := ApiSetting{ApiKey: "123"}
	set := ApiSet{V1: setting}
	assert.Nil(set.ApiV1.BaseApi)
	assert.Nil(set.ApiV2.BaseApi)
	set.Setup()
	assert.NotNil(set.ApiV1.BaseApi)
	assert.Nil(set.ApiV2.BaseApi)
	assert.Equal("123", set.ApiV1.GetKey())
}

func TestLoadFromFile(t *testing.T) {
	assert := require.New(t)
	_, err := LoadFromFile("fake")
	assert.NotNil(err)
	set, err := LoadFromFile(TEST_CRED_FILE)
	assert.Nil(err)
	assert.Equal("123", set.V1.ApiKey)
	assert.Equal("456", set.V2.ApiKey)
	assert.NotNil(set.ApiV1.BaseApi)
	assert.NotNil(set.ApiV2.BaseApi)
	assert.Equal("123", set.ApiV1.GetKey())
	assert.Equal("456", set.ApiV2.GetKey())
}
