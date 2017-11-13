package synq

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func handleV2(w http.ResponseWriter, r *http.Request) {
}

func setupTestApiV2(key string) ApiV2 {
	api := ApiV2{}
	api.Key = key
	setupTestServer("v2")
	api.Url = testServer.URL
	return api
}

func TestCreate2(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2("fake")
	_, err := api.Create()
	assert.Nil(err)
}
