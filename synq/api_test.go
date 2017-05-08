package synq

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testReqs []*http.Request
var testServer *httptest.Server

const (
	VIDEO_ID          = "45d4062f99454c9fb21e5186a09c2119"
	VIDEO_ID2         = "55d4062f99454c9fb21e5186a09c2115"
	API_KEY           = "aba179c14ab349e0bb0d12b7eca5fa24"
	API_KEY2          = "cba179c14ab349e0bb0d12b7eca5fa25"
	INVALID_UUID      = `{"url": "http://docs.synq.fm/api/v1/errors/invalid_uuid","name": "invalid_uuid","message": "Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'."}`
	VIDEO_NOT_FOUND   = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_video","name": "not_found_video","message": "Video not found."}`
	API_KEY_NOT_FOUND = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_api_key","name": "not_found_api_key","message": "API key not found."}`
	HTTP_NOT_FOUND    = `{"url": "http://docs.synq.fm/api/v1/errors/http_not_found","name": "http_not_found","message": "Not found."}`
)

func ServerStub() *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in req", r.RequestURI)
		testReqs = append(testReqs, r)
		if strings.Contains(r.RequestURI, "fail") {
			resp = `{"message":"fail error"}`
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
			resp = `{"created_at": "2017-02-15T03:01:16.767Z","updated_at": "2017-02-16T03:06:31.794Z", "state":"uploaded"}`
		}
		w.Write([]byte(resp))
	}))
}

func setupTestServer() {
	if testServer != nil {
		testServer.Close()
	}
	testReqs = testReqs[:0]
	testServer = ServerStub()
}

func setupTestApi(key string) Api {
	api := Api{Key: key}
	api.Url = testServer.URL
	return api
}

func TestHandleReqFail(t *testing.T) {
	video := Video{}
	assert := assert.New(t)
	setupTestServer()
	form := url.Values{}
	form.Set("test", "value")
	api := Api{}
	req, err := http.NewRequest("POST", testServer.URL+"/fake/fail", strings.NewReader(form.Encode()))
	assert.Nil(err)
	err = api.handleReq(req, &video)
	assert.NotNil(err)
	assert.Equal("fail error", err.Error())
}

func TestHandleReq(t *testing.T) {
	api := Api{}
	video := Video{}
	assert := assert.New(t)
	setupTestServer()
	form := url.Values{}
	form.Set("test", "value")
	req, err := http.NewRequest("POST", testServer.URL+"/fake/path", strings.NewReader(form.Encode()))
	assert.Nil(err)
	err = api.handleReq(req, &video)
	assert.Nil(err)
	assert.Len(testReqs, 1)
	r := testReqs[0]
	assert.Equal("/fake/path", r.RequestURI)
	assert.Equal("uploaded", video.State)
	assert.Equal(time.February, video.CreatedAt.Month())
	assert.Equal(15, video.CreatedAt.Day())
	assert.Equal(2017, video.CreatedAt.Year())
	assert.Equal(16, video.UpdatedAt.Day())
}

func TestHandlePost(t *testing.T) {

}
