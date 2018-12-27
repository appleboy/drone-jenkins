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
		Username: "admin",
		Token:    "1148e37e0d2a18296cd25e4e9ebdbfa3cc",
	}
	jenkins := NewJenkins(auth, "example.com")

	err := jenkins.trigger("drone-jenkins", nil)
	assert.NotNil(t, err)
}

func TestTriggerBuild(t *testing.T) {
	auth := &Auth{
		Username: "admin",
		Token:    "1148e37e0d2a18296cd25e4e9ebdbfa3cc",
	}
	jenkins := NewJenkins(auth, "http://localhost:8080")

	err := jenkins.trigger("drone-jenkins", url.Values{"token": []string{"1148e37e0d2a18296cd25e4e9ebdbfa3cc"}})
	assert.Nil(t, err)
}
