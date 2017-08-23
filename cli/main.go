package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/SYNQfm/SYNQ-Golang/synq"
)

func handleError(err error) {
	if err != nil {
		log.Printf("Error : %s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	var err error
	var video synq.Video
	var (
		c = flag.String("command", "", "one of: details, upload_info, upload, create, uploader_info, uploader, query or create_and_then_multipart_upload")
		a = flag.String("api_key", "", "pass the synq api key")
		v = flag.String("video_id", "", "pass in the video id to get data about")
		f = flag.String("file", "", "path to file you want to upload or userdata")
		q = flag.String("query", "", "query string to use")
	)
	flag.Parse()
	cmd := *c
	vid := *v
	api_key := *a
	file := *f
	if api_key == "" {
		log.Println("missing 'api_key'")
		os.Exit(1)
	}
	api := synq.New(api_key)
	switch cmd {
	case "details":
		if vid == "" {
			log.Println("missing video id")
			os.Exit(1)
		}
		log.Printf("getting video %s\n", vid)
		video, err = api.GetVideo(vid)
	case "upload_info":
		log.Printf("Getting upload info for %s\n", vid)
		video.Api = &api
		video.Id = vid
		err = video.GetUploadInfo()
	case "query":
		videos, err := api.Query(*q)
		handleError(err)
		log.Printf("Found %d videos\n", len(videos))
		for _, video := range videos {
			log.Printf(video.Display())
		}
		os.Exit(0)
	case "upload":
		if file == "" {
			log.Println("missing 'file'")
			os.Exit(1)
		}
		log.Printf("uploading file '%s'\n", file)
		video.Api = &api
		video.Id = vid
		err = video.UploadFile(file)
	case "create":
		log.Printf("Creating new video")
		if file != "" {
			log.Printf("loading userdata file from %s\n", file)
			bytes, err := ioutil.ReadFile(file)
			if err == nil {
				userdata := make(map[string]interface{})
				json.Unmarshal(bytes, &userdata)
				video, err = api.Create(userdata)
			}
		} else {
			video, err = api.Create()
		}
	case "uploader_info":
		if vid == "" {
			log.Println("missing video id")
			os.Exit(1)
		}
		video.Api = &api
		video.Id = vid
		err = video.GetUploaderInfo()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		log.Println("uploader_url:", video.UploaderInfo["uploader_url"])
		os.Exit(0)
	case "uploader":
		if file == "" {
			log.Println("missing 'file'")
			os.Exit(1)
		}
		if vid == "" {
			log.Println("missing video id")
			os.Exit(1)
		}
		video.Api = &api
		video.Id = vid

		log.Printf("uploading file '%s'\n", file)
		err = video.MultipartUpload(file)
		handleError(err)

		video, err = api.GetVideo(video.Id)
	case "create_and_then_multipart_upload":
		if file == "" {
			log.Println("missing 'file'")
			os.Exit(1)
		}

		log.Printf("Creating new video")
		video, err = api.Create()
		handleError(err)

		log.Printf("uploading file '%s'\n", file)
		err = video.MultipartUpload(file)
		handleError(err)

		video, err = api.GetVideo(video.Id)
	default:
		err = errors.New("unknown command '" + cmd + "'")
	}
	handleError(err)
	log.Printf(video.Display())
}
