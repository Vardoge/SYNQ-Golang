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

func validKey(key string) string {
	if len(key) != 32 {
		return INVALID_UUID
	} else if key != API_KEY {
		return API_KEY_NOT_FOUND
	}
	return ""
}

func validVideo(id string) string {
	if len(id) != 32 {
		return INVALID_UUID
	} else if id != VIDEO_ID {
		return VIDEO_NOT_FOUND
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
					ke = validVideo(video_id)
					if ke != "" {
						w.WriteHeader(http.StatusBadRequest)
						resp = []byte(ke)
					} else {
						vResp := new(Video)
						vResp.Id = VIDEO_ID
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

func setupTestVideo(key string) Video {
	if testServer != nil {
		testServer.Close()
	}
	testReqs = testReqs[:0]
	testServer = SynqStub()
	api := Api{Key: key}
	api.Url = testServer.URL
	return Video{Api: api}
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	video := New("key")
	assert.NotNil(video)
	assert.Equal("key", video.Api.Key)
	assert.Equal("", video.Id)
}

func TestGetVideo(t *testing.T) {
	assert := assert.New(t)
	video := setupTestVideo("fake_key")
	assert.NotNil(video)
	video.Id = VIDEO_ID
	e := video.Details()
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	video.Api.Key = API_KEY
	video.Id = "fake"
	e = video.Details()
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	video.Id = VIDEO_ID2
	e = video.Details()
	assert.NotNil(e)
	assert.Equal("Video not found.", e.Error())
	video.Id = VIDEO_ID
	e = video.Details()
	assert.Nil(e)
	assert.NotEmpty(video.Input)
}
