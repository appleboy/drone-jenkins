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

	// Crumb contain the jenkins XSRF token header-name and header-value
	Crumb struct {
		Class             string `json:"_class"`
		Crumb             string `json:"crumb"`
		CrumbRequestField string `json:"crumbRequestField"`
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

func (jenkins *Jenkins) buildURL(path string, queryParams url.Values) (requestURL string) {
	requestURL = jenkins.BaseURL + path
	if queryParams != nil {
		queryString := queryParams.Encode()
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

	// for debug if you would like to show the raw json data
	fmt.Printf("trace - parseResponse - raw data: %s \n", data)

	return json.Unmarshal(data, body)
}

func (jenkins *Jenkins) loadXSRFtoken(body interface{}) (err error) {
	// call the json endpoint of jenkins API for the XSRF Token
	requestURL := jenkins.buildURL("/crumbIssuer/api/json", nil)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("trace - loadXSRFtoken - error by create NewRequest:", err)
		return
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		fmt.Println("trace - loadXSRFtoken - error by sendRequest:", err)
		return
	}

	return jenkins.parseResponse(resp, body)
}

func (jenkins *Jenkins) post(path string, queryParams url.Values, body interface{}, jenkinsCrumb *Crumb, postData *strings.Reader) (err error) {
	requestURL := jenkins.buildURL(path, queryParams)

	req, err := http.NewRequest("POST", requestURL, postData)
	if err != nil {
		fmt.Println("warn - post - error by create NewRequest:", err)
		return
	}

	if postData != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// if exists add the XSRF token as header to the POST request
	if jenkinsCrumb != nil {
		if len(jenkinsCrumb.Class) > 0 {
			fmt.Printf("info - add an XSRF token header to a POST request\n")
			req.Header.Set(jenkinsCrumb.CrumbRequestField, jenkinsCrumb.Crumb)
		}
	}

	fmt.Printf("info - send a POST request to %q\n", path)
	resp, err := jenkins.sendRequest(req)
	if err != nil {
		fmt.Println("warn - post - error by sendRequest:", err)
		return
	}

	// POST return the Location to the new Job
	// https://github.com/jenkinsci/parameterized-remote-trigger-plugin
	// https://wiki.jenkins.io/display/JENKINS/Parameterized+Remote+Trigger+Plugin
	var locationHeader string = resp.Header.Get("Location")
	fmt.Println("HEADER Location:", locationHeader)
	// it is a link to the queue, executable.url link to the job build

	return jenkins.parseResponse(resp, body
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

func (jenkins *Jenkins) trigger(job string, queryParams url.Values) error {
	path := jenkins.parseJobPath(job) + "/build"
	fmt.Printf("info - set api job path to %q\n", path)

	// load XSRF token for the following POST request
	jenkinsCrumb := Crumb{}
	err := jenkins.loadXSRFtoken(&jenkinsCrumb)
	if err != nil {
		fmt.Println("info - could not load an XSRF token:", err)
	}
	fmt.Printf("trace - load jenkinsCrumb: %+v \n", jenkinsCrumb)

	// im demo jenkins muss noch der job dazu ins git
	// das muss noch schöner werden hier - aber der wert kam erstmal im jenkins an - das ist gut
	data := url.Values{}
	data.Set("json", "{\"parameter\": [{\"name\":\"state_value\", \"value\":\"leck mich doch\"}]}")
	postData := strings.NewReader(data.Encode())

	return jenkins.post(path, queryParams, nil, &jenkinsCrumb, postData)
}
