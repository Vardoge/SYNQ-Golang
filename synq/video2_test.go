package synq

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := require.New(t)
	video := setupTestVideoV2()
	asset, err := video.CreateAsset(ASSET_CREATED, ASSET_TYPE, ASSET_LOCATION)
	assert.Nil(err)
	assert.NotNil(asset.Id)
	assert.Equal(ASSET_ID, asset.Id)
}

func TestCreateOrUpdateAsset(t *testing.T) {
	log.Println("Testing CreateAsset")
	assert := require.New(t)
	video := setupTestVideoV2()
	asset := Asset{State: ASSET_CREATED, Type: ASSET_TYPE, Location: ASSET_LOCATION}
	err := video.CreateOrUpdateAsset(&asset)
	assert.Nil(err)
	assert.Equal(ASSET_ID, asset.Id)
	asset.State = ASSET_UPLOADED
	err = video.CreateOrUpdateAsset(&asset)
	assert.Nil(err)
	assert.Len(testReqs, 2)
	req := testReqs[1]
	assert.Equal("PUT", req.Method)
}
