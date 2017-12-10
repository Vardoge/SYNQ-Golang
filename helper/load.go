package helper

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/helpers/common"
)

func LoadFromCache(name string, c common.Cacheable, obj interface{}) bool {
	cacheFile := c.GetCacheFile(name)
	if cacheFile != "" {
		if _, e := os.Stat(cacheFile); e == nil {
			log.Printf("loading from cached file %s\n", cacheFile)
			bytes, _ := ioutil.ReadFile(cacheFile)
			json.Unmarshal(bytes, obj)
			return true
		}
	}
	return false
}

func SaveToCache(name string, c common.Cacheable, obj interface{}) bool {
	cacheFile := c.GetCacheFile(name)
	if cacheFile != "" {
		data, _ := json.Marshal(obj)
		ioutil.WriteFile(cacheFile, data, 0755)
		return true
	}
	return false
}

func LoadVideosByQuery(query, name string, c common.Cacheable, api synq.Api) (videos []synq.Video, err error) {
	ok := LoadFromCache(name, c, &videos)
	if ok {
		return videos, nil
	}
	log.Printf("querying '%s'\n", query)
	videos, err = api.Query(query)
	if err != nil {
		return videos, err
	}
	SaveToCache(name, c, videos)
	return videos, err
}

// fow now, the query will be the account id
func LoadVideosByQueryV2(query, name string, c common.Cacheable, api synq.ApiV2) (videos []synq.VideoV2, err error) {
	ok := LoadFromCache(name, c, &videos)
	if ok {
		return videos, nil
	}
	log.Printf("get all videos (filter by account '%s')\n", query)
	videos, err = api.GetVideos(query)
	if err != nil {
		return videos, err
	}
	SaveToCache(name, c, videos)
	return videos, err
}

func LoadVideo(id string, c common.Cacheable, api synq.Api) (video synq.Video, err error) {
	ok := LoadFromCache(id, c, &video)
	if ok {
		video.Api = &api
		return video, nil
	}
	// need to use the v1 api to get the raw video data
	log.Printf("Getting video %s", id)
	video, e := api.GetVideo(id)
	if e != nil {
		return video, e
	}
	SaveToCache(id, c, &video)
	return video, nil
}

func LoadObjectV2(id string, obj interface{}, c common.Cacheable, api synq.ApiV2) error {
	ok := LoadFromCache(id, c, obj)
	if ok {
		return nil
	}
	// need to use the v1 api to get the raw video data
	if _, ok := obj.(*synq.VideoV2); ok {
		log.Printf("Getting video %s\n", id)
		video, err := api.GetVideo(id)
		if err != nil {
			return err
		}
		obj = &video
	} else if _, ok := obj.(*synq.Asset); ok {
		log.Printf("Getting asset %s\n", id)
		asset, err := api.GetAsset(id)
		if err != nil {
			return err
		}
		obj = &asset
	} else if up, ok := obj.(*UpObj); ok {
		vid := strings.Split(id, "_")[0]
		log.Printf("Getting UploadParameters for video id %s, %+v\n", vid, up)
		params, err := api.GetUploadParams(vid, up.ReqParams)
		if err != nil {
			return err
		}
		up.UploadParams = params
		obj = &up
	} else {
		return errors.New("obj type is unknown")
	}
	SaveToCache(id, c, obj)
	return nil
}

func LoadVideoV2(id string, c common.Cacheable, api synq.ApiV2) (video synq.VideoV2, err error) {
	err = LoadObjectV2(id, &video, c, api)
	if err != nil {
		return video, err
	}
	video.Api = &api
	return video, nil
}

type UpObj struct {
	UploadParams synq.UploadParameters `json:"upload_params"`
	ReqParams    synq.UnicornParam     `json:"request_params"`
}

func LoadUploadParameters(id string, req synq.UnicornParam, c common.Cacheable, api synq.ApiV2) (up synq.UploadParameters, err error) {
	obj := UpObj{
		ReqParams: req,
	}
	err = LoadObjectV2(id+"_up", &obj, c, api)
	if err != nil {
		return up, err
	}
	return obj.UploadParams, err
}

func LoadAsset(id string, c common.Cacheable, api synq.ApiV2) (asset synq.Asset, err error) {
	err = LoadObjectV2(id, &asset, c, api)
	if err != nil {
		return asset, err
	}
	asset.Api = api
	video, e2 := LoadVideoV2(asset.VideoId, c, api)
	if e2 != nil {
		return asset, e2
	}
	asset.Video = video
	return asset, nil
}
