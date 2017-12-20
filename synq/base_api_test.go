package synq

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/assert"
)

type BadReader struct {
}

func loadSample(file string) []byte {
	return test_server.LoadSampleDir(file, DEFAULT_SAMPLE_DIR)
}

func (b BadReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed to read")
}

func TestParseAwsResp(t *testing.T) {
	assert := assert.New(t)
	var v interface{}
	resp := http.Response{
		StatusCode: 204,
	}
	err := errors.New("failure")
	e := parseAwsResp(&resp, err, v)
	assert.NotNil(e)
	assert.Equal("failure", e.Error())

	br := BadReader{}
	resp = http.Response{
		StatusCode: 412,
		Body:       ioutil.NopCloser(br),
	}
	e = parseAwsResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("failed to read", e.Error())

	err_msg := loadSample("upload")
	resp = http.Response{
		StatusCode: 412,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseAwsResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("EOF", e.Error())

	err_msg = loadSample("aws_err.xml")
	resp = http.Response{
		StatusCode: 412,
		Body:       ioutil.NopCloser(bytes.NewBuffer(err_msg)),
	}
	e = parseAwsResp(&resp, nil, v)
	assert.NotNil(e)
	assert.Equal("At least one of the pre-conditions you specified did not hold", e.Error())

	resp = http.Response{
		StatusCode: 204,
	}
	e = parseAwsResp(&resp, nil, v)
	assert.Nil(e)
}

func TestNew(t *testing.T) {
	assert := assert.New(t)
	api := New("key")
	assert.NotNil(api)
	assert.Equal("key", api.GetKey())
	assert.Equal(time.Duration(DEFAULT_TIMEOUT_MS)*time.Millisecond, api.GetTimeout(""))
	assert.Equal(time.Duration(DEFAULT_UPLOAD_MS)*time.Millisecond, api.GetTimeout("upload"))
	api = New("key", time.Duration(15)*time.Second)
	assert.Equal("key", api.GetKey())
	assert.Equal(time.Duration(15)*time.Second, api.GetTimeout(""))
	api = New("key", time.Duration(30)*time.Second, time.Duration(100)*time.Second)
	assert.Equal("key", api.GetKey())
	assert.Equal(time.Duration(30)*time.Second, api.GetTimeout(""))
	assert.Equal(time.Duration(100)*time.Second, api.GetTimeout("upload"))
}
