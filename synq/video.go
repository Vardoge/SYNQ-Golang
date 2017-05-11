package synq

import (
	"fmt"
	"net/url"
	"time"
)

type Player struct {
	Views        int    `json:"views"`
	EmbedUrl     string `json:"embed_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
}

// Structure for Upload information needed to upload a file to Synq
type Upload struct {
	Acl         string `json:"acl"`
	Key         string `json:"key"`
	Policy      string `json:"Policy"`
	Action      string `json:"action"`
	Signature   string `json:"Signature"`
	ContentType string `json:"Content-Type"`
	AwsKey      string `json:"AWSAccessKeyId"`
}

// Sample of the video structure is located in sample/video.json
type Video struct {
	Id         string                 `json:"video_id"`
	Outputs    map[string]interface{} `json:"outputs"`
	Player     Player                 `json:"player"`
	Input      map[string]interface{} `json:"input"`
	State      string                 `json:"state"`
	Userdata   map[string]interface{} `json:"userdata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Api        *Api
	UploadInfo Upload
}

// Helper function to get details for a video, will create video object
func (a *Api) GetVideo(id string) (Video, error) {
	video := Video{}
	video.Id = id
	video.Api = a
	err := video.GetVideo()
	return video, err
}

// Calls the /v1/video/create API to create a new video object
func (a *Api) Create() (Video, error) {
	video := Video{}
	form := url.Values{}
	err := a.handlePost("create", form, &video)
	if err != nil {
		return video, err
	}
	video.Api = a
	return video, nil
}

// Calls the /v1/video/details API to load Video object information
func (v *Video) GetVideo() error {
	form := url.Values{}
	form.Add("video_id", v.Id)
	err := v.Api.handlePost("details", form, v)
	if err != nil {
		return err
	}
	return nil
}

// Calls the /v1/video/upload API to load the UploadInfo struct for the video object
func (v *Video) GetUploadInfo() error {
	form := url.Values{}
	form.Add("video_id", v.Id)
	err := v.Api.handlePost("upload", form, &v.UploadInfo)
	if err != nil {
		return err
	}
	return nil
}

// Uploads a file to the designated Upload location, this will call GetUploadInfo() if needed
func (v *Video) UploadFile(fileName string) error {
	var empty Upload
	if v.UploadInfo == empty {
		v.GetUploadInfo()
	}
	return nil
}

// Helper function to display information about a file
func (v *Video) Display() (str string) {
	if v.Id == "" {
		str = fmt.Sprintf("Empty Video\n")
	} else {
		base := "Video %s\n\tState : %s\n"
		switch v.State {
		case "uploading":
			str = fmt.Sprintf(base+"\tUpload Key : %s\n", v.UploadInfo.Key)
		case "uploaded":
			str = fmt.Sprintf(base+"\tEmbed URL : %s\n\tThumbnail : %s\n", v.Id, v.State, v.Player.EmbedUrl, v.Player.ThumbnailUrl)
		default:
			str = fmt.Sprintf(base, v.Id, v.State)
		}
	}
	return str
}
