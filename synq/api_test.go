package synq

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testReqs []*http.Request
var testServer *httptest.Server

const (
	VIDEO_ID          = "45d4062f99454c9fb21e5186a09c2119"
	API_KEY           = "aba179c14ab349e0bb0d12b7eca5fa24"
	INVALID_UUID      = `{"url": "http://docs.synq.fm/api/v1/errors/invalid_uuid","name": "invalid_uuid","message": "Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'."}`
	VIDEO_NOT_FOUND   = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_video","name": "not_found_video","message": "Video not found."}`
	API_KEY_NOT_FOUND = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_api_key","name": "not_found_api_key","message": "API key not found."}`
	HTTP_NOT_FOUND    = `{"url": "http://docs.synq.fm/api/v1/errors/http_not_found","name": "http_not_found","message": "Not found."}`
)

func validKey(key string) string {
	if len(key) != 32 {
		return INVALID_UUID
	} else if key != API_KEY {
		return API_KEY_NOT_FOUND
	}
	return ""
}

func SynqStub() *httptest.Server {
	var resp []byte
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in synq response", r.RequestURI)
		testReqs = append(testReqs, r)
		if r.Method == "POST" {
			bytes, _ := ioutil.ReadAll(r.Body)
			//Parse response body
			v, _ := url.ParseQuery(string(bytes))
			key := v.Get("api_key")
			ke := validKey(key)
			if ke != "" {
				w.WriteHeader(http.StatusBadRequest)
				resp = []byte(ke)
			} else {
				switch r.RequestURI {
				case "/v1/video/details":
					video_id := v.Get("video_id")
					if video_id == "fake" {
						w.WriteHeader(http.StatusBadRequest)
						resp = []byte(VIDEO_NOT_FOUND)
					} else {
						vResp := new(Video)
						resp, _ = json.Marshal(vResp)
					}
				default:
					w.WriteHeader(http.StatusBadRequest)
					resp = []byte(HTTP_NOT_FOUND)
				}
			}
		}
		w.Write(resp)
	}))
}

func setupSynqStub(api Api) {
	if testServer != nil {
		testServer.Close()
	}
	testReqs = testReqs[:0]
	testServer = SynqStub()
	api.Url = testServer.URL
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	api := Api{Key: "key"}
	assert.NotNil(api)
	assert.Equal("key", api.Key)
}

func TestGetVideo(t *testing.T) {
	assert := assert.New(t)
	api := New("fake_key")
	setupSynqStub(api)
	_, e := api.getVideo(VIDEO_ID)
	assert.NotNil(e)
	api.Key = API_KEY
	_, e = api.getVideo("fake")
	assert.NotNil(e)
	v, e := api.getVideo(VIDEO_ID)
	assert.Nil(e)
	assert.Equal(VIDEO_ID, v.Id)
}
