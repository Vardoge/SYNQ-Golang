package main

import (
	"flag"
	"log"

	"github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
	var (
		api_key  = flag.String("api_key", "", "pass the synq api key")
		video_id = flag.String("video_id", "", "pass in the video id to get data about")
	)
	flag.Parse()
	video := synq.New(*api_key)
	if *video_id != "" {
		log.Printf("getting video %s\n", video_id)
		video.Id = *video_id
		err := video.Details()
		if err != nil {
			log.Fatalln("error getting video details", err.Error())
		}
		log.Println("video %v\n", video)
	}
}
