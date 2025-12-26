package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/appleboy/com/gh"
	"github.com/yassinebenaid/godump"
)

type (
	// Auth contain username and token
	Auth struct {
		Username string
		Token    string
	}

	// Jenkins contain Auth and BaseURL
	Jenkins struct {
		Auth    *Auth
		BaseURL string
		Token   string // Remote trigger token
		Client  *http.Client
		Debug   bool // Enable debug mode to show detailed information
	}

	// QueueItem represents a Jenkins queue item response
	QueueItem struct {
		Blocked      bool  `json:"blocked"`
		Buildable    bool  `json:"buildable"`
		ID           int   `json:"id"`
		InQueueSince int64 `json:"inQueueSince"`
		Executable   *struct {
			Number int    `json:"number"`
			URL    string `json:"url"`
		} `json:"executable"`
		Why string `json:"why"`
	}

	// BuildInfo represents Jenkins build information
	BuildInfo struct {
		Building  bool   `json:"building"`
		Duration  int64  `json:"duration"`
		Result    string `json:"result"` // SUCCESS, FAILURE, ABORTED, UNSTABLE, null if building
		Number    int    `json:"number"`
		URL       string `json:"url"`
		Timestamp int64  `json:"timestamp"`
	}
)

// loadCACert loads a CA certificate from various sources:
// - PEM content (if it starts with "-----BEGIN")
// - File path (if the file exists)
// - HTTP/HTTPS URL (if it starts with "http://" or "https://")
func loadCACert(ctx context.Context, caCert string) ([]byte, error) {
	if caCert == "" {
		return nil, nil
	}

	// Check if it's PEM content (starts with BEGIN marker)
	if strings.HasPrefix(strings.TrimSpace(caCert), "-----BEGIN") {
		return []byte(caCert), nil
	}

	// Check if it's an HTTP/HTTPS URL
	if strings.HasPrefix(caCert, "http://") || strings.HasPrefix(caCert, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, caCert, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for CA certificate URL: %w", err)
		}
		resp, err := http.DefaultClient.Do(req) // #nosec G107 -- URL is user-provided configuration
		if err != nil {
			return nil, fmt.Errorf("failed to fetch CA certificate from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch CA certificate: HTTP %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from URL: %w", err)
		}
		return data, nil
	}

	// Otherwise, treat it as a file path
	data, err := os.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}
	return data, nil
}

// NewJenkins is initial Jenkins object
func NewJenkins(
	ctx context.Context,
	auth *Auth,
	baseURL string,
	token string,
	insecure bool,
	caCert string,
	debug bool,
) (*Jenkins, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	// Load CA certificate if provided
	caCertData, err := loadCACert(ctx, caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Build TLS configuration
	var tlsConfig *tls.Config
	if insecure {
		// #nosec G402 -- InsecureSkipVerify is intentionally configurable by user
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	} else if caCertData != nil {
		// Create certificate pool with custom CA
		certPool, err := x509.SystemCertPool()
		if err != nil {
			// Fall back to empty pool if system pool unavailable
			certPool = x509.NewCertPool()
		}

		if !certPool.AppendCertsFromPEM(caCertData) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig = &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		}
	}

	// Create HTTP client
	client := http.DefaultClient
	if tlsConfig != nil {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
	}

	return &Jenkins{
		Auth:    auth,
		BaseURL: baseURL,
		Token:   token,
		Client:  client,
		Debug:   debug,
	}, nil
}

func (jenkins *Jenkins) buildURL(path string, params url.Values) (requestURL string) {
	requestURL = jenkins.BaseURL + path
	if params != nil {
		queryString := params.Encode()
		if queryString != "" {
			requestURL = requestURL + "?" + queryString
		}
	}

	return
}

func (jenkins *Jenkins) sendRequest(req *http.Request) (*http.Response, error) {
	if jenkins.Auth != nil {
		req.SetBasicAuth(jenkins.Auth.Username, jenkins.Auth.Token)
	}
	return jenkins.Client.Do(req)
}

func (jenkins *Jenkins) get(
	ctx context.Context,
	path string,
	params url.Values,
	body interface{},
) error {
	requestURL := jenkins.buildURL(path, params)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return err
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, string(data))
	}

	if body == nil {
		return nil
	}

	return json.Unmarshal(data, body)
}

// postAndGetLocation performs a POST request and extracts the queue ID from Location header
func (jenkins *Jenkins) postAndGetLocation(
	ctx context.Context,
	path string,
	params url.Values,
) (int, error) {
	requestURL := jenkins.buildURL(path, params)

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := jenkins.sendRequest(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"unexpected response code: %d, body: %s",
			resp.StatusCode,
			string(data),
		)
	}

	// Extract queue ID from Location header
	// Location format: http://jenkins.example.com/queue/item/123/
	location := resp.Header.Get("Location")
	if location == "" {
		return 0, fmt.Errorf("no Location header in response")
	}

	// Parse queue ID from URL
	// Look for /queue/item/{id}/ or /queue/item/{id}
	var queueID int
	// Find the pattern "/queue/item/" and extract the number after it
	queueItemPrefix := "/queue/item/"
	idx := strings.Index(location, queueItemPrefix)
	if idx == -1 {
		return 0, fmt.Errorf("failed to parse queue ID from Location: %s", location)
	}

	// Extract the substring after "/queue/item/"
	afterPrefix := location[idx+len(queueItemPrefix):]
	// Parse the number (stop at / or end of string)
	if _, err := fmt.Sscanf(afterPrefix, "%d", &queueID); err != nil {
		return 0, fmt.Errorf("failed to parse queue ID from Location: %s", location)
	}

	return queueID, nil
}

