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
	Url           string `json:"api_url,omitempty"`
	Timeout       int    `json:"timeout,omitempty"`
	UploadTimeout int    `json:"upload_timeout,omitempty"`
	User          string `json:"user,omitempty"`
	Password      string `json:"password,omitempty"`
}

type ApiSet struct {
	V1    ApiSetting `json:"v1"`
	V2    ApiSetting `json:"v2"`
	ApiV1 synq.Api   `json:"-"`
	ApiV2 synq.ApiV2 `json:"-"`
}

func (a ApiSetting) Configure(api synq.ApiF) {
	if a.Timeout > 0 {
		api.SetTimeout("", time.Duration(a.Timeout)*time.Second)
	}
	if a.UploadTimeout > 0 {
		api.SetTimeout("upload", time.Duration(a.UploadTimeout)*time.Second)
	}
	if a.Url != "" {
		api.SetUrl(a.Url)
	}
}

func (a ApiSetting) SetupV1() synq.Api {
	api := synq.NewV1(a.ApiKey)
	a.Configure(api)
	return api
}

func (a ApiSetting) SetupV2() synq.ApiV2 {
	var api synq.ApiV2
	if a.ApiKey != "" {
		api = synq.NewV2(a.ApiKey)
		if a.Url != "" {
			api.SetUrl(a.Url)
		}
	} else if a.User != "" && a.Password != "" {
		urls := []string{}
		if a.Url != "" {
			urls = append(urls, a.Url)
		}
		api, _ = synq.Login(a.User, a.Password, urls...)
	}
	a.Configure(api)
	return api
}

func (a ApiSetting) Valid() bool {
	return a.ApiKey != "" || (a.User != "" && a.Password != "")
}

func (a *ApiSet) Setup() {
	if a.V1.Valid() {
		a.ApiV1 = a.V1.SetupV1()
	}
	if a.V2.Valid() {
		a.ApiV2 = a.V2.SetupV2()
	}
}

func LoadFromFile(file ...string) (*ApiSet, error) {
	credFile := DEFAULT_CRED_FILE
	if len(file) > 0 {
		credFile = file[0]
	}
	set := &ApiSet{}
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		return set, err
	}
	if bytes, err := ioutil.ReadFile(credFile); err != nil {
		return set, err
	} else {
		json.Unmarshal(bytes, set)
		set.Setup()
	}
	return set, nil
}
