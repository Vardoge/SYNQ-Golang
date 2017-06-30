package synq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=b7230fea53d525948f33abf5f4b893f5")
		assert.Nil(err)
		assert.Equal("b7230fea53d525948f33abf5f4b893f5", token)
	}

	// non-uuid token
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.Nil(err)
		assert.Equal("not32characters", token)
	}

	// no url scheme
	{
		token, err := tokenOfUploaderURL("://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=b7230fea53d525948f33abf5f4b893f5")
		assert.NotNil(err)
		assert.Equal("", token)
	}

	// no query
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115")
		assert.NotNil(err)
		assert.Equal("", token)
	}

	// no "token" parameter
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?notoken=notoken")
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

				token, err := tokenOfUploaderURL(u.String())

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
