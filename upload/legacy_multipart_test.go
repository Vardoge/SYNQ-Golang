package upload

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/stretchr/testify/assert"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func uploaderSignatureUrlFormatOfTestServerUrl(u string) string {
	const f = "%s/uploader/signature/%%s?token=%%s"
	return fmt.Sprintf(f, u)
}

func TestUploaderSignatureUrlFormatOfTestServerUrl(t *testing.T) {
	assert := assert.New(t)

	const u = "http://127.0.0.1:34377"

	f := uploaderSignatureUrlFormatOfTestServerUrl(u)

	assert.Equal("http://127.0.0.1:34377/uploader/signature/%s?token=%s", f)
}

func TestUploaderSignature(t *testing.T) {
	assert := assert.New(t)

	// no server
	{
		uf := uploaderSignatureUrlFormatOfTestServerUrl("http://0.0.0.0:0")

		const video_id = "e3c71a23462f07fea2ef317dcd3b7a9b"
		const token = "568575f9c000b533292adc88f5a2321a"
		const headers = `POST

video/mp4

x-amz-acl:public-read
x-amz-date:Fri, 30 Jun 2017 14:03:55 UTC
/synqfm/projects/00/00/00000000000000000000000000000000/uploads/videos/e3/c7/e3c71a23462f07fea2ef317dcd3b7a9b.mp4?uploads`

		signature, err := UploaderSignature(uf, video_id, token, headers)

		const expectedError = `Post http://0.0.0.0:0/uploader/signature/` +
			`e3c71a23462f07fea2ef317dcd3b7a9b?token=568575f9c000b533292adc88f5a2321a:` +
			` dial tcp 0.0.0.0:0: getsockopt: connection refused`
		const expectedMacError = `Post http://0.0.0.0:0/uploader/signature/` +
			`e3c71a23462f07fea2ef317dcd3b7a9b?token=568575f9c000b533292adc88f5a2321a:` +
			` dial tcp 0.0.0.0:0: connect: can't assign requested address`

		// TODO(mastensg): check errors by some other method than string matching
		assert.Equal([]byte(nil), signature)
		if runtime.GOOS == "darwin" {
			assert.Equal(expectedMacError, err.Error())
		} else {
			assert.Equal(expectedError, err.Error())
		}
	}

	// always internal server error
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "no", http.StatusInternalServerError)
		}))
		defer ts.Close()

		uf := uploaderSignatureUrlFormatOfTestServerUrl(ts.URL)

		const video_id = "e3c71a23462f07fea2ef317dcd3b7a9b"
		const token = "568575f9c000b533292adc88f5a2321a"
		const headers = `POST

video/mp4

x-amz-acl:public-read
x-amz-date:Fri, 30 Jun 2017 14:03:55 UTC
/synqfm/projects/00/00/00000000000000000000000000000000/uploads/videos/e3/c7/e3c71a23462f07fea2ef317dcd3b7a9b.mp4?uploads`

		signature, err := UploaderSignature(uf, video_id, token, headers)

		// TODO(mastensg): check errors by some other method than string matching
		assert.Equal([]byte(nil), signature)
		assert.Equal("HTTP response status not OK.", err.Error())
	}

	// return something which is not json
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "not json")
		}))
		defer ts.Close()

		uf := uploaderSignatureUrlFormatOfTestServerUrl(ts.URL)

		const video_id = "e3c71a23462f07fea2ef317dcd3b7a9b"
		const token = "568575f9c000b533292adc88f5a2321a"
		const headers = `POST

video/mp4

x-amz-acl:public-read
x-amz-date:Fri, 30 Jun 2017 14:03:55 UTC
/synqfm/projects/00/00/00000000000000000000000000000000/uploads/videos/e3/c7/e3c71a23462f07fea2ef317dcd3b7a9b.mp4?uploads`

		signature, err := UploaderSignature(uf, video_id, token, headers)

		// TODO(mastensg): check errors by some other method than string matching
		assert.Equal([]byte(nil), signature)
		assert.Equal("invalid character 'o' in literal null (expecting 'u')", err.Error())
	}

	// empty response
	{
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "")
		}))
		defer ts.Close()

		uf := uploaderSignatureUrlFormatOfTestServerUrl(ts.URL)

		const video_id = "e3c71a23462f07fea2ef317dcd3b7a9b"
		const token = "568575f9c000b533292adc88f5a2321a"
		const headers = `POST

video/mp4

x-amz-acl:public-read
x-amz-date:Fri, 30 Jun 2017 14:03:55 UTC
/synqfm/projects/00/00/00000000000000000000000000000000/uploads/videos/e3/c7/e3c71a23462f07fea2ef317dcd3b7a9b.mp4?uploads`

		signature, err := UploaderSignature(uf, video_id, token, headers)

		// TODO(mastensg): check errors by some other method than string matching
		assert.Equal([]byte(nil), signature)
		assert.Equal("unexpected end of JSON input", err.Error())
	}

	// return signature
	{
		const video_id = "e3c71a23462f07fea2ef317dcd3b7a9b"
		const token = "568575f9c000b533292adc88f5a2321a"
		const headers = `POST

video/mp4

x-amz-acl:public-read
x-amz-date:Fri, 30 Jun 2017 14:50:33 UTC
/synqfm/projects/20/fc/20fc57c626dc489ea285493b3813a0b5/uploads/videos/58/70/58705c8fcb054fb68fe85b61ab4f17af.mp4?uploads`
		const request = `{"headers":"POST\n\nvideo/mp4\n\nx-amz-acl:public-read\nx-amz-date:Fri,` +
			` 30 Jun 2017 14:50:33 UTC\n/synqfm/projects/20/fc/20fc57c626dc489ea285493b3813a0b5` +
			`/uploads/videos/58/70/58705c8fcb054fb68fe85b61ab4f17af.mp4?uploads"}`
		const expectedSignature = "/0OolBcoDZ95IbeDPMt5P+3kCnc="

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := fmt.Sprintf(`{"signature":"%s"}`, expectedSignature)
			fmt.Fprint(w, response)
		}))
		defer ts.Close()

		uf := uploaderSignatureUrlFormatOfTestServerUrl(ts.URL)

		signature, err := UploaderSignature(uf, video_id, token, headers)

		assert.Equal([]byte(expectedSignature), signature)
		assert.Nil(err)
	}
}

