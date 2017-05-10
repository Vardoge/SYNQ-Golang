package synq

import (
	"fmt"
	"net/url"
	"time"
)

/*
{
  "input": {
    "url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/uploads/videos/45/d4/45d4063d00454c9fb21e5186a09c3115.mp4",
    "width": 720,
    "height": 1280,
    "duration": 17.48,
    "file_size": 16706384,
    "framerate": 29.97,
    "uploaded_at": "2017-02-15T03:05:17.978Z"
  },
  "state": "uploaded",
  "player": {
    "views": 0,
    "embed_url": "https://player.synq.fm/embed/45d4063d00454c9fb21e5186a09c3115",
    "thumbnail_url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/thumbnails/45/d4/45d4063d00454c9fb21e5186a09c3115/0000360.jpg"
  },
  "outputs": {
    "hls": {
      "url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/videos/45/d4/45d4063d00454c9fb21e5186a09c3115/hls/45d4063d00454c9fb21e5186a09c3115_hls.m3u8",
      "state": "complete"
    },
    "mp4_360": {
      "url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/videos/45/d4/45d4063d00454c9fb21e5186a09c3115/mp4_360/45d4063d00454c9fb21e5186a09c3115_mp4_360.mp4",
      "state": "complete"
    },
    "mp4_720": {
      "url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/videos/45/d4/45d4063d00454c9fb21e5186a09c3115/mp4_720/45d4063d00454c9fb21e5186a09c3115_mp4_720.mp4",
      "state": "complete"
    },
    "mp4_1080": {
      "url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/videos/45/d4/45d4063d00454c9fb21e5186a09c3115/mp4_1080/45d4063d00454c9fb21e5186a09c3115_mp4_1080.mp4",
      "state": "complete"
    },
    "webm_720": {
      "url": "https://multicdn.synq.fm/projects/0a/bf/0abfe1b849154082993f2fce77a16fd9/derivatives/videos/45/d4/45d4063d00454c9fb21e5186a09c3115/webm_720/45d4063d00454c9fb21e5186a09c3115_webm_720.webm",
      "state": "complete"
    }
  },
  "userdata": {},
  "video_id": "45d4063d00454c9fb21e5186a09c3115",
  "created_at": "2017-02-15T03:01:16.767Z",
  "updated_at": "2017-02-15T03:06:31.794Z"
}
*/

type Player struct {
	Views        int    `json:"views"`
	EmbedUrl     string `json:"embed_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
}

type Video struct {
	Id        string                 `json:"video_id"`
	Outputs   map[string]interface{} `json:"outputs"`
	Player    Player                 `json:"player"`
	Input     map[string]interface{} `json:"input"`
	State     string                 `json:"state"`
	Userdata  map[string]interface{} `json:"userdata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Api       *Api
}

// Helper function to get details for a video, will create video object
func (a *Api) GetVideo(id string) (Video, error) {
	video := Video{}
	video.Id = id
	video.Api = a
	err := video.GetVideo()
	return video, err
}

// Creates a new video
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

// get details for the video in question
func (v *Video) GetVideo() error {
	form := url.Values{}
	form.Add("video_id", v.Id)
	err := v.Api.handlePost("details", form, v)
	if err != nil {
		return err
	}
	return nil
}

func (v *Video) Upload() error {
	form := url.Values{}
	form.Add("video_id", v.Id)
	err := v.Api.handlePost("details", form, v)
	if err != nil {
		return err
	}
	return nil
}

func (v *Video) Display() (str string) {
	if v.Id == "" {
		str = fmt.Sprintf("Empty Video\n")
	} else {
		switch v.State {
		case "uploaded":
			str = fmt.Sprintf("Video %s\n\tState : %s\n\tEmbed URL : %s\n\tThumbnail : %s\n", v.Id, v.State, v.Player.EmbedUrl, v.Player.ThumbnailUrl)
		default:
			str = fmt.Sprintf("Video %s\n\tState : %s\n", v.Id, v.State)
		}
	}
	return str
}
