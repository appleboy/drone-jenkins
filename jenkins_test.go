package main

import (
	"github.com/stretchr/testify/assert"

	"net/url"
	"testing"
)

func TestParseJobPath(t *testing.T) {
	auth := &Auth{
		Username: "appleboy",
		Token:    "1234",
	}
	jenkins := NewJenkins(auth, "http://example.com")

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
	jenkins := NewJenkins(auth, "example.com")

	err := jenkins.trigger("drone-jenkins", nil)
	assert.NotNil(t, err)
}

func TestTriggerBuild(t *testing.T) {
	auth := &Auth{
		Username: "foo",
		Token:    "bar",
	}
	jenkins := NewJenkins(auth, "http://example.com")

	err := jenkins.trigger("drone-jenkins", url.Values{"token": []string{"bar"}})
	assert.Nil(t, err)
}
