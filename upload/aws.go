package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type AwsUploadF interface {
	Upload(io.Reader) (*s3manager.UploadOutput, error)
}

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

type V4Request struct {
	Method   string            `json:"method"`
	Action   string            `json:"action"`
	Path     string            `json:"path"`
	RawQuery string            `json:"raw_query"`
	Headers  map[string]string `json:"headers"`
}

type V4Response struct {
	Authorization string `json:"authorization"`
	Date          string `json:"date"`
}

func CreateV4Request(action string, req *request.Request) V4Request {
	r := V4Request{}
	hreq := req.HTTPRequest
	r.Method = hreq.Method
	r.Action = action
	r.Path = hreq.URL.Path
	r.RawQuery = hreq.URL.RawQuery
	r.Headers = make(map[string]string)
	for header, _ := range req.SignedHeaderVals {
		r.Headers[header] = hreq.Header.Get(header)
	}
	return r
}

func (r *V4Request) BuildRequest() *http.Request {
	req, _ := http.NewRequest(r.Method, r.Action, nil)
	req.URL.Path = r.Path
	req.URL.RawQuery = r.RawQuery
	for header, val := range r.Headers {
		req.Header.Set(header, val)
	}
	return req
}

func (r *V4Request) Sign(awsKey, awsSecret string) (resp V4Response, err error) {
	// use the v4 signer automatically
	provider := credentials.StaticProvider{}
	provider.Value.AccessKeyID = awsKey
	provider.Value.SecretAccessKey = awsSecret
	cred := credentials.NewCredentials(&provider)
	signer := v4.NewSigner(cred)
	req := r.BuildRequest()
	_, err = signer.Sign(req, nil, "s3", "us-east-1", time.Now())
	if err != nil {
		return resp, err
	}
	date := req.Header.Get("X-Amz-Date")
	auth := req.Header.Get("Authorization")
	return V4Response{
		Authorization: auth,
		Date:          date,
	}, nil
}

var CreatorFn func(UploadParameters) (AwsUploadF, error)

func init() {
	CreatorFn = NewAwsUpload
}

// This is the struct that contains all the AWS settings
type AwsUpload struct {
	UploadParams UploadParameters
	Uploader     *s3manager.Uploader
}

// UploadParameters is retrieved from the Unicorn API, so we're creating an AwsUpload from the settings
func NewAwsUpload(params UploadParameters) (AwsUploadF, error) {
	au := &AwsUpload{
		UploadParams: params,
	}
	provider := credentials.StaticProvider{}
	// use dummy values
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

	customSigner := true
	if customSigner {
		// sign handler
		signer := au.Signer()
		svc.Handlers.Sign.PushBack(signer)
	}

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
	return BucketOfUploadAction(a.Url())
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
	return RegionOfUploadAction(a.Url())
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
		err := a.SignRequest(r)
		if err != nil {
			return // TODO(mastensg): how to report errors from handlers?
		}
	}

	return signer
}

func (a *AwsUpload) ServerSignV2(r *request.Request) (string, error) {
	v4 := CreateV4Request(a.UploadParams.Action, r)
	resp, err := a.V4Sig(v4)
	if err != nil {
		return "", err
	}
	if resp.Date != "" {
		// reset the data
		r.HTTPRequest.Header.Set("X-Amz-Date", resp.Date)
	}
	return resp.Authorization, nil
}

func (a *AwsUpload) ServerSignV1(r *http.Request) (string, error) {
	bucket, err := a.GetBucket()
	if err != nil {
		return "", err
	}
	if err := RewriteXAmzDateHeader(r.Header); err != nil {
		return "", err
	}
	headers := ""
	x_amz_date := r.Header.Get("X-Amz-Date")
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
		return "", errors.New("Unknown request type.")
	}
	sig, err := a.UploaderSignature(headers)
	if err != nil {
		return "", err
	}
	delete(r.Header, "X-Amz-Content-Sha256")
	return fmt.Sprintf("AWS %s:%s", a.AwsKeyId(), sig), nil
}

// This runs as a handler within the Sign HandlerList and uses unicorn to sign the request
// This will replace whats in awssdk-go/aws/signer/v4/v4.go and its own "signWithBody" method
// https://docs.aws.amazon.com/general/latest/gr/signature-version-2.html
// https://docs.aws.amazon.com/general/latest/gr/sigv4_signing.html
// https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html
func (a *AwsUpload) SignRequest(r *request.Request) error {
	// construct "headers" string to send to
	// https://uploader.synq.fm/uploader/signature
	auth, err := a.ServerSignV2(r)
	if err != nil {
		return err
	}

	// rewrite authorization header(s)
	delete(r.HTTPRequest.Header, "Authorization")
	r.HTTPRequest.Header.Set("Authorization", auth)

	return nil
}

// UploaderSignature uses the backend of the embeddable web uploader to sign
// multipart upload requests.
func (a *AwsUpload) UploaderSignature(headers string) ([]byte, error) {
	reqStruct := UploaderSignatureRequest{Headers: headers}
	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	respBody, err := a.Request(reqBody)

	// parse response
	respStruct := UploaderSignatureResponse{}
	err = json.Unmarshal(respBody, &respStruct)
	if err != nil {
		log.Println("error unmarshaling", err.Error())
		return nil, err
	}
	return []byte(respStruct.Signature), nil
}

func (a *AwsUpload) V4Sig(req V4Request) (resp V4Response, err error) {
	secret := os.Getenv("AWS_SECRET")
	if secret != "" {
		return req.Sign(a.UploadParams.AwsAccessKeyId, secret)
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return resp, err
	}
	respBody, err := a.Request(reqBody)
	if err != nil {
		return resp, err
	}
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (a *AwsUpload) Request(body []byte) ([]byte, error) {
	url := a.UploaderSigUrl()

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("could not call %s : %s\n", url, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// TODO(mastensg): report status and maybe body
		// TODO(mastensg): handle known error responses specifically
		log.Printf("invalid response code %d from response\n", resp.StatusCode)
		return nil, errors.New("HTTP response status not OK.")
	}
	// read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response body", err.Error())
		return nil, err
	}
	return respBody, nil
}
