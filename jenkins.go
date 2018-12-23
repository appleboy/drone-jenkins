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

	// XSRF contain header name and value
	////"xpath=at(//crumbRequestField,\":\",//crumb)"
	Crumb struct {
		crumbRequestField string
		crumb string
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

func (jenkins *Jenkins) loadXSRFtoken(body interface{}) (err error) {
	// call the json endpoint of jenkins API for the XSRF Token
	requestURL := jenkins.buildURL("crumbIssuer/api/json", nil)
	

	req, err := http.NewRequest("GET", requestURL, nil)
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

func (jenkins *Jenkins) post(path string, params url.Values, body interface{}, Crumb jenkinsCrumb) (err error) {
	requestURL := jenkins.buildURL(path, params)
	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if exists add the XSRF token as header to the POST request
	if jenkinsCrumb != nil {
		req.Header.Set(jenkinsCrumb.crumbRequestField, jenkinsCrumb.crumb)
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

	// load XSRF token the the following trigger request
	var animals []Crumb
	err = jenkins.loadXSRFtoken(&animals)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animals)

	//set token object to the jenkins as child
	var jenkinsCrumb Crumb

	fmt.Printf("send a request to %q\n", path)

	return jenkins.post(path, params, nil, jenkinsCrumb)
}
