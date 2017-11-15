package synq

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	TEST_AUTH       = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwczovL3Rlc3QuYXV0aDAuY29tLyIsInN1YiI6ImF1dGgwfDU3MjE4MjFiM2ExYWFmYmUxNTlkZGE2NSIsImF1ZCI6InRESzZBdUF0QVc0ckFySzhOSTltMXdJRW5WQU9RcjUxIiwiZXhwIjoxNDkzNDM5NTExLCJpYXQiOjE0NjE4MTcxMTF9.29JkFxoHqCRPIH2wVbT-ZNIMBK8xXLwkjbLmyWxpquE"
	V2_INVALID_AUTH = `{"message" : "invalid auth"}`
	V2_VIDEO_ID     = "9e9dc8c8-f705-41db-88da-b3034894deb9"
)

func handleV2(w http.ResponseWriter, r *http.Request) {
	var resp []byte
	auth := r.Header.Get("Authorization")
	k := validateAuth(auth)
	if k != "" {
		w.WriteHeader(http.StatusBadRequest)
		resp = []byte(k)
	} else {
		switch r.URL.Path {
		case "/v2/videos":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				bytes, _ := ioutil.ReadAll(r.Body)
				if strings.Contains(string(bytes), "user_data") {
					resp = loadSample("new_video2_meta")
				} else {
					resp = loadSample("new_video2")
				}
				w.WriteHeader(http.StatusCreated)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	w.Write(resp)
}

func validateAuth(key string) string {
	if !strings.Contains(key, "Bearer ") {
		return V2_INVALID_AUTH
	}
	ret := strings.Split(key, "Bearer ")
	k := ret[1]
	if k == "fake" {
		return V2_INVALID_AUTH
	}
	return ""
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
	assert.NotNil(err)
	api.Key = TEST_AUTH
	video, err := api.Create()
	assert.Nil(err)
	assert.Equal(V2_VIDEO_ID, video.Id)
}
