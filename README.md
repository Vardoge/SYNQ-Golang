[![CircleCI](https://circleci.com/gh/SYNQfm/SYNQ-Golang.svg?style=svg)](https://circleci.com/gh/SYNQfm/SYNQ-Golang)
[![Coverage Status](https://coveralls.io/repos/github/SYNQfm/SYNQ-Golang/badge.svg?branch=master)](https://coveralls.io/github/SYNQfm/SYNQ-Golang?branch=master)

## Introduction 

This is the Golang SDK for the SYNQ [API](https://synq.fm/docs)

## Installing
```
go get -u github.com/SYNQfm/SYNQ-Golang
```

## Usage (API)

Here's an example of a simple main script that uses our SDK

```golang
package main

import (
  "log"

  "github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
  api := synq.New("myapikey")
  video, _ := api.GetVideo("myvideo")
  log.Printf("video returned %v", video)
}
```

## Usage (CLI)

You can also exercise the code via the command line using our `cli`

Here's the build and usage
```
cd cli
go build
./cli -h

Usage of ./cli:
  -api_key string
      pass the synq api key
  -command string
      one of: details, upload_info, upload, create, uploader_info, uploader, query or create_and_then_multipart_upload
  -file string
      path to file you want to upload or userdata
  -query string
      query string to use
  -video_id string
      pass in the video id to get data about
```

```bash
# Create a new video object
./cli -api_key=<key> -command create
# Upload a file
./cli -api_key=<key> -video_id=<vid> -file <file name> -command upload
# Get details for a video
./cli -api_key=<key> -video_id=<vid> -command details
```
