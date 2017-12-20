package test_server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/SYNQfm/helpers/common"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var testServers []*TestServer
var recvParams []upload.UploadParameters
var UploadError error

const (
	VIDEO_ID           = "45d4063d00454c9fb21e5186a09c3115"
	VIDEO_ID2          = "55d4062f99454c9fb21e5186a09c2115"
	PROJECT_ID         = "1abfe1b849154082993f2fce78a16fda"
	PROJECT_ID2        = "963bab6186a352b6c0e9de5d29418be3"
	LIVE_VIDEO_ID      = "ec37c42b4aab46f18003b33c66e5e641"
	API_KEY            = "aba179c14ab349e0bb0d12b7eca5fa24"
	API_KEY2           = "cba179c14ab349e0bb0d12b7eca5fa25"
	UPLOAD_KEY         = "projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/uploads/videos/55/d4/55d4062f99454c9fb21e5186a09c2115.mp4"
	INVALID_UUID       = `{"url": "http://docs.synq.fm/api/v1/errors/invalid_uuid","name": "invalid_uuid","message": "Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'."}`
	VIDEO_NOT_FOUND    = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_video","name": "not_found_video","message": "Video not found."}`
	API_KEY_NOT_FOUND  = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_api_key","name": "not_found_api_key","message": "API key not found."}`
	HTTP_NOT_FOUND     = `{"url": "http://docs.synq.fm/api/v1/errors/http_not_found","name": "http_not_found","message": "Not found."}`
	V2_INVALID_AUTH    = `{"message" : "invalid auth"}`
	V2_VIDEO_ID        = "9e9dc8c8-f705-41db-88da-b3034894deb9"
	V2_VIDEO_ID2       = "eee2bc43-e973-4f73-857d-7c0bb111a834"
	ASSET_ID           = "01823629-bcf2-4c34-b714-ae21e1a4647f"
	ASSET_ID2          = "fc3e5d9a-a90e-49cc-0c67-224372a59cee"
	TEST_AUTH          = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwczovL3Rlc3QuYXV0aDAuY29tLyIsInN1YiI6ImF1dGgwfDU3MjE4MjFiM2ExYWFmYmUxNTlkZGE2NSIsImF1ZCI6InRESzZBdUF0QVc0ckFySzhOSTltMXdJRW5WQU9RcjUxIiwiZXhwIjoxNDkzNDM5NTExLCJpYXQiOjE0NjE4MTcxMTF9.29JkFxoHqCRPIH2wVbT-ZNIMBK8xXLwkjbLmyWxpquE"
	DEFAULT_AWS_SECRET = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYzEXAMPLEKEY"
	ACCOUNT_ID         = "425b6b3b-f272-4a33-da4c-19d846685211"
	DEFAULT_SAMPLE_DIR = "sample"
)

type TestServer struct {
	Version   string
	SampleDir string
	Server    *httptest.Server
	Reqs      []*http.Request
	Values    []url.Values
}

func (t *TestServer) Close() {
	t.Server.Close()
}

func (t *TestServer) Reset() {
	t.Reqs = t.Reqs[:0]
	t.Values = t.Values[:0]
}

func (t *TestServer) LoadSampleV2(name string) []byte {
	return LoadSampleDir(name, t.SampleDir+"/v2", []byte(`{"data":[]}`))
}

func (t *TestServer) LoadSample(name string) (data []byte) {
	return LoadSampleDir(name, t.SampleDir, []byte(`{}`))
}

func LoadSampleDir(name string, dir string, dataOnMissing ...[]byte) (data []byte) {
	if !strings.Contains(name, ".") {
		name = name + ".json"
	}
	data, err := ioutil.ReadFile(dir + "/" + name)
	if err != nil {
		logFn := log.Panicf
		if len(dataOnMissing) > 0 {
			logFn = log.Printf
			data = dataOnMissing[0]
		}
		logFn("could not load sample file %s : '%s'", name, err.Error())
	}
	return data
}

func (t *TestServer) GetUrl() string {
	return t.Server.URL
}

type TestAwsUpload struct {
}

func SetSampleDir(sampleDir string) {
	log.Printf("Setting sample dir to %s\n", sampleDir)
	testServer := testServers[len(testServers)-1]
	testServer.SampleDir = sampleDir
}

func (s *TestServer) Setup() string {
	s.Server = httptest.NewServer(http.HandlerFunc(s.handle))
	return s.Server.URL
}

func (s *TestServer) GetReqs() ([]*http.Request, []url.Values) {
	return s.Reqs, s.Values
}

func CloseAll() {
	for _, s := range testServers {
		s.Server.Close()
	}
	testServers = testServers[:0]
}

func SetupServer(args ...string) *TestServer {
	ver := "v1"
	sDir := DEFAULT_SAMPLE_DIR
	if len(args) > 0 {
		ver = args[0]
	}
	if len(args) > 1 {
		sDir = args[1]
	}
	testServer := &TestServer{Version: ver, SampleDir: sDir}
	testServer.Setup()
	testServers = append(testServers, testServer)
	return testServer
}

func LastServer() *TestServer {
	return testServers[len(testServers)-1]
}

func GetReqs() ([]*http.Request, []url.Values) {
	testServer := LastServer()
	return testServer.GetReqs()
}

func ResetReqs() {
	testServer := LastServer()
	testServer.Reset()
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
	} else if id != VIDEO_ID && id != LIVE_VIDEO_ID {
		return VIDEO_NOT_FOUND
	}
	return ""
}

