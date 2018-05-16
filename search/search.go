package search

import (
	"bytes"
	"encoding/json"
	"net/http"
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
}

type SearchResponse struct {
	Page        int      `json:"page"`
	NbHits      int      `json:"nbHits"`
	NbPage      int      `json:"nbPages"`
	HitsPerPage int      `json:"hitsPerPage`
	IdList      []string `json:"idList"`
	Hits        []Asset  `json:"hits"`
}

func (r SearchRequest) Search() (resp *http.Response, err error) {
	b, _ := json.Marshal(r.RequestBody)
	body := bytes.NewBuffer(b)

	req, err := http.NewRequest(r.Method, r.Url, body)
	if err != nil {
		return resp, err
	}
	req.Header.Add("Authorization", "Bearer "+r.Token)

	httpClient := &http.Client{}
	resp, err = httpClient.Do(req)
	return resp, err
}
