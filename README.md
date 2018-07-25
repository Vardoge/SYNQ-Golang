[![CircleCI](https://circleci.com/gh/SYNQfm/SYNQ-Golang.svg?style=svg)](https://circleci.com/gh/SYNQfm/SYNQ-Golang)
[![Coverage Status](https://coveralls.io/repos/github/SYNQfm/SYNQ-Golang/badge.svg?branch=master)](https://coveralls.io/github/SYNQfm/SYNQ-Golang?branch=master)

## Introduction 

This is the Golang SDK for the SYNQ [API](https://docs.synq.fm)

## Installing
```
go get -u github.com/SYNQfm/SYNQ-Golang
```

## Usage

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

Video [JSON](https://github.com/SYNQfm/SYNQ-Golang/blob/master/sample/41101458-bc49-40db-badc-1b480831b79b.json)
```javascript
{
    "data": {
        "user_data": {},
        "updated_at": "2018-03-20T23:50:22.119546Z",
        "metadata": {
            "type": "movie",
            "title": {
                "original": {
                    "content": "Tears of Steel"
                },
                "nor": {
                    "content": "Tears of Steel"
                }
            },
            "series": {},
            "regional_content": false,
            "production_year": 2012,
            "parental_rating": "10",
            "metadata_version": "1.0",
            "genres": [
                "Sci-Fi"
            ],
            "expected_duration": "00:12:14:00",
            "description": {
                "nor": {
                    "content-tiny": "Thom just wanted to be an astronaut.",
                    "content-short": "Thom just wanted to be an astronaut. His girlfriend Celia just wanted to create robots - and for him to not be freaked out by her cyborg hand.",
                    "content-medium": "Thom just wanted to be an astronaut. His girlfriend Celia just wanted to create robots - and for him to not be freaked out by her cyborg hand. How was Thom supposed to know that breaking up with her would make her take out her anger on the rest of humanity using her robots...",
                    "content-long": "Thom just wanted to be an astronaut. His girlfriend Celia just wanted to create robots - and for him to not be freaked out by her cyborg hand. How was Thom supposed to know that breaking up with her would make her take out her anger on the rest of humanity using her robots? It seems the only possible way of undoing everything...is to overwrite her memory of what happened 40 years ago."
                }
            },
            "credits": [
                {
                    "role": "actor",
                    "name": "Derek de Lint"
                },
                {
                    "role": "actor",
                    "name": "Sergio Hasselbaink"
                },
                {
                    "role": "actor",
                    "name": "Rogier Schippers"
                },
                {
                    "role": "actor",
                    "name": "Vanja Rukavina"
                },
                {
                    "role": "actor",
                    "name": "Denise Rebergen"
                },
                {
                    "role": "actor",
                    "name": "Jody Bhe"
                },
                {
                    "role": "actor",
                    "name": "Chris Haley"
                }
            ],
            "country_of_origin": [
                "UK"
            ],
            "aspect_ratio": "16:9"
        },
        "id": "41101458-bc49-40db-badc-1b480831b79b",
        "created_at": "2018-03-15T22:34:32.433479Z",
        "assets": [
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/40cf3756-6fda-4270-bb93-c05730d32686-en.ttml",
                "updated_at": "2018-03-21T18:53:25.858595Z",
                "type": "subtitles",
                "state": "completed",
                "metadata": {
                    "format": "http://www.w3.org/ns/ttml"
                },
                "location": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/40cf3756-6fda-4270-bb93-c05730d32686-en.ttml",
                "id": "40cf3756-6fda-4270-bb93-c05730d32686",
                "created_at": "2018-03-21T18:50:58.845768Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/70194ab4-6b05-4ee1-8a1f-4daa00bee3af-no.ttml",
                "updated_at": "2018-03-21T19:00:09.079683Z",
                "type": "subtitles",
                "state": "completed",
                "metadata": {
                    "format": "http://www.w3.org/ns/ttml"
                },
                "location": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/70194ab4-6b05-4ee1-8a1f-4daa00bee3af-no.ttml",
                "id": "70194ab4-6b05-4ee1-8a1f-4daa00bee3af",
                "created_at": "2018-03-21T18:59:05.296735Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/cd1dbe90-4064-4dae-8e06-3b4d146023c9-es.ttml",
                "updated_at": "2018-03-21T20:19:15.637034Z",
                "type": "subtitles",
                "state": "completed",
                "metadata": {
                    "format": "http://www.w3.org/ns/ttml"
                },
                "location": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/cd1dbe90-4064-4dae-8e06-3b4d146023c9-es.ttml",
                "id": "cd1dbe90-4064-4dae-8e06-3b4d146023c9",
                "created_at": "2018-03-21T20:18:39.553785Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_270p_2.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=e1eeaaf90734e5a4486aa2e7ee74f7e6840fddd222b4ef145d90b4833112751f",
                "updated_at": "2018-03-20T19:30:43.759247Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 480,
                    "video_framerate": 12,
                    "video_codec": "h264",
                    "size": 19218899,
                    "height": 270,
                    "format": "iso5",
                    "duration": 734.083,
                    "dar": 1.778,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 0,
                    "audio_codec": ""
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_270p_2.mp4",
                "id": "47106242-db7c-42da-b305-095c52ca3d2d",
                "created_at": "2018-03-20T19:08:01.881762Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/41101458bc4940dbbadc1b480831b79b.mov?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=cdf37f81a33958653d09b2386f7f256c827e0a9da897bc825ed6d25fff2c25e5",
                "updated_at": "2018-03-20T19:03:16.093760Z",
                "type": "source",
                "state": "completed",
                "metadata": {
                    "width": 3840,
                    "video_framerate": 24,
                    "video_codec": "h264",
                    "size": 6737592810,
                    "height": 1714,
                    "format": "mp4",
                    "duration": 734,
                    "dar": 2.24,
                    "content_type": "",
                    "audio_framerate": 44100,
                    "audio_codec": "aac_lc"
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/41101458bc4940dbbadc1b480831b79b.mov",
                "id": "44c01dd9-a481-4d36-95de-0155033cc8e2",
                "created_at": "2018-03-20T18:45:48.738546Z",
                "account_id": "83944a4f-1cb0-4e1b-bb03-70f1afc1d6a8"
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_audio.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=3402b38e39bde1a28b2602ee75d97bcc82629c5c25a695f2120ec84c2272bee7",
                "updated_at": "2018-03-20T19:30:43.215291Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 0,
                    "video_framerate": 0,
                    "video_codec": "",
                    "size": 11900493,
                    "height": 0,
                    "format": "iso5",
                    "duration": 734.043,
                    "dar": 0,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 48000,
                    "audio_codec": "aac_lc"
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_audio.mp4",
                "id": "558d03f6-58c9-472e-80a8-c145c4b98c0a",
                "created_at": "2018-03-20T19:08:02.404957Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_540p.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=d3a8a3d692252b669c184bfa462d92b5fbbfb7cd642c6e22b7a6bc9a00fe3ad5",
                "updated_at": "2018-03-20T19:30:41.685063Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 960,
                    "video_framerate": 24,
                    "video_codec": "h264",
                    "size": 120543630,
                    "height": 540,
                    "format": "iso5",
                    "duration": 734,
                    "dar": 1.778,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 0,
                    "audio_codec": ""
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_540p.mp4",
                "id": "286227ae-2800-4243-9c27-c80f1ccfc851",
                "created_at": "2018-03-20T19:08:05.728673Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_270p.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=dc6474bae97afc994023a138b7258cad28e5b4edacd895d146954107868c0055",
                "updated_at": "2018-03-20T19:30:40.544090Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 480,
                    "video_framerate": 12,
                    "video_codec": "h264",
                    "size": 30950462,
                    "height": 270,
                    "format": "iso5",
                    "duration": 734.083,
                    "dar": 1.778,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 0,
                    "audio_codec": ""
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_270p.mp4",
                "id": "e64565e6-409a-4c80-85a8-da2f789fedb1",
                "created_at": "2018-03-20T19:08:01.353119Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_360p.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=f9669cd90ccbb00d99edf895fec3adc96fe001f0be91d24abb9e6fb632016407",
                "updated_at": "2018-03-20T19:30:41.092669Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 640,
                    "video_framerate": 24,
                    "video_codec": "h264",
                    "size": 61509707,
                    "height": 360,
                    "format": "iso5",
                    "duration": 734,
                    "dar": 1.778,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 0,
                    "audio_codec": ""
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_360p.mp4",
                "id": "9466fe8e-ac3c-40cf-9ade-403c9f4298c0",
                "created_at": "2018-03-20T19:08:00.809547Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_720p.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=e97e949be97980379bcbc70f288d5049dc5c820c060633d18b50183bd7ba6bd3",
                "updated_at": "2018-03-20T19:30:42.199285Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 1280,
                    "video_framerate": 24,
                    "video_codec": "h264",
                    "size": 246015788,
                    "height": 720,
                    "format": "iso5",
                    "duration": 734,
                    "dar": 1.778,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 0,
                    "audio_codec": ""
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_720p.mp4",
                "id": "9f0173ef-aa54-4130-9f94-2f9ae71f201d",
                "created_at": "2018-03-20T19:08:05.070094Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn-eu.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_1080p.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJ35NCRJNRP2WL3SA%2F20180720%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20180720T171232Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=25f94f684382711146ca5355b913f514850a34d0ba671c9b85f9ec231ef7f3db",
                "updated_at": "2018-03-20T19:30:42.723844Z",
                "type": "fmp4",
                "state": "completed",
                "metadata": {
                    "width": 1920,
                    "video_framerate": 24,
                    "video_codec": "h264",
                    "size": 507837786,
                    "height": 1080,
                    "format": "iso5",
                    "duration": 734,
                    "dar": 1.778,
                    "content_type": "application/octet-stream",
                    "audio_framerate": 0,
                    "audio_codec": ""
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/fmp4/41101458-bc49-40db-badc-1b480831b79b_Layer_1080p.mp4",
                "id": "7ea0143f-2b86-4066-a825-23d39db5b4bb",
                "created_at": "2018-03-20T19:08:04.276364Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://synq-player-zscvoibvg1af.stackpathdns.com/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/dash/signed_manifest.mpd",
                "updated_at": "2018-03-20T19:48:40.464188Z",
                "type": "dash",
                "state": "completed",
                "metadata": {
                    "layers": [
                        {
                            "width": 480,
                            "id": "47106242-db7c-42da-b305-095c52ca3d2d",
                            "height": 270,
                            "bitrate": 209446.6
                        },
                        {
                            "id": "558d03f6-58c9-472e-80a8-c145c4b98c0a",
                            "bitrate": 129698.05
                        },
                        {
                            "width": 960,
                            "id": "286227ae-2800-4243-9c27-c80f1ccfc851",
                            "height": 540,
                            "bitrate": 1313827.03
                        },
                        {
                            "width": 480,
                            "id": "e64565e6-409a-4c80-85a8-da2f789fedb1",
                            "height": 270,
                            "bitrate": 337296.59
                        },
                        {
                            "width": 640,
                            "id": "9466fe8e-ac3c-40cf-9ade-403c9f4298c0",
                            "height": 360,
                            "bitrate": 670405.53
                        },
                        {
                            "width": 1280,
                            "id": "9f0173ef-aa54-4130-9f94-2f9ae71f201d",
                            "height": 720,
                            "bitrate": 2681370.99
                        },
                        {
                            "width": 1920,
                            "id": "7ea0143f-2b86-4066-a825-23d39db5b4bb",
                            "height": 1080,
                            "bitrate": 5535016.74
                        }
                    ],
                    "duration": 734.083
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/dash/master_manifest.mpd",
                "id": "be132343-e9b9-49ca-8022-a579f8da1954",
                "created_at": "2018-03-20T19:31:28.072369Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://synq-player-zscvoibvg1af.stackpathdns.com/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/hls/signed_manifest.m3u8",
                "updated_at": "2018-03-20T19:49:15.647278Z",
                "type": "hls",
                "state": "completed",
                "metadata": {
                    "layers": [
                        {
                            "width": 480,
                            "id": "47106242-db7c-42da-b305-095c52ca3d2d",
                            "height": 270,
                            "bitrate": 209446.6
                        },
                        {
                            "id": "558d03f6-58c9-472e-80a8-c145c4b98c0a",
                            "bitrate": 129698.05
                        },
                        {
                            "width": 960,
                            "id": "286227ae-2800-4243-9c27-c80f1ccfc851",
                            "height": 540,
                            "bitrate": 1313827.03
                        },
                        {
                            "width": 480,
                            "id": "e64565e6-409a-4c80-85a8-da2f789fedb1",
                            "height": 270,
                            "bitrate": 337296.59
                        },
                        {
                            "width": 640,
                            "id": "9466fe8e-ac3c-40cf-9ade-403c9f4298c0",
                            "height": 360,
                            "bitrate": 670405.53
                        },
                        {
                            "width": 1280,
                            "id": "9f0173ef-aa54-4130-9f94-2f9ae71f201d",
                            "height": 720,
                            "bitrate": 2681370.99
                        },
                        {
                            "width": 1920,
                            "id": "7ea0143f-2b86-4066-a825-23d39db5b4bb",
                            "height": 1080,
                            "bitrate": 5535016.74
                        }
                    ],
                    "duration": 734.083
                },
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/hls/master_manifest.m3u8",
                "id": "de15f049-570b-465a-b5e3-96b3b0de4654",
                "created_at": "2018-03-20T19:32:09.224649Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://synq-player-zscvoibvg1af.stackpathdns.com/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/smooth/signed_manifest.ism/manifest",
                "updated_at": "2018-03-20T20:08:12.995477Z",
                "type": "smooth",
                "state": "completed",
                "metadata": null,
                "location": "s3://synq-frankfurt/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/smooth/manifest.ism",
                "id": "560a6ef0-9067-42c6-8e62-0f53f1c8bcdd",
                "created_at": "2018-03-20T19:46:29.933076Z",
                "account_id": null
            },
            {
                "video_id": "41101458-bc49-40db-badc-1b480831b79b",
                "url": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/thumbnails/ff679a75-bbc6-4c70-814d-ea97ed8e43cf.jpg",
                "updated_at": "2018-03-20T20:49:37.619508Z",
                "type": "thumbnail",
                "state": "completed",
                "metadata": {},
                "location": "https://multicdn.synq.fm/videos/41/10/41101458-bc49-40db-badc-1b480831b79b/thumbnails/ff679a75-bbc6-4c70-814d-ea97ed8e43cf.jpg",
                "id": "ff679a75-bbc6-4c70-814d-ea97ed8e43cf",
                "created_at": "2018-03-20T20:46:29.108781Z",
                "account_id": null
            }
        ],
        "account_ids": [
            "1cf77745-dcdf-4ec1-8c83-cf92f0fb304d",
            "716411e9-2101-4d12-a565-04a61c73980e",
            "d2aa14aa-a32b-437c-ae34-7d47d95e28d9",
            "95130d23-9893-4dd7-be9b-3ed742ec314b",
            "83944a4f-1cb0-4e1b-bb03-70f1afc1d6a8",
            "e4080e7b-3c0f-4981-9b80-4fb83e23777b",
            "cf17ce25-eed4-4dd0-bd3e-95b820e92ada",
            "874150c0-3032-4de5-b3e1-4852e30f7b7f",
            "4c84d2e2-faed-480c-b447-99c52f0cd4c4",
            "8c9c4189-825d-4d5b-9e47-11469fa14d75",
            "db0de2e2-b812-4c99-a6e2-d9bf48b86e42",
            "fcc90ae1-601a-4249-a562-ec4c1a52d53a"
        ]
    }
}
```

### Utilizing the testing framework

There's a pretty powerful mocked server in test_server/server.go which can be used for testing your service connected to the SDK.  Here's an example of how to use it

```golang
```

## Usage (CLI)

You can also exercise the code via the command line using our `cli`.  View our more detailed [readme](https://github.com/SYNQfm/SYNQ-Golang/blob/master/cli/README.md)
