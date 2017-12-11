package synq

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

type UploadParameters struct {
	Action         string `json:"action"`
	AwsAccessKeyId string `json:"AWSAccessKeyId"`
	ContentType    string `json:"Content-Type"`
	Policy         string `json:"policy"`
	Signature      string `json:"signature"`
	Acl            string `json:"acl"`
	Key            string `json:"key"`
	SuccessStatus  string `json:"success_action_status"`
	SignatureUrl   string `json:"signature_url"`
	VideoId        string `json:"video_id"`
	AssetId        string `json:"asset_id"`
}

// This is the struct that contains all the AWS settings
type AwsUpload struct {
	UploadParams   UploadParameters
	UploaderFormat string
	Uploader       *s3manager.Uploader
}

// UploadInfo is retrieved from the Unicorn API, so we're creating an AwsUpload from it
func NewAwsUpload(params UploadParameters) (*AwsUpload, error) {
	au := &AwsUpload{
		UploadParams:   params,
		UploaderFormat: UploaderSignatureUrlFormat,
	}
	provider := credentials.StaticProvider{}
	provider.Value.AccessKeyID = multipartUploadAwsAccessKeyId
	provider.Value.SecretAccessKey = multipartUploadAwsSecretAccessKey
	credentials := credentials.NewCredentials(&provider)

	region, e := au.GetRegion()
	if e != nil {
		return au, e
	}
	// session
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials,
		Region:      aws.String(region),
	})
	if err != nil {
		return au, err
	}

	svc := s3.New(sess)

	// sign handler
	signer := au.Signer()
	svc.Handlers.Sign.PushBack(signer)

	// s3manager uploader
	au.Uploader = s3manager.NewUploaderWithClient(svc)
	return au, nil
}

func (a *AwsUpload) Url() string {
	return a.UploadParams.Action
}

func (a *AwsUpload) Key() string {
	return a.UploadParams.Key
}

func (a *AwsUpload) Acl() string {
	return a.UploadParams.Acl
}

func (a *AwsUpload) ContentType() string {
	return a.UploadParams.ContentType
}

func (a *AwsUpload) AwsKeyId() string {
	return a.UploadParams.AwsAccessKeyId
}

func (a *AwsUpload) UploaderSigUrl() string {
	// take the UploadParams signature and append it to the uploader url
	return a.UploadParams.SignatureUrl
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
func (a AwsUpload) GetBucket() (string, error) {
	return bucketOfUploadAction(a.Url())
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
func (a AwsUpload) GetRegion() (string, error) {
	return regionOfUploadAction(a.Url())
}

func (a *AwsUpload) Upload(body io.Reader) (*s3manager.UploadOutput, error) {
	// upload parameters
	acl := a.Acl()
	bucket, err := a.GetBucket()
	if err != nil {
		return nil, err
	}
	contentType := a.ContentType()
	key := a.Key()
	uploadInput := &s3manager.UploadInput{
		ACL:         &acl,
		Bucket:      &bucket,
		Body:        body,
		ContentType: &contentType,
		Key:         &key,
	}
	return a.Uploader.Upload(uploadInput)
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
//         signer := a.Signer()
//
//         // Register handler as the last handler of the signing phase.
//         svc.Handlers.Sign.PushBack(signer)
//
//         // S3 requests are now signed by signer().

func (a AwsUpload) Signer() func(r *request.Request) {
	signer := func(r *request.Request) {
		err := a.SignRequest(r.HTTPRequest)
		if err != nil {
			return // TODO(mastensg): how to report errors from handlers?
		}
	}

	return signer
}

func (a *AwsUpload) SignRequest(r *http.Request) error {
	if err := RewriteXAmzDateHeader(r.Header); err != nil {
		return err
	}

	x_amz_date := r.Header.Get("X-Amz-Date")

	bucket, err := a.GetBucket()
	if err != nil {
		return err
	}
	// construct "headers" string to send to
	// https://uploader.synq.fm/uploader/signature
	headers := ""
	if r.URL.RawQuery == "uploads=" {
		// Initiate multi-part upload

		headers = fmt.Sprintf("%s\n\n%s\n\nx-amz-acl:%s\nx-amz-date:%s\n/%s%s",
			r.Method,
			a.ContentType(),
			a.Acl(),
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

	signature, err := a.UploaderSignature(headers)
	if err != nil {
		return err
	}

	// rewrite authorization header(s)
	delete(r.Header, "X-Amz-Content-Sha256")
	delete(r.Header, "Authorization")
	authorization := fmt.Sprintf("AWS %s:%s", a.AwsKeyId(), signature)
	r.Header.Set("Authorization", authorization)

	return nil
}

// UploaderSignature uses the backend of the embeddable web uploader to sign
// multipart upload requests.
func (a *AwsUpload) UploaderSignature(headers string) ([]byte, error) {
	url := a.UploaderSigUrl()

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
