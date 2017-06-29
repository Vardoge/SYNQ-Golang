package synq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// These Amazon Web Services credentials are provided to the AWS SDK, which is
// used to upload content in multiple parts. There is no IAM user with these
// credentials; they are supplied because the AWS SDK requires some credentials
// to attempt to start uploading. This package replaces the AWS SDK's request
// signing method with its own method.
const multipartUploadAwsAccessKeyId = "AAAAAAAAAAAAAAAAAAAA"
const multipartUploadAwsSecretAccessKey = "ssssssssssssssssssssssssssssssssssssssss"

// TODO(mastensg): Determine region from bucket, or /v1/video/uploader
const multipartUploadS3BucketRegion = "us-east-1"

// UploaderSignatureUrlFormat is a printf format string which is used when
// signing multipart upload requests.
// TODO(mastensg): Determine this format (or at least prefix) at runtime, from
// the Synq HTTP API.
const UploaderSignatureUrlFormat = "https://uploader.synq.fm/uploader/signature/%s?token=%s"

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

// UploaderSignature uses the backend of the embeddable web uploader to sign
// multipart upload requests.
func UploaderSignature(url_fmt, video_id, token, headers string) ([]byte, error) {
	url := fmt.Sprintf(url_fmt, video_id, token)

	// construct request body
	reqStruct := UploaderSignatureRequest{Headers: headers}
	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	// perform request
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// parse response
	respStruct := UploaderSignatureResponse{}
	err = json.Unmarshal(respBody, &respStruct)
	if err != nil {
		return nil, err
	}

	return []byte(respStruct.Signature), nil
}

// tokenOfUploaderURL parses an uploader URL string, and returns its token
// parameter.
//
// Example:
//         const u = "https://uploader.synq.fm/uploader/" +
//         "00000000000000000000000000000000" +
//         "?token=11111111111111111111111111111111"
//
//         token, err := tokenOfUploaderURL(u)
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(token) // prints 11111111111111111111111111111111
func tokenOfUploaderURL(uploaderURL string) (string, error) {
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

// makeMultipartUploadSigner returns a function that can be added to an s3
// client's list of handlers. The function will take over signing of requests
// from aws-sdk-go.
//
// The signer function uses SYNQ's embeddable web uploader's remote procedure
// to sign requests.
//
// Example:
//         // AWS session.
//         sess := session.Must(session.NewSession())
//
//         // S3 service client.
//         svc := s3.New(sess)
//
//         signer := makeSigner()
//
//         // Register handler as the last handler of the signing phase.
//         svc.Handlers.Sign.PushBack(signer)
//
//         // S3 requests are now signed by signer().
func makeMultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, token, video_id string) func(r *request.Request) {
	signer := func(r *request.Request) {
		hr := r.HTTPRequest

		// rewrite the X-Amz-Date header into the format that
		// https://uploader.synq.fm/uploader/signature expects
		{
			x_amz_date_in := hr.Header.Get("X-Amz-Date")
			if x_amz_date_in == "" {
				return // TODO(mastensg): how to report errors from handlers?
			}
			x_amz_date_t, err := time.Parse("20060102T150405Z", x_amz_date_in)
			if err != nil {
				return // TODO(mastensg): how to report errors from handlers?
			}
			x_amz_date := x_amz_date_t.Format("Mon, 2 Jan 2006 15:04:05 MST")
			delete(hr.Header, "X-Amz-Date") // TODO(mastensg): enough to just set and not delete?
			hr.Header.Set("X-Amz-Date", x_amz_date)
		}

		x_amz_date := hr.Header.Get("X-Amz-Date")

		// extract url query parameters and http headers in the formats
		// that https://uploader.synq.fm/uploader/signature expects
		query := hr.URL.Path
		headers := ""
		if hr.URL.RawQuery == "uploads=" {
			// Initiate multi-part upload
			query += "?uploads"

			// TODO(mastensg): parameterize bucket name, content-type, acl
			headers = fmt.Sprintf("%s\n\n%s\n\n%s\nx-amz-date:%s\n/synqfm%s",
				hr.Method,
				"video/mp4",
				"x-amz-acl:public-read",
				x_amz_date,
				query,
			)
		} else if hr.Method == "PUT" {
			// Upload one part
			query += "?" + hr.URL.RawQuery

			// TODO(mastensg): parameterize bucket name
			headers = fmt.Sprintf("%s\n\n%s\n\nx-amz-date:%s\n/synqfm%s",
				hr.Method,
				"",
				x_amz_date,
				query,
			)
		} else if hr.Method == "POST" {
			// Finish multi-part upload
			query += "?" + hr.URL.RawQuery

			// TODO(mastensg): parameterize bucket name, content-type(?)
			headers = fmt.Sprintf("%s\n\n%s\n\nx-amz-date:%s\n/synqfm%s",
				hr.Method,
				"application/xml; charset=UTF-8",
				x_amz_date,
				query,
			)

			// TODO(mastensg): the content-type header set by
			// aws-sdk-go is not exactly the one expected by
			// uploader/signature, maybe
			hr.Header.Set("Content-Type", "application/xml; charset=UTF-8")
		} else {
			// Unknown request type
			return // TODO(mastensg): how to report errors from handlers?
		}

		signature, err := UploaderSignature(UploaderSignatureUrlFormat, video_id, token, headers)
		if err != nil {
			return // TODO(mastensg): how to report errors from handlers?
		}

		// rewrite authorization header(s)
		delete(hr.Header, "X-Amz-Content-Sha256") //TODO(mastensg): can this be left in?
		delete(hr.Header, "Authorization")
		authorization := fmt.Sprintf("AWS %s:%s", awsAccessKeyId, signature)
		hr.Header.Set("Authorization", authorization)
	}

	return signer
}

// multipartUpload uploads a file as the video's original_file.
// This procedure will use Amazon S3's Multipart Upload API:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/uploadobjusingmpu.html
//
// This is the internal function to make uploads, which is called by the public
// MultipartUpload. This function uses s3manager from aws-sdk-go to upload.
func multipartUpload(body io.Reader, acl, awsAccessKeyId, bucket, contentType, key, uploaderURL, video_id string) (*s3manager.UploadOutput, error) {
	token, err := tokenOfUploaderURL(uploaderURL)
	if err != nil {
		return nil, err
	}

	// credentials
	provider := credentials.StaticProvider{}
	provider.Value.AccessKeyID = multipartUploadAwsAccessKeyId
	provider.Value.SecretAccessKey = multipartUploadAwsSecretAccessKey
	credentials := credentials.NewCredentials(&provider)

	// session
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials,
		Region:      aws.String(multipartUploadS3BucketRegion),
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	// sign handler
	signer := makeMultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, token, video_id)
	svc.Handlers.Sign.PushBack(signer)

	// s3manager uploader
	uploader := s3manager.NewUploaderWithClient(svc)

	// upload parameters
	uploadInput := &s3manager.UploadInput{
		ACL:         &acl,
		Body:        body,
		Bucket:      &bucket,
		ContentType: &contentType,
		Key:         &key,
	}

	return uploader.Upload(uploadInput)
}
