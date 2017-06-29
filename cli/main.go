package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
	var err error
	var video synq.Video
	var (
		c = flag.String("command", "one of: create_and_then_multipart_upload, details, uploader_info", "pass in command")
		a = flag.String("api_key", "", "pass the synq api key")
		v = flag.String("video_id", "", "pass in the video id to get data about")
		f = flag.String("file", "", "path to file you want to upload")
	)
	flag.Parse()
	cmd := *c
	vid := *v
	api_key := *a
	file := *f
	if api_key == "" {
		log.Println("missing api_key")
		os.Exit(-1)
	}
	api := synq.New(api_key)
	switch cmd {
	case "details":
		if vid == "" {
			log.Println("missing video id")
			os.Exit(-1)
		}
		log.Printf("getting video %s\n", vid)
		video, err = api.GetVideo(vid)
	case "upload_info":
		log.Printf("Getting upload info for %s\n", vid)
		video.Api = &api
		video.Id = vid
		err = video.GetUploadInfo()
	case "upload":
		if file == "" {
			log.Println("missing 'file'")
			os.Exit(-1)
		}
		log.Printf("uploading file '%s'\n", file)
		video.Api = &api
		video.Id = vid
		err = video.UploadFile(file)
	case "create":
		log.Printf("Creating new video")
		video, err = api.Create()
	case "uploader_info":
		if vid == "" {
			log.Println("missing video id")
			os.Exit(-1)
		}
		video.Api = &api
		video.Id = vid
		err = video.GetUploaderInfo()
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}
		log.Println("uploader_url:", video.UploaderInfo["uploader_url"])
		os.Exit(0)
	case "uploader":
		if file == "" {
			log.Println("missing 'file'")
			os.Exit(-1)
		}
		if vid == "" {
			log.Println("missing video id")
			os.Exit(-1)
		}
		video.Api = &api
		video.Id = vid

		log.Printf("uploading file '%s'\n", file)
		err = video.MultipartUpload(file)
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}

		video, err = api.GetVideo(video.Id)
	case "create_and_then_multipart_upload":
		if file == "" {
			log.Println("missing 'file'")
			os.Exit(-1)
		}

		log.Printf("Creating new video")
		video, err = api.Create()
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}

		log.Printf("uploading file '%s'\n", file)
		err = video.MultipartUpload(file)
		if err != nil {
			log.Println(err)
			os.Exit(-1)
		}

		video, err = api.GetVideo(video.Id)
	default:
		err = errors.New("unknown command " + cmd)
	}
	if err != nil {
		log.Printf("Error : %s\n", err.Error())
		os.Exit(-1)
	}
	log.Printf(video.Display())
}
