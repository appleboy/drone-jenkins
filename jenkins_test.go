package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJobPath(t *testing.T) {
	auth := &Auth{
		Username: "appleboy",
		Token:    "1234",
	}
	jenkins := NewJenkins(auth, "http://example.com", "", false)

	assert.Equal(t, "/job/foo", jenkins.parseJobPath("/foo/"))
	assert.Equal(t, "/job/foo", jenkins.parseJobPath("foo/"))
	assert.Equal(t, "/job/foo/job/bar", jenkins.parseJobPath("foo/bar"))
	assert.Equal(t, "/job/foo/job/bar", jenkins.parseJobPath("foo///bar"))
}

func TestUnSupportProtocol(t *testing.T) {
	auth := &Auth{
		Username: "foo",
		Token:    "bar",
	}
	jenkins := NewJenkins(auth, "example.com", "", false)

	err := jenkins.trigger("drone-jenkins", nil)
	assert.NotNil(t, err)
}

func TestTriggerBuild(t *testing.T) {
	// Create a mock Jenkins server
	var receivedParams url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedParams = r.URL.Query()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	auth := &Auth{
		Username: "foo",
		Token:    "bar",
	}
	jenkins := NewJenkins(auth, server.URL, "remote-token", false)

	params := url.Values{"param": []string{"value"}}
	err := jenkins.trigger("drone-jenkins", params)

	assert.NoError(t, err)
	assert.Equal(t, "value", receivedParams.Get("param"))
	assert.Equal(t, "remote-token", receivedParams.Get("token"))
}
