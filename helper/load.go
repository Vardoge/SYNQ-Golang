package helper

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

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

func LoadVideoV2(id string, c common.Cacheable, api synq.ApiV2) (video synq.VideoV2, err error) {
	ok := LoadFromCache(id, c, &video)
	if ok {
		video.Api = &api
		return video, nil
	}
	log.Printf("Getting video %s\n", id)
	video, err = api.GetVideo(id)
	if err != nil {
		return video, err
	}
	SaveToCache(id, c, &video)
	video.Api = &api
	return video, nil
}

func LoadUploadParameters(id string, req synq.UnicornParam, c common.Cacheable, api synq.ApiV2) (up synq.UploadParameters, err error) {
	ok := LoadFromCache(id+"_up", c, &up)
	if ok {
		return up, nil
	}
	log.Printf("Getting upload parameters for %s\n", id)
	up, err = api.GetUploadParams(id, req)
	if err != nil {
		return up, err
	}
	SaveToCache(id+"_up", c, &up)
	return up, nil
}

func LoadAsset(id string, c common.Cacheable, api synq.ApiV2) (asset synq.Asset, err error) {
	ok := LoadFromCache(id, c, &asset)
	if ok {
		asset.Api = api
		return asset, nil
	}
	log.Printf("Getting asset %s\n", id)
	asset, err = api.GetAsset(id)
	if err != nil {
		return asset, err
	}
	asset.Api = api
	video, e2 := LoadVideoV2(asset.VideoId, c, api)
	if e2 != nil {
		return asset, e2
	}
	log.Printf("Getting asset %s\n", id)
	asset, err = api.GetAsset(id)
	if err != nil {
		return asset, err
	}
	asset.Video = video
	SaveToCache(id, c, &asset)
	return asset, nil
}
