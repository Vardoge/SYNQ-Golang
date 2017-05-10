package main

import (
	"flag"
	"log"
	"os"

	"github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
	var (
		c = flag.String("command", "details", "pass in command")
		a = flag.String("api_key", "", "pass the synq api key")
		v = flag.String("video_id", "", "pass in the video id to get data about")
	)
	flag.Parse()
	cmd := *c
	vid := *v
	api_key := *a
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
		video, err := api.GetVideo(vid)
		if err != nil {
			log.Println("error getting video details", err.Error())
			os.Exit(-1)
		}
		log.Printf("video %v\n", video)
	default:
		log.Println("unknown command " + cmd)
		os.Exit(-1)
	}
}
