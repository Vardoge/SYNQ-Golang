package synq

import (
	"bytes"
	"encoding/json"
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
	"github.com/stretchr/testify/require"
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

func SynqStub() *httptest.Server {
	var resp []byte
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in synq response", r.RequestURI)
		testReqs = append(testReqs, r)
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
				case "/v1/video/upload":
					resp, _ = ioutil.ReadFile("../sample/upload.json")
				case "/v1/video/uploader":
					resp, _ = ioutil.ReadFile("../sample/uploader.json")
				case "/v1/video/update":
					resp, _ = ioutil.ReadFile("../sample/update.json")
				case "/v1/video/query":
					resp, _ = ioutil.ReadFile("../sample/query.json")
				default:
					w.WriteHeader(http.StatusBadRequest)
					resp = []byte(HTTP_NOT_FOUND)
				}
			}
		}
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
	assert.Equal("could not parse : <Error>\n  <Code>PreconditionFailed</Code>\n  <Message>At least one of the pre-conditions you specified did not hold</Message>\n  <Condition>Bucket POST must be of the enclosure-type multipart/form-data</Condition>\n  <RequestId>634081169DAFE345</RequestId>\n  <HostId>80jHDkIWiVJd6ofogZSnvEfIxEUk35ULsvWPYFcH5f6VSUMPhCAevKwzLWN+Iw6gGTEvgogepSY=</HostId>\n</Error>\n", e.Error())
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
	assert.Equal("could not parse : <xml>", e.Error())
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
	assert.Equal("could not parse : ", err.Error())
	err = api.postForm(testServer.URL+"/fake/path_missing", form, &video)
	assert.NotNil(err)
	assert.Equal("could not parse : ", err.Error())
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
	assert.Equal("could not parse : ", err.Error())
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

func TestCreate(t *testing.T) {
	assert := require.New(t)
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
	api := setupTestApi(API_KEY, false)
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

func TestUpdateVideo(t *testing.T) {
	assert := assert.New(t)
	api := setupTestApi(API_KEY, false)
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
