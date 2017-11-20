package test_helper

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

var defaultSampleDir = "sample"

var testReqs []*http.Request
var testValues []url.Values
var testServer *httptest.Server

const (
	VIDEO_ID          = "45d4063d00454c9fb21e5186a09c3115"
	VIDEO_ID2         = "55d4062f99454c9fb21e5186a09c2115"
	PROJECT_ID        = "1abfe1b849154082993f2fce78a16fda"
	PROJECT_ID2       = "963bab6186a352b6c0e9de5d29418be3"
	LIVE_VIDEO_ID     = "ec37c42b4aab46f18003b33c66e5e641"
	API_KEY           = "aba179c14ab349e0bb0d12b7eca5fa24"
	API_KEY2          = "cba179c14ab349e0bb0d12b7eca5fa25"
	UPLOAD_KEY        = "projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/uploads/videos/55/d4/55d4062f99454c9fb21e5186a09c2115.mp4"
	INVALID_UUID      = `{"url": "http://docs.synq.fm/api/v1/errors/invalid_uuid","name": "invalid_uuid","message": "Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'."}`
	VIDEO_NOT_FOUND   = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_video","name": "not_found_video","message": "Video not found."}`
	API_KEY_NOT_FOUND = `{"url": "http://docs.synq.fm/api/v1/errors/not_found_api_key","name": "not_found_api_key","message": "API key not found."}`
	HTTP_NOT_FOUND    = `{"url": "http://docs.synq.fm/api/v1/errors/http_not_found","name": "http_not_found","message": "Not found."}`
	V2_INVALID_AUTH   = `{"message" : "invalid auth"}`
	V2_VIDEO_ID       = "9e9dc8c8-f705-41db-88da-b3034894deb9"
	ASSET_ID          = "01823629-bcf2-4c34-b714-ae21e1a4647f"
)

func SetSampleDir(sampleDir string) {
	log.Printf("Setting sample dir to %s\n", sampleDir)
	defaultSampleDir = sampleDir
}

func LoadSample(name string, sampleDir ...string) (data []byte) {
	sDir := defaultSampleDir
	if len(sampleDir) > 0 {
		sDir = sampleDir[0]
	}
	if !strings.Contains(name, ".") {
		name = name + ".json"
	}
	data, err := ioutil.ReadFile(sDir + "/" + name)
	if err != nil {
		log.Panicf("could not load sample file %s : '%s'", name, err.Error())
	}
	return data
}

func SetupServer(version ...string) string {
	ver := "v1"
	if len(version) > 0 {
		ver = version[0]
	}
	if testServer != nil {
		testServer.Close()
	}
	testReqs = testReqs[:0]
	testValues = testValues[:0]
	if ver == "generic" {
		testServer = ServerStub()
	} else {
		testServer = SynqStub(ver)
	}
	return testServer.URL
}

func GetReqs() ([]*http.Request, []url.Values) {
	return testReqs, testValues
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

func validateAuth(key string) string {
	if !strings.Contains(key, "Bearer ") {
		return V2_INVALID_AUTH
	}
	ret := strings.Split(key, "Bearer ")
	k := ret[1]
	if k == "fake" {
		return V2_INVALID_AUTH
	}
	return ""
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
				resp = LoadSample(action)
			default:
				w.WriteHeader(http.StatusBadRequest)
				resp = []byte(HTTP_NOT_FOUND)
			}
		}
	}
	w.Write(resp)
}

func handleV2(w http.ResponseWriter, r *http.Request) {
	var resp []byte
	auth := r.Header.Get("Authorization")
	k := validateAuth(auth)
	if k != "" {
		w.WriteHeader(http.StatusBadRequest)
		resp = []byte(k)
	} else {
		type_ := "video"
		if strings.Contains(r.URL.Path, "assets") {
			type_ = "asset"
		}
		switch r.URL.Path {
		case "/v2/videos/" + V2_VIDEO_ID,
			"/v2/assets/" + ASSET_ID:
			if r.Method == "GET" || r.Method == "PUT" {
				if type_ == "asset" {
					resp = LoadSample("asset_uploaded")
				} else {
					resp = LoadSample("video2")
				}
				w.WriteHeader(http.StatusOK)
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case "/v2/videos/" + V2_VIDEO_ID + "/assets":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				resp = LoadSample("asset_list")
				w.WriteHeader(http.StatusOK)
			}
		case "/v2/videos",
			"/v2/assets":
			if r.Method != "POST" {
				resp = LoadSample(type_ + "_list")
			} else if r.Method == "POST" {
				if type_ == "video" {
					bytes, _ := ioutil.ReadAll(r.Body)
					if strings.Contains(string(bytes), "user_data") {
						resp = LoadSample("new_video2_meta")
					} else {
						resp = LoadSample("new_video2")
					}
				} else if type_ == "asset" {
					resp = LoadSample("asset_created")
					w.WriteHeader(http.StatusCreated)
				}
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	w.Write(resp)
}

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
		resp = LoadSample("aws_err.xml")
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
