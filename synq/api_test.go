package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handleV1(w http.ResponseWriter, r *http.Request) {
	var resp []byte
	if r.Method == "POST" {
		bytes, _ := ioutil.ReadAll(r.Body)
		//Parse response body
		v, _ := url.ParseQuery(string(bytes))
		testValues = append(testValues, v)
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
			case "/v1/video/upload",
				"/v1/video/uploader",
				"/v1/video/update",
				"/v1/video/query":
				paths := strings.Split(r.RequestURI, "/")
				action := paths[len(paths)-1]
				resp = loadSample(action + ".json")
			default:
				w.WriteHeader(http.StatusBadRequest)
				resp = []byte(HTTP_NOT_FOUND)
			}
		}
	}
	w.Write(resp)
}

func setupTestApi(key string, type_ ...string) Api {
	api := Api{}
	api.Key = key
	SetupTestServer(type_...)
	api.Url = testServer.URL
	return api
}

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

func TestMakeReq(t *testing.T) {
	assert := require.New(t)
	api := setupTestApi("fake")
	form := make(url.Values)
	req, err := api.makeReq("create", form)
	assert.Nil(err)
	assert.NotNil(req)
	assert.Equal("/v1/video/create", req.URL.Path)
	assert.Equal("POST", req.Method)
	body, err := ioutil.ReadAll(req.Body)
	assert.Nil(err)
	assert.Equal("api_key=fake", string(body))
}

func TestHandlePostFail(t *testing.T) {
	api := setupTestApi("fake", "generic")
	assert := assert.New(t)
	form := url.Values{}
	video := Video{}
	form.Set("test", "value")
	err := api.handlePost("path_missing", form, &video)
	assert.NotNil(err)
	assert.Equal("could not parse : ", err.Error())
	api.Url = ":://noprotocol.com"
	err = api.handlePost("path", form, &video)
	assert.NotNil(err)
	assert.Equal("parse :://noprotocol.com/v1/video/path: missing protocol scheme", err.Error())
}

func TestHandlePost(t *testing.T) {
	api := setupTestApi("fake", "generic")
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

func TestParseSynqResp(t *testing.T) {
	assert := assert.New(t)
	var v interface{}
	resp := http.Response{
		StatusCode: 200,
	}
	a := Api{}
	err := errors.New("failure")
	e := parseSynqResp(a, &resp, err, v)
	assert.NotNil(e)
	assert.Equal("failure", e.Error())
	br := BadReader{}
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(br),
	}
	e = parseSynqResp(a, &resp, nil, v)
	assert.NotNil(e)
	assert.Equal("failed to read", e.Error())
	err_msg := loadSample("aws_err.xml")
	resp = http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseSynqResp(a, &resp, nil, v)
	assert.NotNil(e)
	assert.Equal("could not parse : <Error>\n  <Code>PreconditionFailed</Code>\n  <Message>At least one of the pre-conditions you specified did not hold</Message>\n  <Condition>Bucket POST must be of the enclosure-type multipart/form-data</Condition>\n  <RequestId>634081169DAFE345</RequestId>\n  <HostId>80jHDkIWiVJd6ofogZSnvEfIxEUk35ULsvWPYFcH5f6VSUMPhCAevKwzLWN+Iw6gGTEvgogepSY=</HostId>\n</Error>\n", e.Error())
	err_msg = []byte(INVALID_UUID)
	resp = http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseSynqResp(a, &resp, nil, v)
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	msg := []byte("<xml>")
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(msg)),
	}
	e = parseSynqResp(a, &resp, nil, v)
	assert.NotNil(e)
	assert.Equal("could not parse : <xml>", e.Error())
	msg = loadSample("video.json")
	var video Video
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(msg)),
	}
	e = parseSynqResp(a, &resp, nil, &video)
	assert.Nil(e)
	assert.Equal(VIDEO_ID, video.Id)
	assert.NotEmpty(video.Input)
}

func TestCreate(t *testing.T) {
	assert := require.New(t)
	api := setupTestApi("fake")
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
	// create user userdata
	userdata := make(map[string]interface{})
	userdata["importer"] = make(map[string]interface{})
	import_data := make(map[string]string)
	import_data["content_file"] = "myfile"
	import_data["id"] = "1234"
	userdata["importer"] = import_data
	_, e = api.Create(userdata)
	assert.Nil(e)
	assert.Len(testReqs, 3)
	assert.Len(testValues, 3)
	req := *testReqs[2]
	val := testValues[2]
	assert.Equal("/v1/video/create", req.URL.Path)
	data := val.Get("userdata")
	log.Println(data)
	val_data := make(map[string]interface{})
	json.Unmarshal([]byte(data), &val_data)
	d := val_data["importer"].(map[string]interface{})
	assert.Equal(import_data["content_file"], d["content_file"].(string))
	assert.Equal(import_data["id"], d["id"].(string))
}

func TestQuery(t *testing.T) {
	assert := require.New(t)
	api := setupTestApi(API_KEY)
	assert.NotNil(api)
	filter := `if (video.state == "uploaded") { return video }`
	videos, err := api.Query(filter)
	assert.Nil(err)
	assert.Len(videos, 3)
	assert.Equal(VIDEO_ID, videos[0].Id)
	assert.Equal("de98a4c92152411fbac3b7027c8f2df7", videos[1].Id)
	assert.Equal("bb98e6cea8224ea29e6bf00e36632bdf", videos[2].Id)
	for _, video := range videos {
		assert.Equal("uploaded", video.State)
		assert.Equal(2, video.Player.Views)
	}
}

func TestGetVideo(t *testing.T) {
	assert := require.New(t)
	api := setupTestApi("fake")
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

func TestUpdateVideo(t *testing.T) {
	assert := require.New(t)
	api := setupTestApi(API_KEY)
	assert.NotNil(api)
	source := "video.userdata = {};"
	video, e := api.Update(VIDEO_ID, source)
	assert.Nil(e)
	val := video.Userdata["user"].(string)
	assert.Equal("data", val)
	v := testValues[0]
	src := v.Get("source")
	assert.Equal(source, src)
}
