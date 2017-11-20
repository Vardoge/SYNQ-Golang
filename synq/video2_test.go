package synq

import (
	"log"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_helper"
	"github.com/stretchr/testify/require"
)

func TestCreateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := require.New(t)
	video := setupTestVideoV2()
	asset, err := video.CreateAsset(ASSET_CREATED, ASSET_TYPE, ASSET_LOCATION)
	assert.Nil(err)
	assert.NotNil(asset.Id)
	assert.Equal(testAssetId, asset.Id)
}

func TestCreateOrUpdateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := require.New(t)
	video := setupTestVideoV2()
	asset := Asset{State: ASSET_CREATED, Type: ASSET_TYPE, Location: ASSET_LOCATION}
	err := video.CreateOrUpdateAsset(&asset)
	assert.Nil(err)
	assert.Equal(testAssetId, asset.Id)
	asset.State = ASSET_UPLOADED
	err = video.CreateOrUpdateAsset(&asset)
	reqs, _ := test_helper.GetReqs()
	assert.Nil(err)
	assert.Len(reqs, 2)
	req := reqs[1]
	assert.Equal("PUT", req.Method)
}
