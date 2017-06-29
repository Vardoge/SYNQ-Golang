package synq

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Player struct {
	Views        int    `json:"views"`
	EmbedUrl     string `json:"embed_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
}

// Structure for Upload information needed to upload a file to Synq
type Upload map[string]string

// Structure for Uploader information needed to upload a file to Synq in
// multiple parts
type Uploader map[string]string

// Sample of the video structure is located in sample/video.json
type Video struct {
	Id           string                 `json:"video_id"`
	Outputs      map[string]interface{} `json:"outputs"`
	Player       Player                 `json:"player"`
	Input        map[string]interface{} `json:"input"`
	State        string                 `json:"state"`
	Userdata     map[string]interface{} `json:"userdata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Api          *Api
	UploadInfo   Upload
	UploadResp   interface{}
	UploaderInfo Uploader
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

func (u Upload) valid() bool {
	return u["key"] != ""
}

func (u Upload) setURL(url string) {
	u["action"] = url
}

func (u Upload) dstFileName() string {
	return u["key"]
}

func (u Upload) url() string {
	return u["action"]
}

func (u Upload) createUploadReq(fileName string) (req *http.Request, err error) {
	if !u.valid() {
		return req, errors.New("no valid upload data")
	}
	f, err := os.Open(fileName)
	defer f.Close()
	if os.IsNotExist(err) {
		return req, errors.New("file '" + fileName + "' does not exist")
	}
	// add file last
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// add fields
	for key, value := range u {
		if key == "action" {
			continue
		}
		fw, err := w.CreateFormField(key)
		if err != nil {
			return req, err
		}
		if _, err = fw.Write([]byte(value)); err != nil {
			return req, err
		}
	}
	// file last
	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return req, err
	}
	if _, err = io.Copy(fw, f); err != nil {
		return req, err
	}
	w.Close()

	req, err = http.NewRequest("POST", u.url(), &b)
	if err != nil {
		return req, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func (u Uploader) valid() bool {
	return u["uploader_url"] != ""
}

func (u Uploader) url() string {
	return u["uploader_url"]
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
	if v.UploadInfo.valid() {
		log.Println("Upload Info already loaded, skipping")
		return nil
	}
	form := url.Values{}
	form.Add("video_id", v.Id)
	err := v.Api.handlePost("upload", form, &v.UploadInfo)
	if err != nil {
		return err
	}
	return nil
}

// Uploader returns information needed to use the embeddable web uploader:
// https://www.synq.fm/video-uploader/
// This procedure calls the remote procedure "/v1/video/uploader".
//
// Example:
//     var err error
//     var video synq.Video
//
//     api := synq.New("00000000000000000000000000000000")
//
//     // Create a new video object.
//     video, err = api.Create()
//     if err != nil {
//             log.Println(err)
//             os.Exit(-1)
//     }
//
//     // Get information needed to upload a file in multiple parts.
//     err = video.GetUploaderInfo()
//     if err != nil {
//             log.Println(err)
//             os.Exit(-1)
//     }
func (v *Video) GetUploaderInfo() error {
	if v.UploaderInfo.valid() {
		log.Println("Uploader Info already loaded, skipping")
		return nil
	}
	form := url.Values{}
	form.Add("video_id", v.Id)
	err := v.Api.handlePost("uploader", form, &v.UploaderInfo)
	if err != nil {
		return err
	}
	return nil
}

// Uploads a file to the designated Upload location, this will call GetUploadInfo() if needed
func (v *Video) UploadFile(fileName string) error {
	if err := v.GetUploadInfo(); err != nil {
		log.Println("failed to getUploadInfo()")
		return err
	}
	// now use the UploadInfo to upload the specific file
	req, err := v.UploadInfo.createUploadReq(fileName)
	if err != nil {
		log.Println("failed to create upload req")
		return err
	}
	if err := v.Api.handleUploadReq(req, &v.UploadResp); err != nil {
		log.Println("failed to call handleUploadReq")
		return err
	}
	return nil
}

// MultipartUpload uploads a file as the video's original_file.
// This procedure will use Amazon S3's Multipart Upload API:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/uploadobjusingmpu.html
//
// Example:
//     var err error
//     var video synq.Video
//
//     api := synq.New("00000000000000000000000000000000")
//
//     // Create a new video object.
//     video, err = api.Create()
//     if err != nil {
//             log.Println(err)
//             os.Exit(-1)
//     }
//
//     // Upload the file video.mp4 in multiple parts.
//     err = video.MultipartUpload("video.mp4")
//     if err != nil {
//             log.Println(err)
//             os.Exit(-1)
//     }
func (v *Video) MultipartUpload(fileName string) error {
	if err := v.GetUploaderInfo(); err != nil {
		log.Println("failed to getUploaderInfo()")
		return err
	}

	// NOTE(mastensg): getUploadInfo for parameters: acl, bucket, contentType, key
	if err := v.GetUploadInfo(); err != nil {
		log.Println("failed to getUploadInfo()")
		return err
	}

	f, err := os.Open(fileName)
	defer f.Close()
	if os.IsNotExist(err) {
		return errors.New("file '" + fileName + "' does not exist")
	}

	// extract upload action url
	actionURL := v.UploadInfo.url()

	// extract uploader url
	uploaderURL := v.UploaderInfo.url()

	// multipartUploadWithUploaderURL
	acl := v.UploadInfo["acl"]
	awsAccessKeyId := v.UploadInfo["AWSAccessKeyId"]
	contentType := v.UploadInfo["Content-Type"]
	key := v.UploadInfo["key"]
	if _, err := multipartUpload(f, acl, actionURL, awsAccessKeyId, contentType, key, uploaderURL, v.Id); err != nil {
		return err
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
			str = fmt.Sprintf(base, v.Id, v.State)
		case "uploaded":
			str = fmt.Sprintf(base+"\tEmbed URL : %s\n\tThumbnail : %s\n", v.Id, v.State, v.Player.EmbedUrl, v.Player.ThumbnailUrl)
		default:
			str = fmt.Sprintf(base, v.Id, v.State)
		}
	}
	return str
}
