package main

import (
	"errors"
	"strings"
)

type (
	// Plugin values.
	Plugin struct {
		BaseURL  string
		Username string
		Token    string
		Job      []string
	}
)

func trimElement(keys []string) []string {
	var newKeys []string

	for _, value := range keys {
		value = strings.Trim(value, " ")
		if len(value) == 0 {
			continue
		}
		newKeys = append(newKeys, value)
	}

	return newKeys
}

// Exec executes the plugin.
func (p Plugin) Exec() error {

	if len(p.BaseURL) == 0 || len(p.Username) == 0 || len(p.Token) == 0 {
		return errors.New("missing jenkins config")
	}

	jobs := trimElement(p.Job)

	if len(jobs) == 0 {
		return errors.New("missing jenkins job")
	}

	auth := &Auth{
		Username: p.Username,
		Token:    p.Token,
	}

	jenkins := NewJenkins(auth, p.BaseURL)

	for _, value := range jobs {
		jenkins.trigger(value, nil)
	}

	return nil
}
