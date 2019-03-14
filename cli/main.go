package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/helper"
	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/SYNQfm/helpers/common"
)

var cli common.Cli

func init() {
	cli = common.NewCli()
	cli.DefaultSetup("for v2 'upload', get_video', for v1 : details, upload_info, upload, create, uploader_info, uploader, query or create_and_then_multipart_upload", "upload")
	cli.String("version", "v2", "version to use")
	cli.String("upload_url", synq.DEFAULT_UPLOADER_URL, "upload url to use")
	cli.String("video_id", "", "video id to access")
	cli.String("asset_id", "", "asset id to access")
	cli.String("file", "", "path to file you want to upload or userdata")
	cli.String("query", "", "query string to use")
	cli.String("cred_file", "", "credential file to use")
	cli.Parse()
}

func handleError(err error) {
	if err != nil {
		log.Printf("Error : %s\n", err.Error())
		os.Exit(1)
	}
}

func handleV2(api synq.ApiV2) {
	vid := cli.GetString("video_id")
	aid := cli.GetString("asset_id")
	ret := common.NewRet(cli.Command)
	switch cli.Command {
	case "upload":
		var asset synq.Asset
		var err error
		upload_url := cli.GetString("upload_url")
		if upload_url == "" {
			upload_url = synq.DEFAULT_UPLOADER_URL
		}
		log.Printf("using uploader url %s\n", upload_url)
		api.UploadUrl = upload_url
		file := cli.GetString("file")
		if file == "" {
			err = errors.New("file missing")
			handleError(err)
		}
		ext := filepath.Ext(file)
		ctype := common.ExtToCtype(ext)
		if ctype == "" {
			handleError(errors.New("can not find ctype for " + ext))
		}
		if aid == "" {
			video, err := helper.LoadVideoV2(vid, cli, api)
			if err == nil {
				var found synq.Asset
				for _, a := range video.Assets {
					if ctype == a.Type {
						found = a
						break
					}
				}
				if found.Id != "" {
					log.Printf("using existing asset %s for '%s'\n", found.Id, ctype)
					asset = found
				} else {
					log.Printf("creating new asset with ctype '%s'\n", ctype)
					if !cli.Simulate {
						req := upload.UploadRequest{
							ContentType: ctype,
						}
						asset, err = video.CreateAssetForUpload(req)
						common.PurgeFromCache(vid, cli)
					}
				}
			}
		} else {
			log.Printf("getting existing asset %s\n", aid)
			asset, err = helper.LoadAsset(aid, cli, api)
		}
		handleError(err)
		if !cli.Simulate {
			acl := ""
			typ := ""
			if strings.Contains(ctype, "video") {
				acl = "private"
				typ = "source"
			} else {
				acl = "public-read"
				typ = "metadata"
			}
			params := upload.UploadRequest{
				ContentType: ctype,
				AssetId:     asset.Id,
				Acl:         acl,
				Type:        typ,
			}
			up, e := helper.LoadUploadParameters(asset.VideoId, params, cli, api)
			handleError(e)
			log.Printf("Got upload params for %s", up.Key)
			asset.UploadParameters = up
		}

		cli.Printf("uploading file %s\n", file)
		if !cli.Simulate {
			err = asset.UploadFile(file)
			handleError(err)
			log.Printf("uploaded file %s\n", file)
		}
	case "get_raw_videos",
		"get_videos":
		api.PageSize = 500
		raw := strings.Contains(cli.Command, "raw")
		str := fmt.Sprintf("getting all videos (page size %d)", api.PageSize)
		name := "videos"
		if raw {
			str = str + " (raw format)"
			name = name + "_raw"
		}
		var bytes []byte
		var err error
		vidCt := 0
		log.Println(str)
		if raw {
			var videos []json.RawMessage
			videos, err = api.GetRawVideos("")
			vidCt = len(videos)
			bytes, _ = json.Marshal(videos)
		} else {
			var videos []synq.VideoV2
			videos, err = api.GetVideos("")
			bytes, _ = json.Marshal(videos)
		}
		handleError(err)
		log.Printf("found %d\n", vidCt)
		ret.AddFor("videos", vidCt)
		ret.AddDurFor("videos", time.Since(ret.Start))
		ioutil.WriteFile(cli.CacheDir+"/"+name+".json", bytes, 0755)
	case "update":
		id := "4a15e1fc-a422-466d-8cad-677c1605983c"
		video, _ := api.GetVideo(id)
		log.Printf("Got video %s", video.Id)
		video.CompletenessScore = 10.1
		err := video.Update()
		if err != nil {
			log.Printf("Got error %s", err.Error)
		} else {
			log.Printf("Got video score %.1f\n", video.CompletenessScore)
		}
	default:
		handleError(errors.New("unknown command '" + cli.Command + "'"))
	}
	log.Println(ret.String())
}

func main() {
	set, err := helper.LoadFromFile(cli.GetString("cred_file"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	handleV2(set.ApiV2)
}
