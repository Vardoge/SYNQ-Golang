package synq

import (
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ASSET_ID       = "01823629-bcf2-4c34-b714-ae21e1a4647f"
	ASSET_TYPE     = "mp4"
	ASSET_CREATED  = "created"
	ASSET_UPLOADED = "uploaded"
	ASSET_LOCATION = "https://s3.amazonaws.com/synq-jessica/uploads/01/82/01823629bcf24c34b714ae21e1a4647f/01823629bcf24c34b714ae21e1a4647f.mp4"
)

func setupTestVideoV2() VideoV2 {
	api := setupTestApiV2(TEST_AUTH)
	video, _ := api.Create()
	setupTestServer("v2")
	video.Api.Url = testServer.URL
	return video
}

func handleAsset(w http.ResponseWriter, r *http.Request) {
}

func TestGetAssetList(t *testing.T) {
	log.Println("Testing GetAssetList")
	assert := require.New(t)
	api := setupTestApiV2(TEST_AUTH)
	assets, err := api.GetAssetList()
	assert.Nil(err)
	assert.NotEmpty(assets)
	assert.Equal(V2_VIDEO_ID, assets[0].VideoId)
	assert.Equal(ASSET_ID, assets[0].Id)
	assert.Equal(ASSET_TYPE, assets[0].Type)
	assert.Equal(ASSET_LOCATION, assets[0].Location)
	assert.Equal(ASSET_CREATED, assets[0].State)
}

func TestGetVideoAssetList(t *testing.T) {
	log.Println("Testing GetVideoAssetList")
	assert := require.New(t)
	video := setupTestVideoV2()
	err := video.GetVideoAssetList()
	assert.Nil(err)
	assert.NotEmpty(video.Assets)
	assert.Equal(V2_VIDEO_ID, video.Assets[0].VideoId)
	assert.Equal(ASSET_ID, video.Assets[0].Id)
	assert.Equal(ASSET_TYPE, video.Assets[0].Type)
	assert.Equal(ASSET_LOCATION, video.Assets[0].Location)
	assert.Equal(ASSET_CREATED, video.Assets[0].State)
}

func TestGetAsset(t *testing.T) {
	log.Println("Testing GetAsset")
	assert := assert.New(t)
	video := setupTestVideoV2()
	asset, err := video.GetAsset(ASSET_ID)
	assert.Equal(V2_VIDEO_ID, asset.VideoId)
	assert.Equal(ASSET_TYPE, asset.Type)
	assert.Equal(ASSET_UPLOADED, asset.State)
	assert.Equal(ASSET_LOCATION, asset.Location)
	assert.Nil(err)
}

func TestCreateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := assert.New(t)
	video := setupTestVideoV2()
	asset, err := video.CreateAsset(ASSET_CREATED, ASSET_TYPE, ASSET_LOCATION)
	assert.Nil(err)
	assert.NotNil(asset.Id)
	assert.Equal(ASSET_ID, asset.Id)
}

func TestUpdate(t *testing.T) {
	log.Println("Testing Update")
	assert := assert.New(t)
	video := setupTestVideoV2()
	asset, _ := video.GetAsset(ASSET_ID)
	assert.NotEmpty(asset)
	asset.State = ASSET_UPLOADED
	err := asset.Update()
	assert.Nil(err)
	asset, _ = video.GetAsset(ASSET_ID)
	assert.Equal(ASSET_UPLOADED, asset.State)
}

func TestDelete(t *testing.T) {
	log.Println("Testing Delete")
	assert := assert.New(t)
	video := setupTestVideoV2()
	asset, _ := video.GetAsset(ASSET_ID)
	assert.NotEmpty(asset)
	err := asset.Delete()
	assert.Nil(err)
}

func TestHandleAssetReq(t *testing.T) {
	log.Println("Testing TestHandleAssetReq")
	assert := assert.New(t)
	video := setupTestVideoV2()
	asset, _ := video.GetAsset(ASSET_ID)
	ogAsset := asset
	url := video.Api.getBaseUrl() + "/assets/" + ASSET_ID
	err := asset.handleAssetReq("GET", url, nil)
	assert.Nil(err)
	assert.Equal(ogAsset, asset)
}
