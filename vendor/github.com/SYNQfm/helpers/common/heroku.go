package common

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const HEROKU_BASE_URL = "https://api.heroku.com/"

func UpdateHerokuVar(authKey, appName string, config interface{}) error {
	herokuUrl := HEROKU_BASE_URL + "apps/" + appName + "/config-vars"

	data, _ := json.Marshal(config)
	body := bytes.NewBuffer(data)

	req, err := makeRequest("PATCH", herokuUrl, authKey, body)
	if err != nil {
		return err
	}

	err = handleRequest(req, config)
	if err != nil {
		return err
	}

	return nil
}

func makeRequest(method, url, token string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		log.Println("could not create request: ", err.Error())
		return req, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.heroku+json; version=3")

	return req, nil
}

func handleRequest(req *http.Request, f interface{}) error {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("could not make http request: ", err.Error())
		return err
	}
	return parseResponse(resp, f)
}

func parseResponse(resp *http.Response, f interface{}) error {
	defer resp.Body.Close()
	responseAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("could not read resp body", err.Error())
		return err
	}

	err = json.Unmarshal(responseAsBytes, &f)
	if err != nil {
		log.Println("could not parse response: ", err.Error())
		return NewError("could not parse : %s", string(responseAsBytes))
	}

	return nil
}
