package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var invalidUuid string
var testVideoIdV1 string
var testVideoId2V1 string
var testApiKeyV1 string
var uploadKey string

const (
	sampleDir = "../sample"
)

func init() {
	invalidUuid = test_server.INVALID_UUID
	testVideoIdV1 = test_server.VIDEO_ID
	testVideoId2V1 = test_server.VIDEO_ID2
	testApiKeyV1 = test_server.API_KEY
	uploadKey = test_server.UPLOAD_KEY
}

func setupTestApi(key string, type_ ...string) Api {
	api := NewV1(key)
	url := test_server.SetupServer(type_...)
	api.SetUrl(url)
	return api
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
	assert.Equal("application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
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
	reqs, vals := test_server.GetReqs()
	assert.Len(reqs, 1)
	r := reqs[0]
	v := vals[0]
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
	err_msg = []byte(invalidUuid)
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
	msg = loadSample("video")
	var video Video
	resp = http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBuffer(msg)),
	}
	e = parseSynqResp(a, &resp, nil, &video)
	assert.Nil(e)
	assert.Equal(testVideoIdV1, video.Id)
	assert.NotEmpty(video.Input)
}

func TestCreate(t *testing.T) {
	assert := require.New(t)
	api := setupTestApi("fake")
	assert.NotNil(api)
	_, e := api.Create()
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	api.SetKey(testApiKeyV1)
	v, e := api.Create()
	assert.Nil(e)
	assert.Equal("created", v.State)
	assert.NotNil(v.CreatedAt)
	assert.NotNil(v.UpdatedAt)
	assert.Equal(testVideoId2V1, v.Id)
	// create user userdat
	userdata := make(map[string]interface{})
	userdata["importer"] = make(map[string]interface{})
	import_data := make(map[string]string)
	import_data["content_file"] = "myfile"
	import_data["id"] = "1234"
	userdata["importer"] = import_data
	_, e = api.Create(userdata)
	assert.Nil(e)
	reqs, vals := test_server.GetReqs()
	assert.Len(reqs, 3)
	assert.Len(vals, 3)
	req := *reqs[2]
	val := vals[2]
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
	api := setupTestApi(testApiKeyV1)
	assert.NotNil(api)
	filter := `if (video.state == "uploaded") { return video }`
	videos, err := api.Query(filter)
	assert.Nil(err)
	assert.Len(videos, 3)
	assert.Equal(testVideoIdV1, videos[0].Id)
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
	_, e := api.GetVideo(testVideoIdV1)
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	api.SetKey(testApiKeyV1)
	_, e = api.GetVideo("fake")
	assert.NotNil(e)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", e.Error())
	_, e = api.GetVideo(testVideoId2V1)
	assert.NotNil(e)
	assert.Equal("Video not found.", e.Error())
	video, e := api.GetVideo(testVideoIdV1)
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
	api := setupTestApi(testApiKeyV1)
	assert.NotNil(api)
	source := "video.userdata = {};"
	video, e := api.Update(testVideoIdV1, source)
	assert.Nil(e)
	val := video.Userdata["user"].(string)
	assert.Equal("data", val)
	_, vals := test_server.GetReqs()
	v := vals[0]
	src := v.Get("source")
	assert.Equal(source, src)
}
