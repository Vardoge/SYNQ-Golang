package upload

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/stretchr/testify/require"
)

func TestCreateV4Request(t *testing.T) {
	assert := require.New(t)
	header := "test-header"
	signed := http.Header{}
	reqHeaders := http.Header{}
	signed.Add(header, "val")
	reqHeaders.Add(header, "val2")
	req := &request.Request{SignedHeaderVals: signed}
	u := url.URL{Path: "url", RawQuery: "a=b&c=d"}
	hr := http.Request{URL: &u, Header: reqHeaders, Method: "PUT"}
	req.HTTPRequest = &hr
	params := UploadParameters{}
	v4 := CreateV4Request(params, req)
	assert.NotNil(v4)
	assert.Equal("us-east-1", v4.Region)
	assert.Equal("PUT", v4.Method)
	assert.Equal("url", v4.Path)
	assert.Equal("a=b&c=d", v4.RawQuery)
	assert.Equal("", v4.Headers[header])
	assert.Equal("val2", v4.Headers["Test-Header"])
	built := v4.BuildRequest()
	assert.Equal(v4.Method, built.Method)
	assert.Equal(v4.RawQuery, built.URL.RawQuery)
	assert.Equal(v4.Path, built.URL.Path)
	assert.Equal("val2", built.Header.Get("Test-Header"))
}
