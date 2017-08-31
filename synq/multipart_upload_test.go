package synq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/stretchr/testify/assert"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestBucketOfUploadAction(t *testing.T) {
	assert := assert.New(t)

	// real example
	{
		bucket, err := bucketOfUploadAction("https://synqfm.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("synqfm", bucket)
	}

	// another bucket
	{
		bucket, err := bucketOfUploadAction("https://another-bucket.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("another-bucket", bucket)
	}

	// yet another bucket with slash at the end
	{
		bucket, err := bucketOfUploadAction("https://yet-another-bucket.s3.amazonaws.com/")
		assert.Nil(err)
		assert.Equal("yet-another-bucket", bucket)
	}

	// not https but http
	{
		bucket, err := bucketOfUploadAction("http://not-https.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("not-https", bucket)
	}

	// not the special "us-east-1" region
	{
		bucket, err := bucketOfUploadAction("https://a-bucket-in-another-region.s3-eu-west-1.amazonaws.com")
		assert.Nil(err)
		assert.Equal("a-bucket-in-another-region", bucket)
	}

	// invalid region
	{
		bucket, err := bucketOfUploadAction("https://invalid-region.not-s3-eu-west-1.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// another kind of url
	{
		bucket, err := bucketOfUploadAction("https://uploader.synq.fm")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// yet another kind of url
	{
		bucket, err := bucketOfUploadAction("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// no bucket, and the bucket that is not there is not named "s3"
	{
		bucket, err := bucketOfUploadAction("https://s3.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// invalid url
	{
		bucket, err := bucketOfUploadAction("https://bucket.s3.amazonaws.com/%")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// not "amazonaws"
	{
		bucket, err := bucketOfUploadAction("https://bucket.s3.amazon.com/")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// not "com"
	{
		bucket, err := bucketOfUploadAction("https://bucket.s3.amazonaws.org/")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// any non-empty bucket name
	{
		p := gopter.NewProperties(nil)

		p.Property("extract any bucket name", prop.ForAll(
			func(v string) bool {
				const format = "https://%s.s3.amazonaws.com"

				bucket, err := bucketOfUploadAction(fmt.Sprintf(format, v))

				return bucket == v && err == nil
			},
			gen.RegexMatch("^[a-z][a-z0-9_-]+$"),
		))

		p.TestingRun(t)
	}
}

func TestRegionOfUploadAction(t *testing.T) {
	assert := assert.New(t)

	// real example
	{
		region, err := regionOfUploadAction("https://synqfm.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("us-east-1", region)
	}

	// another region
	{
		region, err := regionOfUploadAction("https://a-bucket.s3-us-west-2.amazonaws.com")
		assert.Nil(err)
		assert.Equal("us-west-2", region)
	}

	// yet another region with slash at the end
	{
		region, err := regionOfUploadAction("https://some-bucket.s3-ap-south-1.amazonaws.com/")
		assert.Nil(err)
		assert.Equal("ap-south-1", region)
	}

	// not https but http
	{
		region, err := regionOfUploadAction("http://not-https.s3-sa-east-1.amazonaws.com")
		assert.Nil(err)
		assert.Equal("sa-east-1", region)
	}

	// invalid region
	{
		region, err := regionOfUploadAction("https://invalid-region.not-s3-eu-west-1.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// another kind of url
	{
		region, err := regionOfUploadAction("https://player.synq.fm")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// yet another kind of url
	{
		region, err := regionOfUploadAction("https://player.synq.fm/embed/55d4062f99454c9fb21e5186a09c2115")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// no "s3", and the region that is not there is not named "us-east-1"
	{
		region, err := regionOfUploadAction("https://bucket.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// invalid url
	{
		region, err := regionOfUploadAction("https://bucket.s3.amazonaws.com/%")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// not "amazonaws"
	{
		region, err := regionOfUploadAction("https://bucket.s3.amazon.com/")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// not "com"
	{
		region, err := regionOfUploadAction("https://bucket.s3.amazonaws.org/")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// start with "s3", but not with "s3-"
	{
		region, err := regionOfUploadAction("https://invalid-region.s3eu-west-1.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// any non-empty region name
	{
		p := gopter.NewProperties(nil)

		p.Property("extract any region name", prop.ForAll(
			func(v string) bool {
				const format = "https://bucket.s3-%s.amazonaws.com"

				region, err := regionOfUploadAction(fmt.Sprintf(format, v))

				return region == v && err == nil
			},
			gen.RegexMatch("^[a-z]+-[a-z]-[0-9]+$"),
		))

		p.TestingRun(t)
	}
}

func TestTokenOfUploaderURL(t *testing.T) {
	assert := assert.New(t)

	// real example
	{
		token, err := TokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=b7230fea53d525948f33abf5f4b893f5")
		assert.Nil(err)
		assert.Equal("b7230fea53d525948f33abf5f4b893f5", token)
	}

	// non-uuid token
	{
		token, err := TokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.Nil(err)
		assert.Equal("not32characters", token)
	}

	// no url scheme
	{
		token, err := TokenOfUploaderURL("://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=b7230fea53d525948f33abf5f4b893f5")
		assert.NotNil(err)
		assert.Equal("", token)
	}

	// no query
	{
		token, err := TokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115")
		assert.NotNil(err)
		assert.Equal("", token)
	}

	// no "token" parameter
	{
		token, err := TokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?notoken=notoken")
		assert.NotNil(err)
		assert.Equal("", token)
	}

	// any non-empty token string
	{
		p := gopter.NewProperties(nil)

		p.Property("extract any token string", prop.ForAll(
			func(v string) bool {
				const base = "https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115"

				u, err := url.Parse(base)
				if err != nil {
					panic(err)
				}

				values := url.Values{}
				values.Set("token", v)
				u.RawQuery = values.Encode()

				token, err := TokenOfUploaderURL(u.String())

				return token == v && err == nil
			},
			gen.AnyString().SuchThat(
				func(v string) bool { return v != "" },
			),
		))

		p.TestingRun(t)
	}
}

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

		// TODO(mastensg): check errors by some other method than string matching
		assert.Equal([]byte(nil), signature)
		assert.Equal(expectedError, err.Error())
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
