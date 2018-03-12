package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	baseUrl = "https://api.keen.io/3.0/projects/"
)

type Client struct {
	WriteKey   string
	ReadKey    string
	ProjectId  string
	Collection string
	HttpClient http.Client
	Events     map[string][]KeenEvent
}

type KeenEvent struct {
	Keen        KeenProperties `json:"keen"`
	CDNEvent    CDNEvent       `json:"cdn_event"`
	GeoLocation GeoLocation    `json:"geo"`
	Browser     Browser        `json:"browser"`
}

type KeenProperties struct {
	Timestamp string `json:"timestamp"`
}

type CDNEvent struct {
	CdnName       string `json:"cdn"`
	CreatedAt     string `json:"created_at"`
	HTTPMethod    string `json:"method"`
	ServerIP      string `json:"server_ip"`
	Protocol      string `json:"protocol"`
	Referrer      string `json:"referrer"`
	Filesize      int64  `json:"filesize"`
	BytesRequest  int64  `json:"bytes_request"`
	BytesResponse int64  `json:"bytes_response"`
	Duration      string `json:"duration"`
	StatusCode    string `json:"status"`
	URI           string `json:"url"`
	ProjectId     string `json:"project_uuid"`
	VideoId       string `json:"video_uuid"`
}

type GeoLocation struct {
	GeoInfo   GeoInfo `json:"info"`
	IpAddress string  `json:"ip_address"`
}

type GeoInfo struct {
	City        string    `json:"city"`
	State       string    `json:"state"`
	Country     string    `json:"country"`
	Continent   string    `json:"continent"`
	Coordinates [2]string `json:"coordinates"` // [longitude, latitude]
}

type Browser struct {
	UserAgent string `json:"useragent"`
}

// Timestamp formats a time.Time object in the ISO-8601 format keen expects
func Timestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func (c *Client) AddEvent(event KeenEvent) (string, error) {
	resp, err := c.Request("POST", fmt.Sprintf("/events/%s", c.Collection), event)
	if err != nil {
		return "", err
	}

	return c.Response(resp)
}

func (c *Client) AddEvents(events map[string][]KeenEvent) (string, error) {
	resp, err := c.Request("POST", "/events", events)
	if err != nil {
		return "", err
	}

	return c.Response(resp)
}

func (c *Client) Response(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	response := string(data)

	if resp.StatusCode != 200 {
		response = strconv.Itoa(resp.StatusCode) + " " + resp.Status
	}

	return response, nil
}

func (c *Client) Request(method, path string, payload interface{}) (*http.Response, error) {
	// serialize payload
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// construct url
	url := baseUrl + c.ProjectId + path

	// new request
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// add auth if POST
	// key is added in query string for GET
	if method == "POST" {
		req.Header.Add("Authorization", c.WriteKey)
	}

	// set length/content-type
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
		req.ContentLength = int64(len(body))
	}

	return c.HttpClient.Do(req)
}
