package main

import (
	"errors"
	"strings"
)

type (
	// Plugin values.
	Plugin struct {
		BaseURL   string
		Username  string
		Token     string
		Job       []string
		Parameter []string
	}
)

func trimElement(keys []string) []string {
	newKeys := []string{}

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

	parameters := trimElement(p.Parameter)

	parameter := map[string]string(nil)

	if len(parameters) > 0 {
		parameter = make(map[string]string, len(parameters))
		for _, v := range parameters {
			keyAndValue := strings.SplitN(v, "=", 2)
			if len(keyAndValue) < 2 {
				return errors.New("please each jenkins-parameter as 'key'='value' string")
			}
			parameter[keyAndValue[0]] = keyAndValue[1]
		}
	}

	auth := &Auth{
		Username: p.Username,
		Token:    p.Token,
	}

	jenkins := NewJenkins(auth, p.BaseURL, parameter)

	for _, v := range jobs {
		if err := jenkins.trigger(v, nil); err != nil {
			return err
		}
	}

	return nil
}
