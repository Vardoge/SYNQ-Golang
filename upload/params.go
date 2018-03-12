package upload

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/SYNQfm/helpers/common"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	DefaultCtype = "video/mp4"
	DefaultAcl   = "private"
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
	Region         string `json:"region"`
	Key            string `json:"key"`
	SuccessStatus  string `json:"success_action_status"`
	SignatureUrl   string `json:"signature_url"`
	VideoId        string `json:"video_id"`
	AssetId        string `json:"asset_id"`
}

type UploadRequest struct {
	AssetId     string `json:"asset_id"`
	ContentType string `json:"content_type"`
	Ext         string `json:"ext"`
	Type        string `json:"type"`
	Acl         string `json:"acl"`
}

func NewUploadRequest(data []byte) (UploadRequest, error) {
	var req UploadRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return req, err
	}
	err = req.ProcessCtype()
	if err != nil {
		return req, err
	}
	return req, nil
}

func parseCtype(ctype string) (string, error) {
	newType := ctype
	if len(strings.Split(newType, "/")) != 2 {
		return "", common.NewError("invalid ctype '%s'", newType)
	}
	switch ctype {
	case "image/jpg": // causes error in policy, change to jpeg
		newType = "image/jpeg"
	case "video/msvideo":
		newType = "video/avi"
	}
	return newType, nil
}

func (u *UploadRequest) ProcessCtype() error {
	ctype := u.ContentType
	if ctype != "" {
		c, err := parseCtype(ctype)
		if err != nil {
			return err
		}
		u.ContentType = c
	} else {
		u.ContentType = DefaultCtype
	}
	return nil
}

func (u UploadRequest) GetAcl() string {
	if u.Acl != "" {
		return u.Acl
	}
	return DefaultAcl
}

func (u UploadRequest) GetCType() string {
	return u.ContentType
}

func (u UploadRequest) GetType() string {
	if u.Type != "" {
		return u.Type
	}
	return u.ContentType
}

func (u UploadRequest) GetExt() string {
	if u.Ext != "" {
		return u.Ext
	}
	return common.GetFileExtension(u.GetCType())
}
