package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/SYNQ-Golang/test_server"
)

const (
	SYNQ_LEGACY_VERSION = "v0"
	SYNQ_VERSION        = "v1"
)

var DEFAULT_CRED_FILE = os.Getenv("HOME") + "/.synq/credentials.json"

type ApiSetup struct {
	Key     string
	Version string
	Url     string
}

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
	if len(file) > 0 && file[0] != "" {
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

func SetupSynq() synq.Api {
	api := SetupSynqApi()
	return api.(synq.Api)
}

func SetupSynqV2() synq.ApiV2 {
	config := GetSetupByEnv(SYNQ_VERSION)
	api := SetupSynqApi(config)
	return api.(synq.ApiV2)
}

func GetSetupByEnv(version string) ApiSetup {
	key := os.Getenv(fmt.Sprintf("SYNQ_API%s_KEY", version))
	url := os.Getenv(fmt.Sprintf("SYNQ_API%s_URL", version))
	return ApiSetup{
		Key:     key,
		Version: version,
		Url:     url,
	}
}

func SetupSynqApi(setup ...ApiSetup) (api synq.ApiF) {
	var config ApiSetup
	if len(setup) > 0 {
		config = setup[0]
	} else {
		config = GetSetupByEnv("")
	}
	if config.Key == "" {
		log.Println("WARNING : no Synq API key specified")
	}
	if strings.Contains(config.Key, ".") || config.Version == SYNQ_VERSION {
		api = synq.NewV2(config.Key)
	} else {
		api = synq.NewV1(config.Key)
	}
	if config.Url != "" {
		api.SetUrl(config.Url)
	}
	return api
}

func SetupForTestV1() synq.Api {
	server := test_server.SetupServer(SYNQ_LEGACY_VERSION)
	url := server.GetUrl()
	api := synq.NewV1(test_server.API_KEY)
	api.SetUrl(url)
	return api
}

func SetupForTestV2() synq.ApiV2 {
	server := test_server.SetupServer(SYNQ_VERSION)
	url := server.GetUrl()
	api := synq.NewV2(test_server.TEST_AUTH)
	api.SetUrl(url)
	return api
}
