[![CircleCI](https://circleci.com/gh/SYNQfm/SYNQ-Golang.svg?style=svg)](https://circleci.com/gh/SYNQfm/SYNQ-Golang)
[![Coverage Status](https://coveralls.io/repos/github/SYNQfm/SYNQ-Golang/badge.svg?branch=master)](https://coveralls.io/github/SYNQfm/SYNQ-Golang?branch=master)

## Introduction 

This is the Golang SDK for the SYNQ [API](https://synq.fm/docs)

## Installing
```
go get -u github.com/SYNQfm/SYNQ-Golang
```

## Usage (API v2)

Here's an example of a simple main script that uses our SDK

```golang
package main

import (
  "log"

  "github.com/SYNQfm/SYNQ-Golang/synq"
)

func main() {
  // create API using username and password
  api := synq.Login("email", "password")
  // create API using a valid token
  api = synq.NewV2("token")
  video, _ := api.GetVideo("myvideo")
  log.Printf("video returned %v", video)
}
```

V2 Video [JSON](https://github.com/SYNQfm/SYNQ-Golang/blob/master/sample/video2.json)
```javascript
{
  "data": {
    "user_data": {
      "type": "test",
      "description": "test in postman"
    },
    "metadata": {},
    "id": "9e9dc8c8-f705-41db-88da-b3034894deb9",
    "assets": [
      {
        "video_id": "9e9dc8c8-f705-41db-88da-b3034894deb9",
        "updated_at": "2017-11-16T16:37:14.547327Z",
        "type": "mp4",
        "state": "created",
        "metadata": null,
        "location": "https://s3.amazonaws.com/synq-jessica/uploads/01/82/01823629bcf24c34b714ae21e1a4647f/01823629bcf24c34b714ae21e1a4647f.mp4",
        "id": "01823629-bcf2-4c34-b714-ae21e1a4647f",
        "created_at": "2017-11-16T16:37:13.606310Z",
        "account_id": ""
      }
    ],
    "updated_at": "2017-11-14T23:43:10.540544Z",
    "created_at": "2017-11-14T23:43:10.517985Z"
  }
}
```


## Usage (API v1)

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

V1 Video [JSON](https://github.com/SYNQfm/SYNQ-Golang/blob/master/sample/video.json)
```javascript
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

```

## Usage (CLI)

You can also exercise the code via the command line using our `cli`.  View our more detailed [readme](https://github.com/SYNQfm/SYNQ-Golang/blob/master/cli/README.md)
