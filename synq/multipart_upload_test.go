package synq

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestTokenOfUploaderURL(t *testing.T) {
	assert := assert.New(t)

	// real example
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=b7230fea53d525948f33abf5f4b893f5")
		assert.Equal(nil, err)
		assert.Equal("b7230fea53d525948f33abf5f4b893f5", token)
	}

	// non-uuid token
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.Equal(nil, err)
		assert.Equal("not32characters", token)
	}

	// no url scheme
	{
		token, err := tokenOfUploaderURL("://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=b7230fea53d525948f33abf5f4b893f5")
		assert.NotEqual(nil, err)
		assert.Equal("", token)
	}

	// no query
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115")
		assert.NotEqual(nil, err)
		assert.Equal("", token)
	}

	// no "token" parameter
	{
		token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?notoken=notoken")
		assert.NotEqual(nil, err)
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

func TestBucketOfUploadAction(t *testing.T) {
	assert := assert.New(t)

	// real example
	{
		bucket, err := bucketOfUploadAction("https://synqfm.s3.amazonaws.com")
		assert.Equal(nil, err)
		assert.Equal("synqfm", bucket)
	}

	// another bucket
	{
		bucket, err := bucketOfUploadAction("https://another-bucket.s3.amazonaws.com")
		assert.Equal(nil, err)
		assert.Equal("another-bucket", bucket)
	}

	// yet another bucket with slash at the end
	{
		bucket, err := bucketOfUploadAction("https://yet-another-bucket.s3.amazonaws.com/")
		assert.Equal(nil, err)
		assert.Equal("yet-another-bucket", bucket)
	}

	// not https but http
	{
		bucket, err := bucketOfUploadAction("http://not-https.s3.amazonaws.com")
		assert.Equal(nil, err)
		assert.Equal("not-https", bucket)
	}

	// another kind of url
	{
		bucket, err := bucketOfUploadAction("https://uploader.synq.fm")
		assert.NotEqual(nil, err)
		assert.NotEqual("uploader", bucket)
	}

	// yet another kind of url
	{
		bucket, err := bucketOfUploadAction("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.NotEqual(nil, err)
		assert.NotEqual("uploader", bucket)
	}

	// no bucket, and the bucket that is not there is not named "s3"
	{
		bucket, err := bucketOfUploadAction("https://s3.amazonaws.com")
		assert.NotEqual(nil, err)
		assert.NotEqual("s3", bucket)
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
			gen.AnyString().SuchThat(
				func(v string) bool { return v != "" },
			),
		))

		p.TestingRun(t)
	}
}