func validateAuth(r *http.Request) string {
	// no auth needed for login
	if r.URL.Path == "/v2/login" {
		return ""
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		// check if "token" is in url
		auth = r.URL.Query().Get("token")
		if auth != TEST_AUTH {
			return V2_INVALID_AUTH
		}
	} else {
		if !strings.Contains(auth, "Bearer ") {
			return V2_INVALID_AUTH
		}
		ret := strings.Split(auth, "Bearer ")
		k := ret[1]
		if k == "fake" {
			return V2_INVALID_AUTH
		}
	}
	return ""
}

func (s *TestServer) handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("here in response %s (server type '%s')", r.RequestURI, s.Version)
	s.Reqs = append(s.Reqs, r)
	switch s.Version {
	case "v2":
		s.handleV2(w, r)
	case "v1":
		s.handleV1(w, r)
	case "s3":
		s.handleS3(w, r)
	default:
		s.handleBasic(w, r)
	}
}

func (s *TestServer) handleBasic(w http.ResponseWriter, r *http.Request) {
	var resp string
	bytes, _ := ioutil.ReadAll(r.Body)
	v, _ := url.ParseQuery(string(bytes))
	s.Values = append(s.Values, v)
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
}

func (s *TestServer) handleS3(w http.ResponseWriter, r *http.Request) {
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
	resp := s.LoadSample("aws_err.xml")
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusPreconditionFailed)
	w.Write(resp)
}

func (s *TestServer) handleV1(w http.ResponseWriter, r *http.Request) {
	var resp []byte
	if r.Method == "POST" {
		bytes, _ := ioutil.ReadAll(r.Body)
		//Parse response body
		v, _ := url.ParseQuery(string(bytes))
		s.Values = append(s.Values, v)
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
					resp = s.LoadSample("video")
				}
			case "/v1/video/create":
				resp = s.LoadSample("new_video")
			case "/v1/video/upload",
				"/v1/video/uploader",
				"/v1/video/update",
				"/v1/video/query":
				paths := strings.Split(r.RequestURI, "/")
				action := paths[len(paths)-1]
				resp = s.LoadSample(action)
			default:
				w.WriteHeader(http.StatusBadRequest)
				resp = []byte(HTTP_NOT_FOUND)
			}
		}
	}
	w.Write(resp)
}

func (s *TestServer) handleV2(w http.ResponseWriter, r *http.Request) {
	var resp []byte
	var k string
	k = validateAuth(r)
	if k != "" {
		w.WriteHeader(http.StatusBadRequest)
		resp = []byte(k)
	} else {
		type_ := "video"
		if strings.Contains(r.URL.Path, "assets") {
			type_ = "asset"
		}
		bytes, _ := ioutil.ReadAll(r.Body)
		body_str := string(bytes)
		v := url.Values{}
		v.Add("body", body_str)
		s.Values = append(s.Values, v)
		switch r.URL.Path {
		case "/v2/videos/" + V2_VIDEO_ID,
			"/v2/assets/" + ASSET_ID:
			if r.Method == "GET" || r.Method == "PUT" {
				if type_ == "asset" {
					resp = s.LoadSample("asset_uploaded")
				} else {
					if r.Method == "PUT" {
						resp = s.LoadSample("video2_update")
					} else {
						resp = s.LoadSample("video2")
					}
				}
				w.WriteHeader(http.StatusOK)
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case "/v2/assets/" + ASSET_ID + "/signature":
			obj := struct {
				Headers string `json:"headers"`
			}{}
			json.Unmarshal(bytes, &obj)
			resp = common.GetMultipartSignature(obj.Headers, DEFAULT_AWS_SECRET)
			w.WriteHeader(http.StatusOK)
		case "/v2/videos/" + V2_VIDEO_ID + "/assets":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				resp = s.LoadSampleV2("asset_list")
				w.WriteHeader(http.StatusOK)
			}
		case "/v2/videos/" + V2_VIDEO_ID + "/upload":
			if r.Method != "POST" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				resp = s.LoadSampleV2("upload")
				w.WriteHeader(http.StatusOK)
			}
		case "/v2/videos",
			"/v2/assets":
			if r.Method != "POST" {
				page := r.URL.Query().Get("page_number")
				if page != "" && page != "1" {
					// for now, anything thats not page 1, return blank
					resp = s.LoadSampleV2(type_ + "_list_" + page)
				} else {
					resp = s.LoadSampleV2(type_ + "_list")
				}
			} else if r.Method == "POST" {
				if type_ == "video" {
					if strings.Contains(body_str, "user_data") {
						resp = s.LoadSample("new_video2_meta")
					} else {
						resp = s.LoadSample("new_video2")
					}
				} else if type_ == "asset" {
					resp = s.LoadSample("asset_created")
					w.WriteHeader(http.StatusCreated)
				}
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case "/v2/login":
			if r.Method == "POST" && !strings.Contains(body_str, "fake") {
				resp = s.LoadSampleV2("login")
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	w.Write(resp)
}

func (t TestAwsUpload) Upload(body io.Reader) (*s3manager.UploadOutput, error) {
	out := &s3manager.UploadOutput{}
	return out, UploadError
}

func NewTestAwsUpload(params upload.UploadParameters) (upload.AwsUploadF, error) {
	recvParams = append(recvParams, params)
	return TestAwsUpload{}, nil
}

func GetParams() []upload.UploadParameters {
	return recvParams
}
