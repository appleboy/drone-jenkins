package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingConfig(t *testing.T) {
	var plugin Plugin

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestMissingJenkinsConfig(t *testing.T) {
	plugin := Plugin{
		BaseURL: "http://jenkins:8080",
	}

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestMissingJenkinsJob(t *testing.T) {
	plugin := Plugin{
		BaseURL:  "http://jenkins:8080",
		Username: "dev",
		Token:    "devdev",
	}

	err := plugin.Exec()
	assert.NotNil(t, err)

	plugin.Job = []string{"   "}

	err = plugin.Exec()
	assert.NotNil(t, err)
}

func TestPluginTriggerBuild(t *testing.T) {
	plugin := Plugin{
		BaseURL:  "http://jenkins:8080",
		Username: "dev",
		Token:    "devdev",
		Job:      []string{"first-pipeline"},
	}

	err := plugin.Exec()

	assert.Nil(t, err)
}

func TestPluginTriggerBuild2(t *testing.T) {
	plugin := Plugin{
		BaseURL:   "http://jenkins:8080",
		Username:  "dev",
		Token:     "devdev",
		Job:       []string{"another-pipeline"},
		Parameter: []string{"sValue=abc", "sValue2:xyz"},
	}

	err := plugin.Exec()

	assert.NotNil(t, err)
}

func TestPluginTriggerBuild3(t *testing.T) {
	plugin := Plugin{
		BaseURL:   "http://jenkins:8080",
		Username:  "dev",
		Token:     "devdev",
		Job:       []string{"another-pipeline"},
		Parameter: []string{"sValue=abc", "sValue2=xyz"},
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
