package synq

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	VIDEO_ID          = "45d4063d00454c9fb21e5186a09c3115"
	VIDEO_ID2         = "55d4062f99454c9fb21e5186a09c2115"
	API_KEY           = "aba179c14ab349e0bb0d12b7eca5fa24"
	API_KEY2          = "cba179c14ab349e0bb0d12b7eca5fa25"
	UPLOAD_KEY        = "projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/uploads/videos/55/d4/55d4062f99454c9fb21e5186a09c2115.mp4"
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
						resp, _ = ioutil.ReadFile("../sample/video.json")
					}
				case "/v1/video/create":
					resp, _ = ioutil.ReadFile("../sample/new_video.json")
				case "/v1/video/upload":
					resp, _ = ioutil.ReadFile("../sample/upload.json")
				default:
					w.WriteHeader(http.StatusBadRequest)
					resp = []byte(HTTP_NOT_FOUND)
				}
			}
		}
		w.Write(resp)
	}))
}

func TestDisplay(t *testing.T) {
	assert := assert.New(t)
	p := Player{EmbedUrl: "url", ThumbnailUrl: "url2"}
	v := Video{}
	assert.Equal("Empty Video\n", v.Display())
	v.State = "created"
	assert.Equal("Empty Video\n", v.Display())
	v.Id = "abc123"
	assert.Equal("Video abc123\n\tState : created\n", v.Display())
	v.State = "uploading"
	assert.Equal("Video abc123\n\tState : uploading\n", v.Display())
	v.State = "uploaded"
	v.Player = p
	assert.Equal("Video abc123\n\tState : uploaded\n\tEmbed URL : url\n\tThumbnail : url2\n", v.Display())
}

func TestGetVideo(t *testing.T) {
	assert := assert.New(t)
	api := setupTestApi("fake", false)
	assert.NotNil(api)
	_, e := api.GetVideo(VIDEO_ID)
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	api.Key = API_KEY
	_, e = api.GetVideo("fake")
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	_, e = api.GetVideo(VIDEO_ID2)
	assert.NotNil(e)
	assert.Equal("Video not found.", e.Error())
	video, e := api.GetVideo(VIDEO_ID)
	assert.Nil(e)
	assert.Equal("uploaded", video.State)
	assert.NotEmpty(video.Input)
	assert.Equal(float64(720), video.Input["width"].(float64))
	assert.Equal(float64(1280), video.Input["height"].(float64))
	assert.Equal("https://player.synq.fm/embed/45d4063d00454c9fb21e5186a09c3115", video.Player.EmbedUrl)
	assert.Equal("https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/thumbnails/45/d4/45d4063d00454c9fb21e5186a09c3115/0000360.jpg", video.Player.ThumbnailUrl)
	assert.Equal(0, video.Player.Views)
	assert.NotEmpty(video.Outputs)
	assert.Len(video.Outputs, 5)
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)
	api := setupTestApi("fake", false)
	assert.NotNil(api)
	_, e := api.Create()
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	api.Key = API_KEY
	v, e := api.Create()
	assert.Nil(e)
	assert.Equal("created", v.State)
	assert.NotNil(v.CreatedAt)
	assert.NotNil(v.UpdatedAt)
	assert.Equal(VIDEO_ID2, v.Id)
}

func TestGetUploadInfo(t *testing.T) {
	assert := assert.New(t)
	api := setupTestApi("fake", false)
	video := Video{Id: VIDEO_ID2, Api: &api}
	err := video.GetUploadInfo()
	assert.NotNil(err)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", err.Error())
	api.Key = API_KEY
	err = video.GetUploadInfo()
	assert.Nil(err)
	assert.NotEmpty(video.UploadInfo)
	assert.Equal(UPLOAD_KEY, video.UploadInfo.Key)
	assert.Equal("public-read", video.UploadInfo.Acl)
	assert.Equal("https://synqfm.s3.amazonaws.com", video.UploadInfo.Action)
	assert.Equal("video/mp4", video.UploadInfo.ContentType)
}

func TestUploadFile(t *testing.T) {
	assert := assert.New(t)
	api := setupTestApi("fake", false)
	video := Video{Id: VIDEO_ID2, Api: &api}
	err := video.UploadFile("myfile.mp4")
	assert.NotNil(err)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", err.Error())
}
