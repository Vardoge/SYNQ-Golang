package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/SYNQfm/SYNQ-Golang/helper"
	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/helpers/common"
)

var cli common.Cli

func init() {
	cli = common.NewCli()
	cli.DefaultSetup("for v2 'upload', get_video', for v1 : details, upload_info, upload, create, uploader_info, uploader, query or create_and_then_multipart_upload", "upload")
	cli.String("version", "v2", "version to use")
	cli.String("api_key", "", "pass the synq api key")
	cli.String("upload_url", synq.DEFAULT_UPLOADER_URL, "upload url to use")
	cli.String("user", "", "user to use")
	cli.String("password", "", "password to use")
	cli.String("video_id", "", "video id to access")
	cli.String("asset_id", "", "asset id to access")
	cli.String("file", "", "path to file you want to upload or userdata")
	cli.String("query", "", "query string to use")
	cli.Parse()
}

func handleError(err error) {
	if err != nil {
		log.Printf("Error : %s\n", err.Error())
		os.Exit(1)
	}
}

func handleV2(api synq.ApiV2) {
	var err error
	var video synq.VideoV2
	vid := cli.GetString("video_id")
	aid := cli.GetString("asset_id")
	switch cli.Command {
	case "upload":
		var asset synq.Asset
		upload_url := cli.GetString("upload_url")
		if upload_url == "" {
			err = errors.New("missing upload_url")
			handleError(err)
		}
		api.UploadUrl = upload_url
		file := cli.GetString("file")
		if file == "" {
			err = errors.New("file missing")
			handleError(err)
		}
		ext := filepath.Ext(file)
		ctype := mime.TypeByExtension(ext)
		if ctype == "" {
			handleError(errors.New("can not find cypte for " + ext))
		}
		if aid == "" {
			video, err = helper.LoadVideoV2(vid, cli, api)
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
					asset, err = video.CreateAssetForUpload(ctype)
				}
			}
		} else {
			log.Printf("getting existing asset %s\n", aid)
			asset, err = helper.LoadAsset(aid, cli, api)
		}
		handleError(err)
		params := synq.UnicornParam{
			Ctype:   ctype,
			AssetId: asset.Id,
		}
		up, e := helper.LoadUploadParameters(asset.VideoId, params, cli, api)
		handleError(e)
		log.Printf("Got upload params for %s", up.Key)
		asset.UploadParameters = up

		cli.Printf("uploading file %s\n", file)
		if !cli.Simulate {
			err = asset.UploadFile(file)
			handleError(err)
			log.Printf("uploaded file %s\n", file)
		}
	case "get_video":
		log.Printf("getting video %s\n", vid)
		video, err = api.GetVideo(vid)
		handleError(err)
		log.Printf(video.Display())
	default:
		err = errors.New("unknown command '" + cli.Command + "'")
	}
}

func handleV1(api synq.Api) {
	var video synq.Video
	var err error
	vid := cli.GetString("video_id")
	file := cli.GetString("file")
	switch cli.Command {
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
		q := cli.GetString("query")
		videos, err := api.Query(q)
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
		err = errors.New("unknown command '" + cli.Command + "'")
	}
	handleError(err)
	log.Printf(video.Display())
}

func main() {
	user := cli.GetString("user")
	password := cli.GetString("password")
	if user != "" && password != "" {
		api, err := synq.Login(user, password)
		handleError(err)
		handleV2(api)
	} else {
		api_key := cli.GetString("api_key")
		if api_key == "" {
			log.Println("missing api_key")
			os.Exit(1)
		}
		if cli.GetString("version") == "v2" {
			api := synq.NewV2(api_key)
			handleV2(api)
		} else {
			api := synq.NewV1(api_key)
			handleV1(api)
		}
	}
}
