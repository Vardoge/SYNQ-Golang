package upload

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// UploaderSignatureRequest is the request that is sent when using the
// embeddable web uploader's request signing service.
type UploaderSignatureRequest struct {
	Headers string `json:"headers"`
}

// UploaderSignatureResponse is the response that is received when using the
// embeddable web uploader's request signing service.
type UploaderSignatureResponse struct {
	Signature string `json:"signature"`
}

// bucketOfUploadAction parses an "action" URL as received with GetUploadInfo,
// and returns the bucket name part of that URL.
//
// Example:
//         const a = "https://synqfm.s3.amazonaws.com"
//
//         bucket, err := bucketOfUploadAction(a)
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(bucket) // prints synqfm
func BucketOfUploadAction(actionURL string) (string, error) {
	u, err := url.Parse(actionURL)
	if err != nil {
		return "", err
	}

	if u.Hostname() == "127.0.0.1" {
		// this is a test server, return a fake bucket
		return "synq-abucket", nil
	}

	hs := strings.Split(u.Host, ".")
	if len(hs) != 4 {
		return "", errors.New("Invalid action URL. " +
			"Not exactly 4 period-separated words in host.")
	}
	if !strings.HasPrefix(hs[1], "s3") {
		return "", errors.New("Invalid action URL. " +
			"Second word in period-separated host does not start with s3")
	}
	if hs[2] != "amazonaws" {
		return "", errors.New("Invalid action URL. " +
			"Third word in period-separated host is not amazonaws")
	}
	if hs[3] != "com" {
		return "", errors.New("Invalid action URL. " +
			"Fourth word in period-separated host is not com")
	}

	return hs[0], nil
}

// regionOfUploadAction parses an "action" URL as received with GetUploadInfo,
// and returns the region of the bucket that is the URL refers to.
//
// This relies on heuristics, and will not work with certain styles of URLs.
//
// See: http://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
//
// Example:
//         const a = "https://synqfm.s3.amazonaws.com"
//
//         region, err := regionOfUploadAction(a)
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(region) // prints us-east-1
func RegionOfUploadAction(actionURL string) (string, error) {
	u, err := url.Parse(actionURL)
	if err != nil {
		return "", err
	}

	if u.Hostname() == "127.0.0.1" {
		// if its a test server, return us-east-1
		return "us-east-1", nil
	}

	hs := strings.Split(u.Host, ".")
	if len(hs) != 4 {
		return "", errors.New("Invalid action URL. " +
			"Not exactly 4 period-separated words in host.")
	}
	if !strings.HasPrefix(hs[1], "s3") {
		return "", errors.New("Invalid action URL. " +
			"Second word in period-separated host does not start with s3")
	}
	if hs[2] != "amazonaws" {
		return "", errors.New("Invalid action URL. " +
			"Third word in period-separated host is not amazonaws")
	}
	if hs[3] != "com" {
		return "", errors.New("Invalid action URL. " +
			"Fourth word in period-separated host is not com")
	}

	regionPart := hs[1]

	// us-east-1 is the region if nothing else is specified.
	if regionPart == "s3" {
		return "us-east-1", nil
	}

	// If the region part of the host is not exactly "s3", then it must be
	// "s3-something".
	if !strings.HasPrefix(regionPart, "s3-") {
		return "", errors.New("Invalid action URL. " +
			`Second word in period-separated host is not "s3", and does not start with "s3-".`)
	}

	return regionPart[len("s3-"):], nil
}

// TokenOfUploaderURL parses an uploader URL string, and returns its token
// parameter.
//
// Example:
//         const u = "https://uploader.synq.fm/uploader/" +
//         "00000000000000000000000000000000" +
//         "?token=11111111111111111111111111111111"
//
//         token, err := TokenOfUploaderURL(u)
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(token) // prints 11111111111111111111111111111111
func TokenOfUploaderURL(uploaderURL string) (string, error) {
	u, err := url.Parse(uploaderURL)
	if err != nil {
		return "", err
	}

	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	token := values.Get("token")
	if token == "" {
		return "", errors.New("Found no token parameter in URL.")
	}

	return token, nil
}

// ReformatXAmzDate reformats the contents of a X-Amz-Date header into the
// format that https://uploader.synq.fm/uploader/signature expects.
//
// Example:
//         const in = "20060102T150405Z"
//
//         out, err := ReformatXAmzDate(in)
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(out) // Mon, 02 Jan 2006 15:04:05 UTC
func ReformatXAmzDate(in string) (string, error) {
	t, err := time.Parse("20060102T150405Z", in)
	if err != nil {
		return "", err
	}
	return t.Format(time.RFC1123), nil
}

// RewriteXAmzDateHeader rewrites the X-Amz-Date header, of an http.Header,
// into the format that https://uploader.synq.fm/uploader/signature expects.
//
// Example:
//         h := http.Header{}
//         h.Set("X-Amz-Date", "20060102T150405Z")
//
//         err := RewriteXAmzDateHeader(h)
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(h.Get("X-Amz-Date")) // Mon, 2 Jan 2006 15:04:05 UTC
func RewriteXAmzDateHeader(h http.Header) error {
	in := h.Get("X-Amz-Date")
	if in == "" {
		return errors.New("Missing header: X-Amz-Date.")
	}
	out, err := ReformatXAmzDate(in)
	if err != nil {
		return err
	}
	delete(h, "X-Amz-Date") // TODO(mastensg): enough to just set and not delete?
	h.Set("X-Amz-Date", out)
	return nil
}
