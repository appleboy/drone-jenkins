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
		BaseURL: "http://example.com",
	}

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestPluginTriggerBuild(t *testing.T) {
	plugin := Plugin{
		BaseURL:  "http://example.com",
		Username: "foo",
		Token:    "bar",
		Job:      []string{"drone-jenkins"},
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
