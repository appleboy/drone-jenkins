package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testJobBuildPath    = "/job/test-job/build"
	testQueueItemPath   = "/queue/item/123/api/json"
	testBuildStatusPath = "/job/test-job/456/api/json"
)

// TestValidateConfig tests the validateConfig method
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		plugin    Plugin
		wantError bool
		errorMsg  string
	}{
		{
			name:      "missing all config",
			plugin:    Plugin{},
			wantError: true,
			errorMsg:  "jenkins base URL is required",
		},
		{
			name: "missing username and token",
			plugin: Plugin{
				BaseURL: "http://example.com",
			},
			wantError: true,
			errorMsg:  "jenkins username is required",
		},
		{
			name: "missing token",
			plugin: Plugin{
				BaseURL:  "http://example.com",
				Username: "foo",
			},
			wantError: true,
			errorMsg:  "jenkins API token is required",
		},
		{
			name: "all required config present",
			plugin: Plugin{
				BaseURL:  "http://example.com",
				Username: "foo",
				Token:    "bar",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.validateConfig()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTrimWhitespaceFromSlice tests the trimWhitespaceFromSlice function
func TestTrimWhitespaceFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "remove empty and whitespace strings",
			input:    []string{"1", "     ", "3"},
			expected: []string{"1", "3"},
		},
		{
			name:     "no whitespace strings",
			input:    []string{"1", "2"},
			expected: []string{"1", "2"},
		},
		{
			name:     "all whitespace",
			input:    []string{"   ", "\t", "\n"},
			expected: []string{},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "trim surrounding whitespace",
			input:    []string{"  foo  ", " bar ", "baz"},
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "mixed empty and valid",
			input:    []string{"", "valid", "", "also-valid", ""},
			expected: []string{"valid", "also-valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimWhitespaceFromSlice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseParameters tests the parseParameters function
func TestParseParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected url.Values
	}{
		{
			name:  "valid parameters",
			input: []string{"key1=value1", "key2=value2"},
			expected: url.Values{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
			},
		},
		{
			name:  "parameter with multiple equals signs",
			input: []string{"key=value=with=equals"},
			expected: url.Values{
				"key": []string{"value=with=equals"},
			},
		},
		{
			name:  "parameter with spaces in value",
			input: []string{"key=value with spaces"},
			expected: url.Values{
				"key": []string{"value with spaces"},
			},
		},
		{
			name:  "parameter with empty value",
			input: []string{"key="},
			expected: url.Values{
				"key": []string{""},
			},
		},
		{
			name:     "invalid parameter format (no equals)",
			input:    []string{"invalid"},
			expected: url.Values{},
		},
		{
			name:     "parameter with empty key",
			input:    []string{"=value"},
			expected: url.Values{},
		},
		{
			name:  "mixed valid and invalid",
			input: []string{"valid=yes", "invalid", "also=valid"},
			expected: url.Values{
				"valid": []string{"yes"},
				"also":  []string{"valid"},
			},
		},
		{
			name:  "key with surrounding whitespace",
			input: []string{"  key  =value"},
			expected: url.Values{
				"key": []string{"value"},
			},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: url.Values{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseParameters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExecMissingConfig tests Exec with missing configuration
func TestExecMissingConfig(t *testing.T) {
	var plugin Plugin

	err := plugin.Exec()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration error")
	assert.Contains(t, err.Error(), "jenkins base URL is required")
}

// TestExecMissingJenkinsUsername tests Exec with missing username
func TestExecMissingJenkinsUsername(t *testing.T) {
	plugin := Plugin{
		BaseURL: "http://example.com",
	}

	err := plugin.Exec()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration error")
	assert.Contains(t, err.Error(), "jenkins username is required")
}

// TestExecMissingJenkinsToken tests Exec with missing token
func TestExecMissingJenkinsToken(t *testing.T) {
	plugin := Plugin{
		BaseURL:  "http://example.com",
		Username: "foo",
	}

	err := plugin.Exec()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration error")
	assert.Contains(t, err.Error(), "jenkins API token is required")
}

// TestExecMissingJenkinsJob tests Exec with missing or empty job list
func TestExecMissingJenkinsJob(t *testing.T) {
	tests := []struct {
		name string
		jobs []string
	}{
		{
			name: "no jobs",
			jobs: []string{},
		},
		{
			name: "only whitespace jobs",
			jobs: []string{"   ", "\t", "\n"},
		},
		{
			name: "nil jobs",
			jobs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Plugin{
				BaseURL:  "http://example.com",
				Username: "foo",
				Token:    "bar",
				Job:      tt.jobs,
			}

			err := plugin.Exec()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "at least one Jenkins job name is required")
		})
	}
}

// TestExecTriggerBuild tests successful job triggering
func TestExecTriggerBuild(t *testing.T) {
	// Create a mock Jenkins server
	queueID := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().
			Set("Location", fmt.Sprintf("http://jenkins.example.com/queue/item/%d/", queueID))
		w.WriteHeader(http.StatusCreated)
		queueID++
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:  server.URL,
		Username: "foo",
		Token:    "bar",
		Job:      []string{"drone-jenkins"},
	}

	err := plugin.Exec()

	assert.NoError(t, err)
}

// TestExecTriggerMultipleJobs tests triggering multiple jobs
func TestExecTriggerMultipleJobs(t *testing.T) {
	// Create a mock Jenkins server
	jobsTriggered := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobsTriggered++
		w.Header().
			Set("Location", fmt.Sprintf("http://jenkins.example.com/queue/item/%d/", jobsTriggered))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:  server.URL,
		Username: "foo",
		Token:    "bar",
		Job:      []string{"job1", "job2", "job3"},
	}

	err := plugin.Exec()

	assert.NoError(t, err)
	assert.Equal(t, 3, jobsTriggered)
}

// TestExecWithParameters tests job triggering with parameters
func TestExecWithParameters(t *testing.T) {
	// Create a mock Jenkins server
	var receivedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.Header().Set("Location", "http://jenkins.example.com/queue/item/1/")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:    server.URL,
		Username:   "foo",
		Token:      "bar",
		Job:        []string{"parameterized-job"},
		Parameters: []string{"branch=main", "environment=production"},
	}

	err := plugin.Exec()

	assert.NoError(t, err)
	assert.Equal(t, "main", receivedQuery.Get("branch"))
	assert.Equal(t, "production", receivedQuery.Get("environment"))
}

// TestExecWithRemoteToken tests job triggering with remote token
func TestExecWithRemoteToken(t *testing.T) {
	// Create a mock Jenkins server
	var receivedToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedToken = r.URL.Query().Get("token")
		w.Header().Set("Location", "http://jenkins.example.com/queue/item/1/")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:     server.URL,
		Username:    "foo",
		Token:       "bar",
		RemoteToken: "remote-token-123",
		Job:         []string{"secure-job"},
	}

	err := plugin.Exec()

	assert.NoError(t, err)
	assert.Equal(t, "remote-token-123", receivedToken)
}

