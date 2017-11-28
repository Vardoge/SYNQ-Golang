package synq

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/require"
)

func TestMarshalVideo(t *testing.T) {
	assert := require.New(t)
	video := setupTestVideoV2()
	video.Metadata = []byte(`{"test":true}`)
	bytes, err := json.Marshal(video)
	assert.Nil(err)
	v := VideoV2{}
	json.Unmarshal(bytes, &v)
	assert.Equal(video.Metadata, v.Metadata)
}

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
	reqs, vals := test_server.GetReqs()
	assert.Nil(err)
	assert.Len(reqs, 2)
	assert.Len(vals, 2)
	req := reqs[1]
	val := vals[1]
	assert.Equal("PUT", req.Method)
	a := Asset{}
	body := val.Get("body")
	json.Unmarshal([]byte(body), &a)
	assert.Equal(asset.State, a.State)
	assert.Equal(asset.Location, a.Location)
}