func TestReformatXAmzDate(t *testing.T) {
	assert := assert.New(t)

	// Example:
	{
		const in = "20060102T150405Z"
		out, err := ReformatXAmzDate(in)
		assert.Nil(err)
		assert.Equal("Mon, 02 Jan 2006 15:04:05 UTC", out)
	}

	// badly formatted
	{
		const in = "20060102T150405"
		out, err := ReformatXAmzDate(in)
		assert.NotNil(err)
		assert.Equal("", out)
	}

	// zero-padded date
	{
		const in = "20170707T141740Z"
		out, err := ReformatXAmzDate(in)
		assert.Nil(err)
		assert.Equal("Fri, 07 Jul 2017 14:17:40 UTC", out)
	}

	// any time
	{
		p := gopter.NewProperties(nil)

		p.Property("reformat any time", prop.ForAll(
			func(v int64) bool {
				const format_in = "20060102T150405Z"
				const format_expect_out = "Mon, 02 Jan 2006 15:04:05 MST"

				u := time.Unix(v, 0).UTC()

				in := u.Format(format_in)
				expect_out := u.Format(format_expect_out)

				out, err := ReformatXAmzDate(in)

				return out == expect_out && err == nil
			},
			gen.Int64Range(-(1<<35), (1<<37)),
		))

		p.TestingRun(t)
	}
}

func TestRewriteXAmzDate(t *testing.T) {
	assert := assert.New(t)

	// Example:
	{
		h := http.Header{}
		h.Set("X-Amz-Date", "20060102T150405Z")
		err := RewriteXAmzDateHeader(h)
		assert.Nil(err)
		assert.Equal("Mon, 02 Jan 2006 15:04:05 UTC", h.Get("X-Amz-Date"))
	}

	// badly formatted
	{
		h := http.Header{}
		h.Set("X-Amz-Date", "20060102T150405")
		err := RewriteXAmzDateHeader(h)
		assert.NotNil(err)
		assert.Equal("20060102T150405", h.Get("X-Amz-Date"))
	}

	// zero-padded date
	{
		h := http.Header{}
		h.Set("X-Amz-Date", "20170707T141740Z")
		err := RewriteXAmzDateHeader(h)
		assert.Nil(err)
		assert.Equal("Fri, 07 Jul 2017 14:17:40 UTC", h.Get("X-Amz-Date"))
	}

	// missing header
	{
		h := http.Header{}
		err := RewriteXAmzDateHeader(h)
		assert.NotNil(err)
		assert.Equal("", h.Get("X-Amz-Date"))
	}

	// any time
	{
		p := gopter.NewProperties(nil)

		p.Property("reformat any time", prop.ForAll(
			func(v int64) bool {
				const format_in = "20060102T150405Z"
				const format_expect_out = "Mon, 02 Jan 2006 15:04:05 MST"

				u := time.Unix(v, 0).UTC()

				in := u.Format(format_in)
				expect_out := u.Format(format_expect_out)

				h := http.Header{}
				h.Set("X-Amz-Date", in)

				err := RewriteXAmzDateHeader(h)
				out := h.Get("X-Amz-Date")

				return out == expect_out && err == nil
			},
			gen.Int64Range(-(1<<35), (1<<37)),
		))

		p.TestingRun(t)
	}
}