// TestExecWithJobsContainingWhitespace tests job list with whitespace
func TestExecWithJobsContainingWhitespace(t *testing.T) {
	// Create a mock Jenkins server
	jobsTriggered := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobsTriggered++
		w.Header().
			Set("Location", fmt.Sprintf("http://jenkins.example.com/queue/item/%d/", jobsTriggered))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:  server.URL,
		Username: "foo",
		Token:    "bar",
		Job:      []string{"  job1  ", "job2", "   ", "job3"},
	}

	err := plugin.Exec()

	assert.NoError(t, err)
	// Should trigger 3 jobs (whitespace-only entry should be filtered out)
	assert.Equal(t, 3, jobsTriggered)
}

// TestExecWithWaitSuccess tests job execution with wait for successful completion
func TestExecWithWaitSuccess(t *testing.T) {
	// Create a mock Jenkins server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case testJobBuildPath:
			// Trigger build
			w.Header().Set("Location", "http://jenkins.example.com/queue/item/123/")
			w.WriteHeader(http.StatusCreated)
		case testQueueItemPath:
			// Queue item with build number
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":123,"executable":{"number":456}}`))
		case testBuildStatusPath:
			// Build completed successfully
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"number":456,"building":false,"result":"SUCCESS"}`))
		}
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:  server.URL,
		Username: "foo",
		Token:    "bar",
		Job:      []string{"test-job"},
		Wait:     true,
	}

	err := plugin.Exec()

	assert.NoError(t, err)
}

// TestExecWithWaitFailure tests job execution with wait for failed build
func TestExecWithWaitFailure(t *testing.T) {
	// Create a mock Jenkins server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case testJobBuildPath:
			// Trigger build
			w.Header().Set("Location", "http://jenkins.example.com/queue/item/123/")
			w.WriteHeader(http.StatusCreated)
		case testQueueItemPath:
			// Queue item with build number
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":123,"executable":{"number":456}}`))
		case testBuildStatusPath:
			// Build completed with failure
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"number":456,"building":false,"result":"FAILURE"}`))
		}
	}))
	defer server.Close()

	plugin := Plugin{
		BaseURL:  server.URL,
		Username: "foo",
		Token:    "bar",
		Job:      []string{"test-job"},
		Wait:     true,
	}

	err := plugin.Exec()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed with status: FAILURE")
}
