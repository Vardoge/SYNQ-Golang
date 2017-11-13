package synq

import (
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

func SynqStub(version string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in synq response", r.RequestURI)
		testReqs = append(testReqs, r)
		if version == "v2" {
			handleV2(w, r)
		} else {
			handleV1(w, r)
		}
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

func setupTestServer(type_ ...string) {
	if testServer != nil {
		testServer.Close()
	}
	testReqs = testReqs[:0]
	testValues = testValues[:0]
	t := ""
	if len(type_) > 0 {
		t = type_[0]
	}
	switch t {
	case "generic":
		testServer = ServerStub()
	default:
		testServer = SynqStub(t)
	}
}

func setupTestApi(key string, type_ ...string) Api {
	api := Api{}
	api.Key = key
	setupTestServer(type_...)
	api.Url = testServer.URL
	return api
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	api := New("key")
	assert.NotNil(api)
	assert.Equal("key", api.key())
	assert.Equal(time.Duration(DEFAULT_TIMEOUT_MS)*time.Millisecond, api.timeout(""))
	assert.Equal(time.Duration(DEFAULT_UPLOAD_MS)*time.Millisecond, api.timeout("upload"))
	api = New("key", time.Duration(15)*time.Second)
	assert.Equal("key", api.key())
	assert.Equal(time.Duration(15)*time.Second, api.timeout(""))
	api = New("key", time.Duration(30)*time.Second, time.Duration(100)*time.Second)
	assert.Equal("key", api.key())
	assert.Equal(time.Duration(30)*time.Second, api.timeout(""))
	assert.Equal(time.Duration(100)*time.Second, api.timeout("upload"))
}
