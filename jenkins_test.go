package main

import (
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
	auth := &Auth{
		Username: "foo",
		Token:    "bar",
	}
	jenkins := NewJenkins(auth, "http://example.com", "remote-token", false)

	err := jenkins.trigger("drone-jenkins", url.Values{"param": []string{"value"}})
	assert.Nil(t, err)
}
