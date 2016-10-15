package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type (
	// Auth contain username and token
	Auth struct {
		Username string
		Token    string
	}

	// Jenkins contain Auth and BaseURL
	Jenkins struct {
		Auth    *Auth
		BaseURL string
	}
)

// NewJenkins is initial Jenkins object
func NewJenkins(auth *Auth, url string) *Jenkins {
	url = strings.TrimRight(url, "/")
	return &Jenkins{
		Auth:    auth,
		BaseURL: url,
	}
}

func (jenkins *Jenkins) buildURL(path string, params url.Values) (requestURL string) {
	requestURL = jenkins.BaseURL + path
	fmt.Println(requestURL)
	if params != nil {
		queryString := params.Encode()
		if queryString != "" {
			requestURL = requestURL + "?" + queryString
		}
	}

	return
}

func (jenkins *Jenkins) sendRequest(req *http.Request) (*http.Response, error) {
	if jenkins.Auth != nil {
		req.SetBasicAuth(jenkins.Auth.Username, jenkins.Auth.Token)
	}
	return http.DefaultClient.Do(req)
}

func (jenkins *Jenkins) parseResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	if body == nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return json.Unmarshal(data, body)
}

func (jenkins *Jenkins) post(path string, params url.Values, body interface{}) (err error) {
	requestURL := jenkins.buildURL(path, params)
	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	return jenkins.parseResponse(resp, body)
}

func (jenkins *Jenkins) parseJobPath(job string) string {
	var path string

	jobs := strings.Split(strings.TrimPrefix(job, "/"), "/")

	for _, value := range jobs {
		value = strings.Trim(value, " ")
		if len(value) == 0 {
			continue
		}

		path = fmt.Sprintf("%s/job/%s", path, value)
	}

	return path
}

func (jenkins *Jenkins) trigger(job string, params url.Values) error {
	path := jenkins.parseJobPath(job) + "/build"

	return jenkins.post(path, params, nil)
}
