package helper

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/synq"
)

const (
	DEFAULT_CRED_FILE = "~/.synq/credentials.json"
)

type ApiSetting struct {
	// only api_key is required
	ApiKey        string `json:"api_key"`
	Url           string `json:"api_url"`
	Timeout       int    `json:"timeout"`
	UploadTimeout int    `json:"upload_timeout"`
	User          string `json:"user"`
	Password      string `json:"password"`
}

type ApiSet struct {
	v1    ApiSetting `json:"v1"`
	v2    ApiSetting `json:"v2"`
	ApiV1 synq.Api   `json:"-"`
	ApiV2 synq.ApiV2 `json:"-"`
}

func (a ApiSetting) SetupV1() synq.Api {
	api := synq.NewV1(a.ApiKey)
	if a.Url != "" {
		api.SetUrl(a.Url)
	}
	if api.Timeout > 0 {
		api.Timeout = api.Timeout * time.Second
	}
	if api.UploadTimeout > 0 {
		api.Timeout = api.UploadTimeout * time.Second
	}
	return api
}

func (a ApiSetting) SetupV2() synq.ApiV2 {
	var api synq.ApiV2
	if a.ApiKey != "" {
		api = synq.NewV2(a.ApiKey)
	} else if a.User != "" && a.Password != "" {
		api, _ = synq.Login(a.User, a.Password)
	}
	if a.Url != "" {
		api.SetUrl(a.Url)
	}
	if api.Timeout > 0 {
		api.Timeout = api.Timeout * time.Second
	}
	if api.UploadTimeout > 0 {
		api.Timeout = api.UploadTimeout * time.Second
	}
	return api
}

func (a ApiSetting) Valid() bool {
	return a.ApiKey != "" || (a.User != "" && a.Password != "")
}

func (a *ApiSet) Setup() {
	if a.v1.Valid() {
		a.ApiV1 = a.v1.SetupV1()
	}
	if a.v2.Valid() {
		a.ApiV2 = a.v2.SetupV2()
	}
}

func LoadFromFile(file ...string) (*ApiSet, error) {
	credFile := DEFAULT_CRED_FILE
	if len(file) > 0 {
		credFile = file[0]
	}
	api := &ApiSet{}
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return api, err
	}
	bytes, _ := ioutil.ReadFile(credFile)
	json.Unmarshal(bytes, api)
	api.Setup()
	return api, nil
}
