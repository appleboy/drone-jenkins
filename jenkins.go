package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
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
		Data    map[string]string
		client  http.Client
	}

	// Crumb contain the jenkins XSRF token header-name and header-value
	Crumb struct {
		Class             string `json:"_class"`
		Crumb             string `json:"crumb"`
		CrumbRequestField string `json:"crumbRequestField"`
	}
)

// NewJenkins is initial Jenkins object
func NewJenkins(auth *Auth, url string, data map[string]string) *Jenkins {
	url = strings.TrimRight(url, "/")
	var j = Jenkins{
		Auth:    auth,
		BaseURL: url,
		Data:    data,
	}
	j.initClient()
	return &j
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

func (jenkins *Jenkins) initClient() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	jenkins.client = http.Client{
		Jar: jar,
	}
}

func (jenkins *Jenkins) sendRequest(req *http.Request) (*http.Response, error) {
	if jenkins.Auth != nil {
		req.SetBasicAuth(jenkins.Auth.Username, jenkins.Auth.Token)
	}

	resp, err := jenkins.client.Do(req)

	if err != nil {
		log.Fatal(err)
	} else {
		// check for response error 401
		if resp.StatusCode == 401 {
			return resp, errors.New("HTTP 401 - invalid password/token for user")
		}

		for _, cookie := range resp.Cookies() {
			fmt.Printf("trace - parseResponse - SET COOKIE: %s \n", cookie.Name)
		}

	}

	return resp, err
}

func (jenkins *Jenkins) parseResponse(resp *http.Response, body interface{}) (err error) {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// for debug if you would like to show the raw json data
	fmt.Printf("trace - parseResponse - raw data: %s \n", data)

	if body == nil {
		return
	}

	return json.Unmarshal(data, body)
}

func (jenkins *Jenkins) loadXSRFtoken(body interface{}) (err error) {
	// call the json endpoint of jenkins API for the XSRF Token
	requestURL := jenkins.buildURL("/crumbIssuer/api/json", nil)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return
	}

	if resp.StatusCode == 404 {
		log.Print("info - loadXSRFtoken - XSRF is not enabled in remote jenkins")
		return
	}

	jsonError := jenkins.parseResponse(resp, body)
	if jsonError != nil {
		log.Fatal(err)
	} else {
		log.Print("trace - loadXSRFtoken - convert into jenkinsCrumb: ", body)
	}

	return jsonError
}

func (jenkins *Jenkins) post(path string, queryParams url.Values, body interface{}, jenkinsCrumb *Crumb, postData io.Reader) (err error) {
	requestURL := jenkins.buildURL(path, queryParams)

	req, err := http.NewRequest("POST", requestURL, postData)
	if err != nil {
		return
	}

	if postData != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// if exists add the XSRF token as header to the POST request
	if jenkinsCrumb != nil {
		if len(jenkinsCrumb.Class) > 0 {
			log.Print("trace - add an XSRF-Token-Header to a POST request")
			req.Header.Set(jenkinsCrumb.CrumbRequestField, jenkinsCrumb.Crumb)
		}
	}

	fmt.Printf("trace - send a POST request to %q \n", path)

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return
	}

	//TODO check response code if ok

	// POST return the Location to the new Job
	// https://github.com/jenkinsci/parameterized-remote-trigger-plugin
	// https://wiki.jenkins.io/display/JENKINS/Parameterized+Remote+Trigger+Plugin
	locationHeader := resp.Header.Get("Location")
	fmt.Println("HEADER Location:", locationHeader)
	// it is a link to the queue, executable.url link to the job build

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

func (jenkins *Jenkins) trigger(job string, queryParams url.Values) error {
	path := jenkins.parseJobPath(job) + "/build"
	fmt.Printf("info - trigger - set api job path to %q\n", path)

	jenkinsCrumb := Crumb{}

	// load XSRF token for the following POST request
	jenkins.loadXSRFtoken(&jenkinsCrumb)

	var postData io.Reader
	if jenkins.Data != nil {
		// convert the map into the jenkins json DTO and make url encoding too
		var json string
		for name, value := range jenkins.Data {
			json += fmt.Sprintf("{\"name\":\"%s\",\"value\":\"%s\"}", name, value) + ","
		}
		json = strings.TrimRight(json, ",")

		data := url.Values{}
		data.Set("json", fmt.Sprintf("{\"parameter\": [%s]}", json))
		postData = strings.NewReader(data.Encode())
	} else {
		postData = nil
	}

	return jenkins.post(path, queryParams, nil, &jenkinsCrumb, postData)
}
