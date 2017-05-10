package synq

import (
	"io/ioutil"
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
var testValues []url.Values
var testServer *httptest.Server

func ServerStub() *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in req", r.RequestURI)
		testReqs = append(testReqs, r)
		bytes, _ := ioutil.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(bytes))
		testValues = append(testValues, v)
		if strings.Contains(r.RequestURI, "fail_parse") {
			resp = ``
			w.WriteHeader(http.StatusBadRequest)
		} else if strings.Contains(r.RequestURI, "fail") {
			resp = `{"message":"fail error"}`
			w.WriteHeader(http.StatusBadRequest)
		} else if strings.Contains(r.RequestURI, "path_missing") {
			w.WriteHeader(http.StatusOK)
			resp = ``
		} else {
			w.WriteHeader(http.StatusOK)
			resp = `{"created_at": "2017-02-15T03:01:16.767Z","updated_at": "2017-02-16T03:06:31.794Z", "state":"uploaded"}`
		}
		w.Write([]byte(resp))
	}))
}

func setupTestServer(generic bool) {
	if testServer != nil {
		testServer.Close()
	}
	testReqs = testReqs[:0]
	testValues = testValues[:0]
	if generic {
		testServer = ServerStub()
	} else {
		testServer = SynqStub()
	}
}

func setupTestApi(key string, generic bool) Api {
	api := Api{Key: key}
	setupTestServer(generic)
	api.Url = testServer.URL
	return api
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	api := New("key")
	assert.NotNil(api)
	assert.Equal("key", api.Key)
}

func TestHandleReqFail(t *testing.T) {
	video := Video{}
	assert := assert.New(t)
	setupTestServer(true)
	form := url.Values{}
	api := Api{}
	err := api.handleReq("/fake/fail", form, &video)
	assert.NotNil(err)
	assert.Equal("Post /fake/fail: unsupported protocol scheme \"\"", err.Error())
	err = api.handleReq(testServer.URL+"/fake/fail", form, &video)
	assert.NotNil(err)
	assert.Equal("fail error", err.Error())
	err = api.handleReq(testServer.URL+"/fake/fail_parse", form, &video)
	assert.NotNil(err)
	assert.Equal("unexpected end of JSON input", err.Error())
	err = api.handleReq(testServer.URL+"/fake/path_missing", form, &video)
	assert.NotNil(err)
	assert.Equal("unexpected end of JSON input", err.Error())
}

func TestHandleReq(t *testing.T) {
	api := Api{}
	video := Video{}
	assert := assert.New(t)
	setupTestServer(true)
	form := url.Values{}
	err := api.handleReq(testServer.URL+"/fake/path", form, &video)
	assert.Nil(err)
	assert.Len(testReqs, 1)
	r := testReqs[0]
	assert.Equal("/fake/path", r.RequestURI)
	assert.Equal("uploaded", video.State)
	assert.Equal(time.February, video.CreatedAt.Month())
	assert.Equal(15, video.CreatedAt.Day())
	assert.Equal(2017, video.CreatedAt.Year())
	assert.Equal(16, video.UpdatedAt.Day())
	assert.Equal("application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
}

func TestHandlePostFail(t *testing.T) {
	api := setupTestApi("fake", true)
	assert := assert.New(t)
	form := url.Values{}
	video := Video{}
	form.Set("test", "value")
	err := api.handlePost("path_missing", form, &video)
	assert.NotNil(err)
	assert.Equal("unexpected end of JSON input", err.Error())
	api.Url = ":://noprotocol.com"
	err = api.handlePost("path", form, &video)
	assert.NotNil(err)
	assert.Equal("parse :://noprotocol.com/v1/video/path: missing protocol scheme", err.Error())
}

func TestHandlePost(t *testing.T) {
	api := setupTestApi("fake", true)
	assert := assert.New(t)
	form := url.Values{}
	video := Video{}
	form.Set("test", "value")
	err := api.handlePost("create", form, &video)
	assert.Nil(err)
	assert.Len(testReqs, 1)
	r := testReqs[0]
	v := testValues[0]
	assert.Equal("/v1/video/create", r.RequestURI)
	assert.Equal("value", v.Get("test"))
	assert.Equal("fake", v.Get("api_key"))
}
