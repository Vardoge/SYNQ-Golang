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

	// not the special "us-east-1" region
	{
		bucket, err := bucketOfUploadAction("https://a-bucket-in-another-region.s3-eu-west-1.amazonaws.com")
		assert.Equal(nil, err)
		assert.Equal("a-bucket-in-another-region", bucket)
	}

	// invalid region
	{
		bucket, err := bucketOfUploadAction("https://invalid-region.not-s3-eu-west-1.amazonaws.com")
		assert.NotEqual(nil, err)
		assert.Equal("", bucket)
	}

	// another kind of url
	{
		bucket, err := bucketOfUploadAction("https://uploader.synq.fm")
		assert.NotEqual(nil, err)
		assert.Equal("", bucket)
	}

	// yet another kind of url
	{
		bucket, err := bucketOfUploadAction("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.NotEqual(nil, err)
		assert.Equal("", bucket)
	}

	// no bucket, and the bucket that is not there is not named "s3"
	{
		bucket, err := bucketOfUploadAction("https://s3.amazonaws.com")
		assert.NotEqual(nil, err)
		assert.Equal("", bucket)
	}

	// invalid url
	{
		bucket, err := bucketOfUploadAction("https://s3.amazonaws.com/%")
		assert.NotEqual(nil, err)
		assert.Equal("", bucket)
	}

	// not "amazonaws"
	{
		bucket, err := bucketOfUploadAction("https://s3.amazon.com/")
		assert.NotEqual(nil, err)
		assert.Equal("", bucket)
	}

	// not "com"
	{
		bucket, err := bucketOfUploadAction("https://s3.amazon.com/")
		assert.NotEqual(nil, err)
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
		assert.Equal(nil, err)
		assert.Equal("us-east-1", region)
	}

	// another region
	{
		region, err := regionOfUploadAction("https://a-bucket.s3-us-west-2.amazonaws.com")
		assert.Equal(nil, err)
		assert.Equal("us-west-2", region)
	}

	// yet another region with slash at the end
	{
		region, err := regionOfUploadAction("https://some-bucket.s3-ap-south-1.amazonaws.com/")
		assert.Equal(nil, err)
		assert.Equal("ap-south-1", region)
	}

	// not https but http
	{
		region, err := regionOfUploadAction("http://not-https.s3-sa-east-1.amazonaws.com")
		assert.Equal(nil, err)
		assert.Equal("sa-east-1", region)
	}

	// invalid region
	{
		region, err := bucketOfUploadAction("https://invalid-region.not-s3-eu-west-1.amazonaws.com")
		assert.NotEqual(nil, err)
		assert.Equal("", region)
	}

	// another kind of url
	{
		region, err := regionOfUploadAction("https://player.synq.fm")
		assert.NotEqual(nil, err)
		assert.Equal("", region)
	}

	// yet another kind of url
	{
		region, err := regionOfUploadAction("https://player.synq.fm/embed/55d4062f99454c9fb21e5186a09c2115")
		assert.NotEqual(nil, err)
		assert.Equal("", region)
	}

	// no "s3", and the region that is not there is not named "us-east-1"
	{
		region, err := regionOfUploadAction("https://bucket.amazonaws.com")
		assert.NotEqual(nil, err)
		assert.Equal("", region)
	}

	// invalid url
	{
		region, err := regionOfUploadAction("https://s3.amazonaws.com/%")
		assert.NotEqual(nil, err)
		assert.Equal("", region)
	}

	// not "amazonaws"
	{
		region, err := regionOfUploadAction("https://s3.amazon.com/")
		assert.NotEqual(nil, err)
		assert.Equal("", region)
	}

	// not "com"
	{
		region, err := regionOfUploadAction("https://s3.amazon.com/")
		assert.NotEqual(nil, err)
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
