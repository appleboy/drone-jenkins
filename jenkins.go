package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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
		Client  *http.Client
	}
)

// NewJenkins is initial Jenkins object
func NewJenkins(auth *Auth, url string, insecure bool) *Jenkins {
	url = strings.TrimRight(url, "/")

	client := http.DefaultClient
	if insecure {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	return &Jenkins{
		Auth:    auth,
		BaseURL: url,
		Client:  client,
	}
}

func (jenkins *Jenkins) buildURL(path string, params url.Values) (requestURL string) {
	requestURL = jenkins.BaseURL + path
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
	return jenkins.Client.Do(req)
}

func (jenkins *Jenkins) parseResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	if body == nil {
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return json.Unmarshal(data, body)
}

func (jenkins *Jenkins) post(path string, params url.Values, body interface{}) (err error) {
	requestURL := jenkins.buildURL(path, params)
	// formData := params.Encode()
	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
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
	var urlPath string
	if params == nil || len(params) == 0 {
		urlPath = jenkins.parseJobPath(job) + "/build"
	} else {
		urlPath = jenkins.parseJobPath(job) + "/buildWithParameters"
	}

	return jenkins.post(urlPath, params, nil)
}
