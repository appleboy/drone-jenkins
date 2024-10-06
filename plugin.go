package main

import (
	"errors"
	"log"
	"net/url"
	"strings"
)

type (
	// Plugin values.
	Plugin struct {
		BaseURL    string
		Username   string
		Token      string
		Job        []string
		Insecure   bool
		Parameters []string
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

	auth := &Auth{
		Username: p.Username,
		Token:    p.Token,
	}

	jenkins := NewJenkins(auth, p.BaseURL, p.Insecure)

	params := url.Values{}
	for _, v := range p.Parameters {
		kv := strings.Split(v, "=")
		if len(kv) == 2 {
			params.Add(kv[0], kv[1])
		}
	}

	for _, v := range jobs {
		if err := jenkins.trigger(v, params); err != nil {
			return err
		}
		log.Printf("trigger job %s success", v)
	}

	return nil
}
