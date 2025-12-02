package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

type (
	// Plugin represents the configuration for the Jenkins plugin.
	// It contains all necessary credentials and settings to trigger Jenkins jobs.
	Plugin struct {
		BaseURL      string        // Jenkins server base URL
		Username     string        // Jenkins username for authentication
		Token        string        // Jenkins API token for authentication
		RemoteToken  string        // Optional remote trigger token for additional security
		Job          []string      // List of Jenkins job names to trigger
		Insecure     bool          // Whether to skip TLS certificate verification
		Parameters   []string      // Job parameters in key=value format
		Wait         bool          // Whether to wait for job completion
		PollInterval time.Duration // Interval between status checks (default: 10s)
		Timeout      time.Duration // Maximum time to wait for job completion (default: 30m)
		Debug        bool          // Enable debug mode to show detailed parameter information
	}
)

// trimWhitespaceFromSlice removes empty and whitespace-only strings from a slice.
// It returns a new slice containing only non-empty trimmed strings.
func trimWhitespaceFromSlice(items []string) []string {
	result := make([]string, 0, len(items))

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// parseParameters converts a slice of key=value strings into url.Values.
// It logs a warning for any parameters that don't match the expected format.
func parseParameters(params []string) url.Values {
	values := url.Values{}

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			log.Printf("warning: skipping invalid parameter format (expected key=value): %q", param)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := parts[1] // Keep value as-is to preserve intentional spaces

		if key == "" {
			log.Printf("warning: skipping parameter with empty key: %q", param)
			continue
		}

		values.Add(key, value)
	}

	return values
}

// validateConfig checks that all required plugin configuration is present.
// It returns a descriptive error if any required field is missing.
func (p Plugin) validateConfig() error {
	if p.BaseURL == "" {
		return errors.New("jenkins base URL is required")
	}
	if p.Username == "" {
		return errors.New("jenkins username is required")
	}
	if p.Token == "" {
		return errors.New("jenkins API token is required")
	}
	return nil
}

// Exec executes the plugin by triggering the configured Jenkins jobs.
// It validates the configuration, parses parameters, and triggers each job sequentially.
// Returns an error if validation fails or any job trigger fails.
func (p Plugin) Exec() error {
	// Validate required configuration
	if err := p.validateConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Clean and validate job list
	jobs := trimWhitespaceFromSlice(p.Job)
	if len(jobs) == 0 {
		return errors.New("at least one Jenkins job name is required")
	}

	// Set up authentication
	auth := &Auth{
		Username: p.Username,
		Token:    p.Token,
	}

	// Initialize Jenkins client
	jenkins := NewJenkins(auth, p.BaseURL, p.RemoteToken, p.Insecure, p.Debug)

	// Parse job parameters
	params := parseParameters(p.Parameters)

	// Set default values for wait configuration
	pollInterval := p.PollInterval
	if pollInterval == 0 {
		pollInterval = 10 * time.Second
	}

	timeout := p.Timeout
	if timeout == 0 {
		timeout = 30 * time.Minute
	}

	// Trigger each job
	for _, jobName := range jobs {
		queueID, err := jenkins.trigger(jobName, params)
		if err != nil {
			return fmt.Errorf("failed to trigger job %q: %w", jobName, err)
		}
		log.Printf("successfully triggered job: %s (queue #%d)", jobName, queueID)

		// Wait for job completion if requested
		if p.Wait {
			buildInfo, err := jenkins.waitForCompletion(jobName, queueID, pollInterval, timeout)
			if err != nil {
				return fmt.Errorf("error waiting for job %q: %w", jobName, err)
			}

			// Check if build was successful
			if buildInfo.Result != "SUCCESS" {
				return fmt.Errorf(
					"job %q (build #%d) failed with status: %s",
					jobName,
					buildInfo.Number,
					buildInfo.Result,
				)
			}

			log.Printf("job %s (build #%d) completed successfully", jobName, buildInfo.Number)
		}
	}

	return nil
}
