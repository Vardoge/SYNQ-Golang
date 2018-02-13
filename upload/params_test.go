package upload

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCtype(t *testing.T) {
	log.Println("Testing parseCtype")
	assert := assert.New(t)
	ctype, _ := parseCtype("image/jpg")
	assert.Equal("image/jpeg", ctype)
	ctype, _ = parseCtype("video/msvideo")
	assert.Equal("video/avi", ctype)
	ctype, _ = parseCtype("text/xml")
	assert.Equal("text/xml", ctype)
	_, err := parseCtype(".mp4")
	assert.NotNil(err)
	assert.Equal("invalid ctype '.mp4'", err.Error())
}

func TestNewReq(t *testing.T) {
	assert := assert.New(t)
	bytes := []byte(`<xml>`)
	_, err := NewUploadRequest(bytes)
	assert.NotNil(err)
	bytes = []byte(`{"content_type": ".mp4"}`)
	_, err = NewUploadRequest(bytes)
	assert.NotNil(err)
	bytes = []byte(`{}`)
	req, err := NewUploadRequest(bytes)
	assert.Nil(err)
	assert.Equal(DefaultCtype, req.GetCType())
	assert.Equal(DefaultCtype, req.GetType())
	assert.Equal(DefaultAcl, req.GetAcl())
	req.Type = "source"
	assert.Equal("source", req.GetType())
}
