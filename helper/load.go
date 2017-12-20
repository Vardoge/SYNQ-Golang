package helper

import (
	"log"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/SYNQ-Golang/upload"
	"github.com/SYNQfm/helpers/common"
)

func LoadVideosByQuery(query, name string, c common.Cacheable, api synq.Api) (videos []synq.Video, err error) {
	ok := common.LoadFromCache(name, c, &videos)
	if ok {
		return videos, nil
	}
	log.Printf("querying '%s'\n", query)
	videos, err = api.Query(query)
	if err != nil {
		return videos, err
	}
	common.SaveToCache(name, c, videos)
	return videos, err
}

// fow now, the query will be the account id
func LoadVideosByQueryV2(query, name string, c common.Cacheable, api synq.ApiV2) (videos []synq.VideoV2, err error) {
	ok := common.LoadFromCache(name, c, &videos)
	if ok {
		return videos, nil
	}
	log.Printf("get all videos (filter by account '%s')\n", query)
	videos, err = api.GetVideos(query)
	if err != nil {
		return videos, err
	}
	common.SaveToCache(name, c, videos)
	return videos, err
}

func LoadVideo(id string, c common.Cacheable, api synq.Api) (video synq.Video, err error) {
	ok := common.LoadFromCache(id, c, &video)
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
	common.SaveToCache(id, c, &video)
	return video, nil
}

func LoadVideoV2(id string, c common.Cacheable, api synq.ApiV2) (video synq.VideoV2, err error) {
	ok := common.LoadFromCache(id, c, &video)
	if ok {
		video.Api = &api
		assets := []synq.Asset{}
		for _, a := range video.Assets {
			a.Video = video
			a.Api = api
			assets = append(assets, a)
		}
		video.Assets = assets
		return video, nil
	}
	log.Printf("Getting video %s\n", id)
	video, err = api.GetVideo(id)
	if err != nil {
		return video, err
	}
	common.SaveToCache(id, c, &video)
	video.Api = &api
	return video, nil
}

func LoadUploadParameters(id string, req synq.UnicornParam, c common.Cacheable, api synq.ApiV2) (up upload.UploadParameters, err error) {
	lookId := id
	if req.AssetId != "" {
		lookId = req.AssetId
	}
	ok := common.LoadFromCache(lookId+"_up", c, &up)
	if ok {
		return up, nil
	}
	log.Printf("Getting upload parameters for %s\n", id)
	up, err = api.GetUploadParams(id, req)
	if err != nil {
		return up, err
	}
	common.SaveToCache(lookId+"_up", c, &up)
	return up, nil
}

func LoadAsset(id string, c common.Cacheable, api synq.ApiV2) (asset synq.Asset, err error) {
	ok := common.LoadFromCache(id, c, &asset)
	if !ok {
		log.Printf("Getting asset %s\n", id)
		asset, err = api.GetAsset(id)
		if err != nil {
			return asset, err
		}
	} else {
		asset.Api = api
	}

	if asset.Video.Id == "" {
		video, err := LoadVideoV2(asset.VideoId, c, api)
		if err != nil {
			return asset, err
		}
		asset.Video = video
	} else {
		// cache the video for re-use later
		common.SaveToCache(asset.Video.Id, c, &asset.Video)
	}
	common.SaveToCache(id, c, &asset)
	return asset, nil
}
