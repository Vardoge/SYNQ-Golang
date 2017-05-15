package synq

import (
	"bytes"
	"errors"
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

func S3Stub() *httptest.Server {
	var resp []byte
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in s3 req", r.RequestURI)
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			key := r.PostFormValue("key")
			if key != "fakekey" {
				w.Header().Set("Server", "AmazonS3")
				w.Header().Set("X-Amz-Id-2", "vodyoLHQBqirb+3l76iCOoh1Q3Abo8Bm9TntCC1TZso2pL3WGv9aUclvCWloOZynTAEGxNf51hI=")
				w.Header().Set("X-Amz-Request-Id", "9171F45CEDC982B1")
				w.Header().Set("Date", "Fri, 12 May 2017 04:23:53 GMT")
				w.Header().Set("Etag", "9a81d889d4ea7adfa90c9b28b4bbc42f")
				w.Header().Set("Location", key)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		// be default, return error
		resp = loadSample("aws_err.xml")
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Write(resp)
	}))
}

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

type BadReader struct {
}

func (b BadReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed to read")
}

func loadSample(name string) (data []byte) {
	data, err := ioutil.ReadFile("../sample/" + name)
	if err != nil {
		log.Panicf("could not load sample file %s : '%s'", name, err.Error())
	}
	return data
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

func TestParseAwsResp(t *testing.T) {
	assert := assert.New(t)
	var v interface{}
	resp := http.Response{
		StatusCode: 204,
	}
	err := errors.New("failure")
	e := parseAwsResp(&resp, err, v)
	assert.NotNil(e)
	assert.Equal("failure", e.Error())

	br := BadReader{}
	resp = http.Response{
		StatusCode: 412,
		Body:       ioutil.NopCloser(br),
	}
	e = parseAwsResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("failed to read", e.Error())

	err_msg := loadSample("upload.json")
	resp = http.Response{
		StatusCode: 412,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseAwsResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("EOF", e.Error())

	err_msg = loadSample("aws_err.xml")
	resp = http.Response{
		StatusCode: 412,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseAwsResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("At least one of the pre-conditions you specified did not hold", e.Error())

	resp = http.Response{
		StatusCode: 204,
	}
	e = parseAwsResp(&resp, nil, v)
	assert.Nil(e)
}

func TestParseSynqResp(t *testing.T) {
	assert := assert.New(t)
	var v interface{}
	resp := http.Response{
		StatusCode: 200,
	}
	err := errors.New("failure")
	e := parseSynqResp(&resp, err, v)
	assert.NotNil(e)
	assert.Equal("failure", e.Error())
	br := BadReader{}
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(br),
	}
	e = parseSynqResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("failed to read", e.Error())
	err_msg := loadSample("aws_err.xml")
	resp = http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseSynqResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("invalid character '<' looking for beginning of value", e.Error())
	err_msg = []byte(INVALID_UUID)
	resp = http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseSynqResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	msg := []byte("<xml>")
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(msg)),
	}
	e = parseSynqResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("invalid character '<' looking for beginning of value", e.Error())
	msg = loadSample("video.json")
	var video Video
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(msg)),
	}
	e = parseSynqResp(&resp, nil, &video)
	assert.Nil(e)
	assert.Equal(VIDEO_ID, video.Id)
	assert.NotEmpty(video.Input)
}

func TestPostFormFail(t *testing.T) {
	video := Video{}
	assert := assert.New(t)
	setupTestServer(true)
	form := url.Values{}
	api := Api{}
	err := api.postForm("/fake/fail", form, &video)
	assert.NotNil(err)
	assert.Equal("Post /fake/fail: unsupported protocol scheme \"\"", err.Error())
	err = api.postForm(testServer.URL+"/fake/fail", form, &video)
	assert.NotNil(err)
	assert.Equal("fail error", err.Error())
	err = api.postForm(testServer.URL+"/fake/fail_parse", form, &video)
	assert.NotNil(err)
	assert.Equal("unexpected end of JSON input", err.Error())
	err = api.postForm(testServer.URL+"/fake/path_missing", form, &video)
	assert.NotNil(err)
	assert.Equal("unexpected end of JSON input", err.Error())
}

func TestPostForm(t *testing.T) {
	api := Api{}
	video := Video{}
	assert := assert.New(t)
	setupTestServer(true)
	form := url.Values{}
	err := api.postForm(testServer.URL+"/fake/path", form, &video)
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
