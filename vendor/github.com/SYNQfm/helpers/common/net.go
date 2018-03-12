package common

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func DownloadChunk(url string, size int64) (body []byte, err error) {
	var client http.Client
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Range", fmt.Sprintf("bytes=0-%d", size))
	resp, _ := client.Do(req)
	if resp.StatusCode != 200 {
		return body, NewError("could not retrieve url : %d", resp.StatusCode)
	}
	body, _ = ioutil.ReadAll(resp.Body)
	return body, nil
}

func ChunkAndHash(url string, size int64) string {
	body, err := DownloadChunk(url, size)
	if err != nil {
		errors.New("error downloading chunk : " + err.Error())
	}
	m := md5.New()
	return fmt.Sprintf("%x", m.Sum(body))
}
