package synq

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_helper"
	"github.com/stretchr/testify/assert"
)

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

func TestGetUploadInfo(t *testing.T) {
	assert := assert.New(t)
	api := setupTestApi("fake")
	video := Video{Id: testVideoId2V1, Api: &api}
	err := video.GetUploadInfo()
	assert.NotNil(err)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", err.Error())
	api.SetKey(testApiKeyV1)
	err = video.GetUploadInfo()
	assert.Nil(err)
	assert.NotEmpty(video.UploadInfo)
	assert.Len(video.UploadInfo, 7)
	assert.Equal(uploadKey, video.UploadInfo["key"])
	assert.Equal("public-read", video.UploadInfo["acl"])
	assert.Equal("https://synqfm.s3.amazonaws.com", video.UploadInfo.url())
	assert.Equal("video/mp4", video.UploadInfo["Content-Type"])
	assert.Equal(uploadKey, video.UploadInfo.dstFileName())
}

func TestCreateUploadReqErr(t *testing.T) {
	assert := assert.New(t)
	uploadInfo := make(Upload)
	valid_file := "video.go"
	_, err := uploadInfo.createUploadReq(valid_file)
	assert.NotNil(err)
	assert.Equal("no valid upload data", err.Error())
	uploadInfo["action"] = ":://noprotocol.com"
	uploadInfo["key"] = "fake"
	_, err = uploadInfo.createUploadReq(valid_file)
	assert.NotNil(err)
	assert.Equal("parse :://noprotocol.com: missing protocol scheme", err.Error())
	uploadInfo["action"] = "http://valid.com"
	_, err = uploadInfo.createUploadReq("myfile.mp4")
	assert.NotNil(err)
	assert.Equal("file 'myfile.mp4' does not exist", err.Error())
}

func TestCreateUploadReq(t *testing.T) {
	assert := assert.New(t)
	valid_file := "video.go"
	var uploadInfo Upload
	data, err := ioutil.ReadFile("../sample/upload.json")
	assert.Nil(err)
	err = json.Unmarshal(data, &uploadInfo)
	assert.Nil(err)
	req, err := uploadInfo.createUploadReq(valid_file)
	assert.Nil(err)
	assert.NotEmpty(req.Header)
	assert.Contains(req.Header.Get("Content-Type"), "multipart/form-data")
	f, h, err := req.FormFile("file")
	assert.Nil(err)
	assert.Equal(valid_file, h.Filename)
	assert.Nil(err)
	src, _ := ioutil.ReadFile(valid_file)
	b := make([]byte, len(src))
	f.Read(b)
	assert.Equal(src, b)
	assert.Len(req.PostForm, 6)
	for key, value := range uploadInfo {
		if key == "action" {
			assert.Equal("", req.PostFormValue(key))
		} else {
			assert.Equal(value, req.PostFormValue(key))
		}
	}
}

func TestUploadFile(t *testing.T) {
	assert := assert.New(t)
	aws := test_helper.S3Stub()
	api := setupTestApi("fake")
	defer aws.Close()
	video := Video{Id: testVideoId2V1, Api: &api}
	valid_file := "video.go"
	err := video.UploadFile(valid_file)
	assert.NotNil(err)
	assert.Equal("Invalid uuid. Example: '1c0e3ea4529011e6991554a050defa20'.", err.Error())
	api.SetKey(testApiKeyV1)
	err = video.UploadFile("myfile.mp4")
	assert.NotNil(err)
	assert.Equal("file 'myfile.mp4' does not exist", err.Error())
	video.UploadInfo.setURL(aws.URL)
	err = video.UploadFile(valid_file)
	assert.Nil(err)
	// use an invalid key and it should return an error
	video.UploadInfo["key"] = "fakekey"
	err = video.UploadFile(valid_file)
	assert.NotNil(err)
	assert.Equal("At least one of the pre-conditions you specified did not hold", err.Error())

}
