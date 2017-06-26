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
	//"os"
	//"time"

	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploaderSignatureUrlFormat is a printf format string which is used when
// signing multipart upload requests.
const UploaderSignatureUrlFormat = "https://uploader.synq.fm/uploader/signature/%s?token=%s"

// UploaderSignatureResponse is the response that is received when using the
// embeddable web uploader's request signing service.
type UploaderSignatureResponse struct {
	Signature string `json:"signature"`
}

// UploaderSignature uses the backend of the embeddable web uploader to sign
// multipart upload requests.
func UploaderSignature(url_fmt, video_id, token string, payload []byte) ([]byte, error) {
	url := fmt.Sprintf(url_fmt, video_id, token)

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type responseBody struct {
		Signature string `json:"signature"`
	}

	rb := responseBody{}
	err = json.Unmarshal(body, &rb)
	if err != nil {
		return nil, err
	}

	return []byte(rb.Signature), nil
}

// tokenOfUploaderURL parses an uploader URL string, and returns its token parameter.
//
// Example:
//         token, err := tokenOfUploaderURL("https://uploader.synq.fm/uploader/00000000000000000000000000000000?token=11111111111111111111111111111111")
//         if err != nil {
//                 log.Fatal(err)
//         }
//         fmt.Println(token)
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
func makeMultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, uploaderURL string) func(r *request.Request) {
	signer := func(r *request.Request) {
	}

	return signer
}

// multipartUpload uploads a file as the video's original_file.
// This procedure will use Amazon S3's Multipart Upload API:
// http://docs.aws.amazon.com/AmazonS3/latest/dev/uploadobjusingmpu.html
//
// This is the internal function to make uploads, which is called by the public
// MultipartUpload. This function uses s3manager from aws-sdk-go to upload.
func multipartUpload(body io.Reader, acl, awsAccessKeyId, bucket, contentType, key, uploaderURL string) (*s3manager.UploadOutput, error) {
	token, err := tokenOfUploaderURL(uploaderURL)
	if err != nil {
		return nil, err
	}
	fmt.Println(token)

	sess := session.Must(session.NewSession())

	svc := s3.New(sess)

	// sign handler
	signer := makeMultipartUploadSigner(acl, awsAccessKeyId, bucket, contentType, key, uploaderURL)
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