func TestMultipartUploadSignRequest(t *testing.T) {
	assert := assert.New(t)

	const (
		acl            = "public-read"
		awsAccessKeyId = "AAAAAAAAAAAAAAAAAAAA"
		bucket         = "not-synqfm"
		contentType    = "video/mp4"
		key            = "foo.mp4"
		token          = "b7230fea53d525948f33abf5f4b893f5"
		video_id       = "55d4062f99454c9fb21e5186a09c2115"
	)

	const signature = "/0OolBcoDZ95IbeDPMt5P+3kCnc="

	// server that always returns a signature
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf(`{"signature":"%s"}`, signature)
		fmt.Fprint(w, response)
	}))
	defer ts.Close()
	uf := uploaderSignatureUrlFormatOfTestServerUrl(ts.URL)

	// Initiate multi-part upload
	{
		r, err := http.NewRequest("POST", "https://foo.com/bar?uploads=", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		err = multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf, r)
		assert.Nil(err)
	}

	// Upload one part
	{
		r, err := http.NewRequest("PUT", "https://foo.com/bar", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		err = multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf, r)
		assert.Nil(err)
	}

	// Finish multi-part upload
	{
		r, err := http.NewRequest("POST", "https://foo.com/bar", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		err = multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf, r)
		assert.Nil(err)
	}

	// rewrite date
	{
		r, err := http.NewRequest("GET", "", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		err = multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf, r)
		assert.Equal("Mon, 02 Jan 2006 15:04:05 UTC", r.Header.Get("X-Amz-Date"))
	}

	// missing header
	{

		r, err := http.NewRequest("POST", "", strings.NewReader(""))
		assert.Nil(err)

		err = multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf, r)
		assert.Equal("Missing header: X-Amz-Date.", err.Error())
	}

	// Unknown request type
	{

		r, err := http.NewRequest("GET", "", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		err = multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf, r)
		assert.Equal("Unknown request type.", err.Error())
	}
}

func TestMultipartUploadSigner(t *testing.T) {
	assert := assert.New(t)

	const (
		acl            = "public-read"
		awsAccessKeyId = "AAAAAAAAAAAAAAAAAAAA"
		bucket         = "not-synqfm"
		contentType    = "video/mp4"
		key            = "foo.mp4"
		token          = "b7230fea53d525948f33abf5f4b893f5"
		video_id       = "55d4062f99454c9fb21e5186a09c2115"
	)

	const signature = "/0OolBcoDZ95IbeDPMt5P+3kCnc="

	// server that always returns a signature
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf(`{"signature":"%s"}`, signature)
		fmt.Fprint(w, response)
	}))
	defer ts.Close()
	uf := uploaderSignatureUrlFormatOfTestServerUrl(ts.URL)

	signer := MultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uf)

	// rewrite date
	{
		r, err := http.NewRequest("GET", "", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		ar := request.Request{}
		ar.HTTPRequest = r
		signer(&ar)
		assert.Equal("Mon, 02 Jan 2006 15:04:05 UTC", ar.HTTPRequest.Header.Get("X-Amz-Date"))
	}

	// missing header
	{

		r, err := http.NewRequest("POST", "", strings.NewReader(""))
		assert.Nil(err)

		ar := request.Request{}
		ar.HTTPRequest = r
		signer(&ar)
		// TODO(mastensg): how to report errors from handlers?
	}

	// Unknown request type
	{

		r, err := http.NewRequest("GET", "", strings.NewReader(""))
		assert.Nil(err)
		r.Header.Set("X-Amz-Date", "20060102T150405Z")

		ar := request.Request{}
		ar.HTTPRequest = r
		signer(&ar)
		// TODO(mastensg): how to report errors from handlers?
	}
}