func (jenkins *Jenkins) parseJobPath(job string) string {
	var path string

	jobs := strings.Split(strings.TrimPrefix(job, "/"), "/")

	for _, value := range jobs {
		value = strings.Trim(value, " ")
		if len(value) == 0 {
			continue
		}

		path = fmt.Sprintf("%s/job/%s", path, value)
	}

	return path
}

// getQueueItem fetches information about a queue item
func (jenkins *Jenkins) getQueueItem(ctx context.Context, queueID int) (*QueueItem, error) {
	path := fmt.Sprintf("/queue/item/%d/api/json", queueID)

	var queueItem QueueItem
	err := jenkins.get(ctx, path, nil, &queueItem)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue item %d: %w", queueID, err)
	}

	return &queueItem, nil
}

// getBuildInfo fetches information about a specific build
func (jenkins *Jenkins) getBuildInfo(
	ctx context.Context,
	job string,
	buildNumber int,
) (*BuildInfo, error) {
	path := fmt.Sprintf("%s/%d/api/json", jenkins.parseJobPath(job), buildNumber)

	var buildInfo BuildInfo
	err := jenkins.get(ctx, path, nil, &buildInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get build info for %s #%d: %w", job, buildNumber, err)
	}

	return &buildInfo, nil
}

// waitForCompletion waits for a Jenkins build to complete
// It first polls the queue to get the build number, then polls the build status until completion
func (jenkins *Jenkins) waitForCompletion(
	ctx context.Context,
	job string,
	queueID int,
	pollInterval, timeout time.Duration,
) (*BuildInfo, error) {
	deadline := time.Now().Add(timeout)

	// Phase 1: Wait for queue item to be assigned a build number
	log.Printf("waiting for job %s (queue #%d) to start...", job, queueID)
	var buildNumber int

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for job %s to start", job)
		}

		queueItem, err := jenkins.getQueueItem(ctx, queueID)
		if err != nil {
			// Queue item might be deleted after build starts, try to continue
			log.Printf("warning: failed to get queue item: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		// Check if build has started
		if queueItem.Executable != nil && queueItem.Executable.Number > 0 {
			buildNumber = queueItem.Executable.Number
			log.Printf("job %s started as build #%d", job, buildNumber)
			break
		}

		// Log why the job is waiting if available
		if queueItem.Why != "" {
			log.Printf("job %s is queued: %s", job, queueItem.Why)
		}

		time.Sleep(pollInterval)
	}

	// Phase 2: Wait for build to complete
	log.Printf("waiting for job %s (build #%d) to complete...", job, buildNumber)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf(
				"timeout waiting for job %s build #%d to complete",
				job,
				buildNumber,
			)
		}

		buildInfo, err := jenkins.getBuildInfo(ctx, job, buildNumber)
		if err != nil {
			log.Printf("warning: failed to get build info: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		// Check if build is complete
		if !buildInfo.Building {
			log.Printf(
				"job %s (build #%d) completed with status: %s",
				job,
				buildNumber,
				buildInfo.Result,
			)

			// Debug: Display final build info
			if jenkins.Debug {
				log.Println("=== Debug Mode: Build Result ===")
				if err := godump.Dump(buildInfo); err != nil {
					log.Printf("warning: failed to dump build info: %v", err)
				}
				log.Println("================================")
			}

			// Set GitHub Actions output
			if err := gh.SetOutput(map[string]string{
				"result": buildInfo.Result,
				"url":    buildInfo.URL,
			}); err != nil {
				log.Printf("warning: failed to set GitHub output: %v", err)
			}

			return buildInfo, nil
		}

		time.Sleep(pollInterval)
	}
}

func (jenkins *Jenkins) trigger(ctx context.Context, job string, params url.Values) (int, error) {
	// Add remote trigger token to params
	if jenkins.Token != "" {
		if params == nil {
			params = url.Values{}
		}
		params.Set("token", jenkins.Token)
	}

	var urlPath string
	// Check if params contains build parameters (excluding 'token')
	hasBuildParams := false
	for key := range params {
		if key != "token" {
			hasBuildParams = true
			break
		}
	}

	if hasBuildParams {
		urlPath = jenkins.parseJobPath(job) + "/buildWithParameters"
	} else {
		urlPath = jenkins.parseJobPath(job) + "/build"
	}

	// Debug: Display parameters being sent
	if jenkins.Debug {
		log.Println("=== Debug Mode: Jenkins Job Trigger ===")
		log.Printf("Job: %s", job)
		log.Printf("URL Path: %s", urlPath)

		// Build the full URL for display
		fullURL := jenkins.buildURL(urlPath, params)
		// Mask token in URL for display
		if jenkins.Token != "" {
			fullURL = strings.Replace(fullURL, "token="+jenkins.Token, "token=***MASKED***", 1)
		}
		log.Printf("Full URL: %s", fullURL)

		if len(params) > 0 {
			// Create a copy of params with masked token for display
			displayParams := url.Values{}
			for key, values := range params {
				if key == "token" {
					// Mask token values for security
					displayParams[key] = []string{"***MASKED***"}
				} else {
					displayParams[key] = values
				}
			}

			log.Println("Parameters:")
			if err := godump.Dump(displayParams); err != nil {
				log.Printf("warning: failed to dump parameters: %v", err)
			}
		} else {
			log.Println("Parameters: (none)")
		}
		log.Println("======================================")
	}

	// All params (including token) are passed as query parameters
	// Returns the queue item ID for tracking
	return jenkins.postAndGetLocation(ctx, urlPath, params)
}
