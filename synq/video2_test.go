package synq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	ASSET_ID = "1234"
)

func TestAddAsset(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(TEST_AUTH)
	v := VideoV2{Api: &api}
	a := Asset{}
	err := v.AddAsset(a)
	assert.Nil(err)
	assert.Len(v.Assets, 1)
	assert.Equal(ASSET_ID, v.Assets[0].Id)
}

func TestUpdateAsset(t *testing.T) {
	assert := require.New(t)
	api := setupTestApiV2(TEST_AUTH)
	v := VideoV2{Api: &api}
	a := Asset{Id: "1234", Video: v}
	err := a.Update()
	assert.Nil(err)
	assert.Equal(V2_VIDEO_ID, a.VideoId)
}
