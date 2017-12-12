package upload

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

func TestBucketOfUploadAction(t *testing.T) {
	assert := assert.New(t)

	// real example
	{
		bucket, err := BucketOfUploadAction("https://synqfm.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("synqfm", bucket)
	}

	// another bucket
	{
		bucket, err := BucketOfUploadAction("https://another-bucket.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("another-bucket", bucket)
	}

	// yet another bucket with slash at the end
	{
		bucket, err := BucketOfUploadAction("https://yet-another-bucket.s3.amazonaws.com/")
		assert.Nil(err)
		assert.Equal("yet-another-bucket", bucket)
	}

	// not https but http
	{
		bucket, err := BucketOfUploadAction("http://not-https.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("not-https", bucket)
	}

	// not the special "us-east-1" region
	{
		bucket, err := BucketOfUploadAction("https://a-bucket-in-another-region.s3-eu-west-1.amazonaws.com")
		assert.Nil(err)
		assert.Equal("a-bucket-in-another-region", bucket)
	}

	// invalid region
	{
		bucket, err := BucketOfUploadAction("https://invalid-region.not-s3-eu-west-1.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// another kind of url
	{
		bucket, err := BucketOfUploadAction("https://uploader.synq.fm")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// yet another kind of url
	{
		bucket, err := BucketOfUploadAction("https://uploader.synq.fm/uploader/55d4062f99454c9fb21e5186a09c2115?token=not32characters")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// no bucket, and the bucket that is not there is not named "s3"
	{
		bucket, err := BucketOfUploadAction("https://s3.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// invalid url
	{
		bucket, err := BucketOfUploadAction("https://bucket.s3.amazonaws.com/%")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// not "amazonaws"
	{
		bucket, err := BucketOfUploadAction("https://bucket.s3.amazon.com/")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// not "com"
	{
		bucket, err := BucketOfUploadAction("https://bucket.s3.amazonaws.org/")
		assert.NotNil(err)
		assert.Equal("", bucket)
	}

	// any non-empty bucket name
	{
		p := gopter.NewProperties(nil)

		p.Property("extract any bucket name", prop.ForAll(
			func(v string) bool {
				const format = "https://%s.s3.amazonaws.com"

				bucket, err := BucketOfUploadAction(fmt.Sprintf(format, v))

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
		region, err := RegionOfUploadAction("https://synqfm.s3.amazonaws.com")
		assert.Nil(err)
		assert.Equal("us-east-1", region)
	}

	// another region
	{
		region, err := RegionOfUploadAction("https://a-bucket.s3-us-west-2.amazonaws.com")
		assert.Nil(err)
		assert.Equal("us-west-2", region)
	}

	// yet another region with slash at the end
	{
		region, err := RegionOfUploadAction("https://some-bucket.s3-ap-south-1.amazonaws.com/")
		assert.Nil(err)
		assert.Equal("ap-south-1", region)
	}

	// not https but http
	{
		region, err := RegionOfUploadAction("http://not-https.s3-sa-east-1.amazonaws.com")
		assert.Nil(err)
		assert.Equal("sa-east-1", region)
	}

	// invalid region
	{
		region, err := RegionOfUploadAction("https://invalid-region.not-s3-eu-west-1.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// another kind of url
	{
		region, err := RegionOfUploadAction("https://player.synq.fm")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// yet another kind of url
	{
		region, err := RegionOfUploadAction("https://player.synq.fm/embed/55d4062f99454c9fb21e5186a09c2115")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// no "s3", and the region that is not there is not named "us-east-1"
	{
		region, err := RegionOfUploadAction("https://bucket.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// invalid url
	{
		region, err := RegionOfUploadAction("https://bucket.s3.amazonaws.com/%")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// not "amazonaws"
	{
		region, err := RegionOfUploadAction("https://bucket.s3.amazon.com/")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// not "com"
	{
		region, err := RegionOfUploadAction("https://bucket.s3.amazonaws.org/")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// start with "s3", but not with "s3-"
	{
		region, err := RegionOfUploadAction("https://invalid-region.s3eu-west-1.amazonaws.com")
		assert.NotNil(err)
		assert.Equal("", region)
	}

	// any non-empty region name
	{
		p := gopter.NewProperties(nil)

		p.Property("extract any region name", prop.ForAll(
			func(v string) bool {
				const format = "https://bucket.s3-%s.amazonaws.com"

				region, err := RegionOfUploadAction(fmt.Sprintf(format, v))

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
