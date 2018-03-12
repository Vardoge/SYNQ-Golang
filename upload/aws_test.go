package upload

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/stretchr/testify/require"
)

func setupServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(handle))
	return server
}

// we can't use test_server as it will create a circular loop
func handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/sig":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Authorization":"sig123", "Date":"20180223T002913Z"}`))
	}
}

func createTestAwsReq(header ...string) *request.Request {
	signed := http.Header{}
	reqHeaders := http.Header{}
	if len(header) == 0 {
		header = append(header, "test-header")
	}
	for idx, key := range header {
		signed.Add(key, fmt.Sprintf("signed_val_%d", idx))
		reqHeaders.Add(key, fmt.Sprintf("req_val_%d", idx))
	}
	req := &request.Request{SignedHeaderVals: signed}
	u := url.URL{Path: "url", RawQuery: "a=b&c=d"}
	hr := http.Request{URL: &u, Header: reqHeaders, Method: "PUT"}
	req.HTTPRequest = &hr
	return req
}

func TestCreateV4Request(t *testing.T) {
	assert := require.New(t)
	header := "test-header"
	req := createTestAwsReq(header)
	params := UploadParameters{}
	v4 := CreateV4Request(params, req)
	assert.NotNil(v4)
	assert.Equal("us-east-1", v4.Region)
	assert.Equal("PUT", v4.Method)
	assert.Equal("url", v4.Path)
	assert.Equal("a=b&c=d", v4.RawQuery)
	assert.Equal("", v4.Headers[header])
	assert.Equal("req_val_0", v4.Headers["Test-Header"])
	built := v4.BuildRequest()
	assert.Equal(v4.Method, built.Method)
	assert.Equal(v4.RawQuery, built.URL.RawQuery)
	assert.Equal(v4.Path, built.URL.Path)
	assert.Equal("req_val_0", built.Header.Get("Test-Header"))
}

func TestSign(t *testing.T) {
	assert := require.New(t)
	headers := make(map[string]string)
	headers["test-header"] = "val"
	req := V4Request{Region: "us-east-1", Headers: headers}
	resp, err := req.Sign("a", "b")
	assert.Nil(err)
	assert.NotEmpty(resp.Date)
	assert.Contains(resp.Authorization, "AWS4-HMAC-SHA256 Credential=a/20180312/us-east-1/s3/aws4_request, SignedHeaders=host;test-header;x-amz-content-sha256;x-amz-date")
}

func TestNewAwsUpload(t *testing.T) {
	assert := require.New(t)
	params := UploadParameters{
		Key:            "abc",
		Acl:            "private",
		ContentType:    "video/mp4",
		AwsAccessKeyId: "key",
		SignatureUrl:   "sig",
	}
	_, err := NewAwsUpload(params)
	assert.NotNil(err)
	assert.Equal("Invalid action URL. Not exactly 4 period-separated words in host.", err.Error())
	params.Action = "https://synqfm.s3.amazonaws.com"
	u, err := NewAwsUpload(params)
	assert.Nil(err)
	au := u.(*AwsUpload)
	assert.Equal(params.Action, au.Url())
	region, e := au.GetRegion()
	assert.Nil(e)
	assert.Equal("us-east-1", region)
	bucket, e := au.GetBucket()
	assert.Nil(e)
	assert.Equal("synqfm", bucket)
	assert.Equal(params.Key, au.Key())
	assert.Equal(params.Acl, au.Acl())
	assert.Equal(params.ContentType, au.ContentType())
	assert.Equal(params.AwsAccessKeyId, au.AwsKeyId())
	assert.Equal(params.SignatureUrl, au.UploaderSigUrl())
}

func TestServerSign(t *testing.T) {
	assert := require.New(t)
	params := UploadParameters{
		Key:            "abc",
		Acl:            "private",
		ContentType:    "video/mp4",
		AwsAccessKeyId: "key",
		SignatureUrl:   "sig",
		Action:         "https://synqfm.s3.amazonaws.com",
	}
	u, err := NewAwsUpload(params)
	assert.Nil(err)
	au := u.(*AwsUpload)
	r := createTestAwsReq()
	_, err = au.ServerSignV2(r)
	assert.NotNil(err)
	assert.Equal("Post sig: unsupported protocol scheme \"\"", err.Error())
	server := setupServer()
	params.SignatureUrl = server.URL + "/sig"
	u, _ = NewAwsUpload(params)
	au = u.(*AwsUpload)
	sig, err := au.ServerSignV2(r)
	assert.Nil(err)
	assert.Equal("sig123", sig)
}
