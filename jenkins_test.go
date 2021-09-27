package main

import (
	"fmt"

	"github.com/stretchr/testify/assert"

	"net/url"
	"testing"
)

func TestParseJobPath(t *testing.T) {
	auth := &Auth{
		Username: "appleboy",
		Token:    "1234",
	}
	jenkins := NewJenkins(auth, "http://example.com", nil)

	assert.Equal(t, "/job/foo", jenkins.parseJobPath("/foo/"))
	assert.Equal(t, "/job/foo", jenkins.parseJobPath("foo/"))
	assert.Equal(t, "/job/foo/job/bar", jenkins.parseJobPath("foo/bar"))
	assert.Equal(t, "/job/foo/job/bar", jenkins.parseJobPath("foo///bar"))
}

func TestBuildURL(t *testing.T) {
	auth := &Auth{
		Username: "appleboy",
		Token:    "1234",
	}
	jenkins := NewJenkins(auth, "http://example.com", nil)

	assert.Equal(t, "http://example.com/foo/", jenkins.buildURL("/foo/", nil))

	query := url.Values{"ba": []string{"ta"}}
	assert.Equal(t, "http://example.com/foo/?ba=ta", jenkins.buildURL("/foo/", query))
}

func TestUnSupportProtocol(t *testing.T) {
	auth := &Auth{
		Username: "admin",
		Token:    "116ba1f8950c5ddff3ab26d6f5acbc3e41",
	}
	jenkins := NewJenkins(auth, "example.com", nil)

	err := jenkins.trigger("drone-jenkins", nil)
	assert.NotNil(t, err)
}

func TestLoadXSRFToken(t *testing.T) {
	auth := &Auth{
		Username: "admin",
		Token:    "adminadmin",
	}
	jenkins := NewJenkins(auth, "http://jenkins:8080", nil)

	// load XSRF token for the following POST request
	jenkinsCrumb := Crumb{}
	err := jenkins.loadXSRFtoken(&jenkinsCrumb)
	if err != nil {
		fmt.Println("warn - error by load XSRF token:", err)
	}
	fmt.Printf("info - load jenkinsCrumb: %+v \n", jenkinsCrumb)

	assert.Nil(t, err)
}

func TestTriggerBuild1(t *testing.T) {
	auth := &Auth{
		Username: "admin",
		Token:    "adminadmin",
	}
	jenkins := NewJenkins(auth, "http://jenkins:8080", nil)

	err := jenkins.trigger("first-pipeline", url.Values{"token": []string{"117caafd2840748c41157c445762d07624"}})
	assert.Nil(t, err)
}

func TestTriggerBuild2(t *testing.T) {
	auth := &Auth{
		Username: "admin",
		Token:    "adminadmin",
	}

	//var m map[string]string
	var m = make(map[string]string)
	m["sValue"] = "du da"
	m["sValue2"] = "noch was"

	jenkins := NewJenkins(auth, "http://jenkins:8080", m)

	err := jenkins.trigger("another-pipeline", url.Values{"token": []string{"117caafd2840748c41157c445762d07624"}})
	assert.Nil(t, err)
}
