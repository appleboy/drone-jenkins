package main

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestMissingConfig(t *testing.T) {
	var plugin Plugin

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestMissingJenkinsConfig(t *testing.T) {
	plugin := Plugin{
		Config: Config{
			BaseURL: "http://example.com",
		},
	}

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestPluginTriggerBuild(t *testing.T) {
	plugin := Plugin{
		Repo: Repo{
			Name:  "go-hello",
			Owner: "appleboy",
		},
		Build: Build{
			Number:  101,
			Status:  "success",
			Link:    "https://github.com/appleboy/go-hello",
			Author:  "Bo-Yi Wu",
			Branch:  "master",
			Message: "update by drone line plugin.",
			Commit:  "e7c4f0a63ceeb42a39ac7806f7b51f3f0d204fd2",
		},

		Config: Config{
			BaseURL:  "http://example.com",
			Username: "foo",
			Token:    "bar",
			Job:      []string{"drone-jenkins"},
		},
	}

	err := plugin.Exec()

	assert.Nil(t, err)
}

func TestTrimElement(t *testing.T) {
	var input, result []string

	input = []string{"1", "     ", "3"}
	result = []string{"1", "3"}

	assert.Equal(t, result, trimElement(input))

	input = []string{"1", "2"}
	result = []string{"1", "2"}

	assert.Equal(t, result, trimElement(input))
}
