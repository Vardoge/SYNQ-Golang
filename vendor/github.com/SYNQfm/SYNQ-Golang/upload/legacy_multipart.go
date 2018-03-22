package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

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

// UploaderSignatureUrlFormat is a printf format string which is used when
// signing multipart upload requests.
// TODO(mastensg): Determine this format (or at least prefix) at runtime, from
// the Synq HTTP API.
const UploaderSignatureUrlFormat = "https://uploader.synq.fm/uploader/signature/%s?token=%s"

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

	if resp.StatusCode != http.StatusOK {
		// TODO(mastensg): report status and maybe body
		// TODO(mastensg): handle known error responses specifically
		return nil, errors.New("HTTP response status not OK.")
	}

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

// multipartUploadSignRequest signs a single AWS request using SYNQ's
// embeddable web uploader's remote signature procedure.
func multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uploaderSignatureUrlFormat string, r *http.Request) error {
	if err := RewriteXAmzDateHeader(r.Header); err != nil {
		return err
	}

	x_amz_date := r.Header.Get("X-Amz-Date")

	// construct "headers" string to send to
	// https://uploader.synq.fm/uploader/signature
	headers := ""
	if r.URL.RawQuery == "uploads=" {
		// Initiate multi-part upload

		headers = fmt.Sprintf("%s\n\n%s\n\nx-amz-acl:%s\nx-amz-date:%s\n/%s%s",
			r.Method,
			contentType,
			acl,
			x_amz_date,
			bucket,
			r.URL.Path+"?uploads",
		)
	} else if r.Method == "PUT" {
		// Upload one part

		headers = fmt.Sprintf("%s\n\n%s\n\nx-amz-date:%s\n/%s%s",
			r.Method,
			"",
			x_amz_date,
			bucket,
			r.URL.Path+"?"+r.URL.RawQuery,
		)
	} else if r.Method == "POST" {
		// Finish multi-part upload

		// TODO(mastensg): the content-type header set by
		// aws-sdk-go is not exactly the one expected by
		// uploader/signature, maybe
		r.Header.Set("Content-Type", "application/xml; charset=UTF-8")

		headers = fmt.Sprintf("%s\n\n%s\n\nx-amz-date:%s\n/%s%s",
			r.Method,
			r.Header.Get("Content-Type"),
			x_amz_date,
			bucket,
			r.URL.Path+"?"+r.URL.RawQuery,
		)
	} else {
		return errors.New("Unknown request type.")
	}

	signature, err := UploaderSignature(uploaderSignatureUrlFormat, video_id, token, headers)
	if err != nil {
		return err
	}

	// rewrite authorization header(s)
	delete(r.Header, "X-Amz-Content-Sha256")
	delete(r.Header, "Authorization")
	authorization := fmt.Sprintf("AWS %s:%s", awsAccessKeyId, signature)
	r.Header.Set("Authorization", authorization)

	return nil
}

// MultipartUploadSigner returns a function that can be added to an s3 client's
// list of handlers. The function will take over signing of requests from
// aws-sdk-go.
//
// The signer function uses SYNQ's embeddable web uploader's remote procedure
// to sign requests.
//
// This function is used by the internal multipartUpload function.
//
// Example:
//         // AWS session.
//         sess := session.Must(session.NewSession())
//
//         // S3 service client.
//         svc := s3.New(sess)
//
//         // Signer function. Determine the arguments somehow.
//         signer := MultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, token, video_id)
//
//         // Register handler as the last handler of the signing phase.
//         svc.Handlers.Sign.PushBack(signer)
//
//         // S3 requests are now signed by signer().
func MultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uploaderSignatureUrlFormat string) func(r *request.Request) {
	signer := func(r *request.Request) {
		err := multipartUploadSignRequest(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, uploaderSignatureUrlFormat, r.HTTPRequest)
		if err != nil {
			return // TODO(mastensg): how to report errors from handlers?
		}
	}

	return signer
}

// multipartUpload uploads a file as the video's original_file.
// This procedure will use Amazon S3's Multipart Upload API:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/uploadobjusingmpu.html
//
// This is the internal function to make uploads, which is called by the public
// MultipartUpload. This function uses s3manager from aws-sdk-go to manage the
// process of uploading in multiple parts. In particular, this will start
// several goroutines that will upload parts concurrently.
func MultipartUpload(body io.Reader, acl, actionURL, awsAccessKeyId, contentType, key, uploaderURL, video_id string) (*s3manager.UploadOutput, error) {
	bucket, err := BucketOfUploadAction(actionURL)
	if err != nil {
		return nil, err
	}

	region, err := RegionOfUploadAction(actionURL)
	if err != nil {
		return nil, err
	}

	token, err := TokenOfUploaderURL(uploaderURL)
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
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	// sign handler
	signer := MultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, token, video_id, UploaderSignatureUrlFormat)
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
