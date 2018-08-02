package helper

import (
	"testing"

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

func TestSetupSynqV2(t *testing.T) {
	api2 := SetupSynqV2()
	assert.Equal(t, SYNQ_VERSION, api2.Version())
}

func TestLoadFromFile(t *testing.T) {
	assert := require.New(t)
	_, err := LoadFromFile("fake")
	assert.NotNil(err)
	set, err := LoadFromFile(TEST_CRED_FILE)
	assert.Nil(err)
	assert.Equal("456", set.V2.ApiKey)
	assert.NotNil(set.ApiV2.BaseApi)
	assert.Equal("456", set.ApiV2.GetKey())
}
