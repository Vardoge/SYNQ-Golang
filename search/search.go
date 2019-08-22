package search

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/SYNQfm/SYNQ-Golang/synq"
)

type SearchRequest struct {
	Method      string
	Url         string
	Token       string
	RequestBody *SearchRequestBody
}

// SearchRequestBody is the expected request body for search route
type SearchRequestBody struct {
	Query   string                 `json:"query"`
	Params  map[string]interface{} `json:"params"`
	Options SearchOptions          `json:"options"`
}

// SearchOptions is the options for showing search responses
type SearchOptions struct {
	IgnoreHits          bool `json:"ignore_hits"`
	IgnoreHighlightHits bool `json:"ignore_highlight_hits"`
	IgnoreIDList        bool `json:"ignore_id_list"`
	WithThumbnails      bool `json:"with_thumbnails"`
}

type SearchResponse struct {
	Page        int          `json:"page"`
	NbHits      int          `json:"nbHits"`
	NbPage      int          `json:"nbPages"`
	HitsPerPage int          `json:"hitsPerPage`
	IdList      []string     `json:"idList"`
	Hits        []synq.Asset `json:"hits"`
}

func (r SearchRequest) Search() (searchResp SearchResponse, err error) {
	b, _ := json.Marshal(r.RequestBody)
	body := bytes.NewBuffer(b)

	req, err := http.NewRequest(r.Method, r.Url, body)
	if err != nil {
		return searchResp, err
	}
	req.Header.Add("Authorization", "Bearer "+r.Token)

	httpClient := &http.Client{}
	rsp, err := httpClient.Do(req)
	if err != nil {
		log.Println("error with request")
		return searchResp, err
	}

	defer rsp.Body.Close()
	responseAsBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Println("could not read response body")
		return searchResp, err
	}

	err = json.Unmarshal(responseAsBytes, &searchResp)
	return searchResp, err
}
