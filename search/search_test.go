package search

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SYNQfm/SYNQ-Golang/synq"
	"github.com/SYNQfm/SYNQ-Golang/test_server"
	"github.com/stretchr/testify/require"
)

var testServer *httptest.Server

func handle(w http.ResponseWriter, r *http.Request) {
	asset := synq.Asset{}
	resp := SearchResponse{1, 1, 1, 100, []string{test_server.V2_VIDEO_ID}, []synq.Asset{asset}}
	rbytes, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(rbytes)
}

func init() {
	testServer = httptest.NewServer(http.HandlerFunc(handle))
}

func createRequestBody() SearchRequestBody {
	return SearchRequestBody{
		Query: "",
		Params: map[string]interface{}{
			"filters": map[string]string{
				"video_id": "",
				"type":     "",
			},
			"custom_filters": "upload_info.checksum: 'acda50d6da2d2784fea378b05d6b66d3'",
			"hits_per_page":  100,
			"page":           1,
		},
		Options: SearchOptions{false, true, false},
	}
}

func TestSearch(t *testing.T) {
	log.Println("Testing Search")
	assert := require.New(t)

	reqBody := createRequestBody()
	request := SearchRequest{"POST", testServer.URL, test_server.TEST_AUTH, &reqBody}
	resp, err := request.Search()
	assert.Nil(err)
	assert.Equal(1, resp.Page)
	assert.Equal(1, resp.NbHits)
	assert.Equal(1, resp.NbPage)
	assert.Equal(100, resp.HitsPerPage)
	assert.Equal(test_server.V2_VIDEO_ID, resp.IdList[0])
}
