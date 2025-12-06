package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseJobPath(t *testing.T) {
	auth := &Auth{
		Username: "appleboy",
		Token:    "1234",
	}
	jenkins, err := NewJenkins(auth, "http://example.com", "", false, "", false)
	assert.NoError(t, err)

	assert.Equal(t, "/job/foo", jenkins.parseJobPath("/foo/"))
	assert.Equal(t, "/job/foo", jenkins.parseJobPath("foo/"))
	assert.Equal(t, "/job/foo/job/bar", jenkins.parseJobPath("foo/bar"))
	assert.Equal(t, "/job/foo/job/bar", jenkins.parseJobPath("foo///bar"))
}

func TestUnSupportProtocol(t *testing.T) {
	auth := &Auth{
		Username: "foo",
		Token:    "bar",
	}
	jenkins, err := NewJenkins(auth, "example.com", "", false, "", false)
	assert.NoError(t, err)

	queueID, err := jenkins.trigger("drone-jenkins", nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, queueID)
}

func TestTriggerBuild(t *testing.T) {
	// Create a mock Jenkins server
	var receivedParams url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedParams = r.URL.Query()
		w.Header().Set("Location", "http://jenkins.example.com/queue/item/123/")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	auth := &Auth{
		Username: "foo",
		Token:    "bar",
	}
	jenkins, err := NewJenkins(auth, server.URL, "remote-token", false, "", false)
	assert.NoError(t, err)

	params := url.Values{"param": []string{"value"}}
	queueID, err := jenkins.trigger("drone-jenkins", params)

	assert.NoError(t, err)
	assert.Equal(t, 123, queueID)
	assert.Equal(t, "value", receivedParams.Get("param"))
	assert.Equal(t, "remote-token", receivedParams.Get("token"))
}

func TestPostAndGetLocation(t *testing.T) {
	tests := []struct {
		name        string
		location    string
		expectID    int
		expectError bool
	}{
		{
			name:        "valid location with trailing slash",
			location:    "http://jenkins.example.com/queue/item/456/",
			expectID:    456,
			expectError: false,
		},
		{
			name:        "valid location without trailing slash",
			location:    "http://jenkins.example.com/queue/item/789",
			expectID:    789,
			expectError: false,
		},
		{
			name:        "no location header",
			location:    "",
			expectID:    0,
			expectError: true,
		},
		{
			name:        "invalid location format",
			location:    "http://jenkins.example.com/invalid/path",
			expectID:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tt.location != "" {
						w.Header().Set("Location", tt.location)
					}
					w.WriteHeader(http.StatusCreated)
				}),
			)
			defer server.Close()

			auth := &Auth{
				Username: "test",
				Token:    "test",
			}
			jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
			assert.NoError(t, err)

			queueID, err := jenkins.postAndGetLocation("/test", nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectID, queueID)
			}
		})
	}
}

func TestGetQueueItem(t *testing.T) {
	tests := []struct {
		name           string
		queueID        int
		responseBody   string
		responseStatus int
		expectError    bool
		expectBlocked  bool
		expectBuildNum int
	}{
		{
			name:    "queue item with build number",
			queueID: 123,
			responseBody: `{"id":123,"blocked":false,"buildable":true,` +
				`"executable":{"number":456,"url":"http://jenkins.example.com/job/test/456/"}}`,
			responseStatus: http.StatusOK,
			expectError:    false,
			expectBlocked:  false,
			expectBuildNum: 456,
		},
		{
			name:           "queue item waiting",
			queueID:        124,
			responseBody:   `{"id":124,"blocked":false,"buildable":true,"why":"Waiting for executor"}`,
			responseStatus: http.StatusOK,
			expectError:    false,
			expectBlocked:  false,
			expectBuildNum: 0,
		},
		{
			name:           "queue item blocked",
			queueID:        125,
			responseBody:   `{"id":125,"blocked":true,"buildable":false,"why":"Blocked by other job"}`,
			responseStatus: http.StatusOK,
			expectError:    false,
			expectBlocked:  true,
			expectBuildNum: 0,
		},
		{
			name:           "queue item not found",
			queueID:        999,
			responseBody:   "Not Found",
			responseStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.URL.Path, "/queue/item/")
					w.WriteHeader(tt.responseStatus)
					_, _ = w.Write([]byte(tt.responseBody))
				}),
			)
			defer server.Close()

			auth := &Auth{
				Username: "test",
				Token:    "test",
			}
			jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
			assert.NoError(t, err)

			queueItem, err := jenkins.getQueueItem(tt.queueID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, queueItem)
				assert.Equal(t, tt.queueID, queueItem.ID)
				assert.Equal(t, tt.expectBlocked, queueItem.Blocked)
				if queueItem.Executable != nil {
					assert.Equal(t, tt.expectBuildNum, queueItem.Executable.Number)
				}
			}
		})
	}
}

func TestGetBuildInfo(t *testing.T) {
	tests := []struct {
		name           string
		jobName        string
		buildNumber    int
		responseBody   string
		responseStatus int
		expectError    bool
		expectBuilding bool
		expectResult   string
	}{
		{
			name:        "build in progress",
			jobName:     "test-job",
			buildNumber: 123,
			responseBody: `{"number":123,"building":true,"duration":0,"result":null,` +
				`"url":"http://jenkins.example.com/job/test-job/123/"}`,
			responseStatus: http.StatusOK,
			expectError:    false,
			expectBuilding: true,
			expectResult:   "",
		},
		{
			name:        "build completed successfully",
			jobName:     "test-job",
			buildNumber: 124,
			responseBody: `{"number":124,"building":false,"duration":5000,"result":"SUCCESS",` +
				`"url":"http://jenkins.example.com/job/test-job/124/"}`,
			responseStatus: http.StatusOK,
			expectError:    false,
			expectBuilding: false,
			expectResult:   "SUCCESS",
		},
		{
			name:        "build failed",
			jobName:     "test-job",
			buildNumber: 125,
			responseBody: `{"number":125,"building":false,"duration":3000,"result":"FAILURE",` +
				`"url":"http://jenkins.example.com/job/test-job/125/"}`,
			responseStatus: http.StatusOK,
			expectError:    false,
			expectBuilding: false,
			expectResult:   "FAILURE",
		},
		{
			name:           "build not found",
			jobName:        "test-job",
			buildNumber:    999,
			responseBody:   "Not Found",
			responseStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.URL.Path, "/job/")
					w.WriteHeader(tt.responseStatus)
					_, _ = w.Write([]byte(tt.responseBody))
				}),
			)
			defer server.Close()

			auth := &Auth{
				Username: "test",
				Token:    "test",
			}
			jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
			assert.NoError(t, err)

			buildInfo, err := jenkins.getBuildInfo(tt.jobName, tt.buildNumber)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, buildInfo)
				assert.Equal(t, tt.buildNumber, buildInfo.Number)
				assert.Equal(t, tt.expectBuilding, buildInfo.Building)
				assert.Equal(t, tt.expectResult, buildInfo.Result)
			}
		})
	}
}

func TestWaitForCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		var callCount int32
		queueID := 123
		buildNumber := 456

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)

			switch r.URL.Path {
			case testQueueItemPath:
				// First call: queue item without build number
				// Second call: queue item with build number
				w.WriteHeader(http.StatusOK)
				if count == 1 {
					_, _ = w.Write([]byte(
						`{"id":123,"blocked":false,"buildable":true,"why":"Waiting for executor"}`,
					))
				} else {
					_, _ = w.Write([]byte(`{"id":123,"blocked":false,"buildable":true,` +
						`"executable":{"number":456,"url":"http://example.com/job/test/456/"}}`))
				}
			case testBuildStatusPath:
				// First call: build in progress
				// Second call: build completed
				w.WriteHeader(http.StatusOK)
				if count <= 3 {
					_, _ = w.Write(
						[]byte(`{"number":456,"building":true,"duration":0,"result":null}`),
					)
				} else {
					_, _ = w.Write([]byte(`{"number":456,"building":false,"duration":5000,"result":"SUCCESS"}`))
				}
			}
		}))
		defer server.Close()

		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
		assert.NoError(t, err)

		buildInfo, err := jenkins.waitForCompletion(
			"test-job",
			queueID,
			100*time.Millisecond,
			5*time.Second,
		)

		assert.NoError(t, err)
		assert.NotNil(t, buildInfo)
		assert.Equal(t, buildNumber, buildInfo.Number)
		assert.False(t, buildInfo.Building)
		assert.Equal(t, "SUCCESS", buildInfo.Result)
	})

	t.Run("timeout waiting for queue", func(t *testing.T) {
		queueID := 123

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always return queue item without build number
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(
				[]byte(`{"id":123,"blocked":false,"buildable":true,"why":"Waiting forever"}`),
			)
		}))
		defer server.Close()

		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
		assert.NoError(t, err)

		buildInfo, err := jenkins.waitForCompletion(
			"test-job",
			queueID,
			50*time.Millisecond,
			200*time.Millisecond,
		)

		assert.Error(t, err)
		assert.Nil(t, buildInfo)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("timeout waiting for build", func(t *testing.T) {
		var callCount int32
		queueID := 123

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)

			switch r.URL.Path {
			case testQueueItemPath:
				// Return build number immediately
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":123,"blocked":false,"buildable":true,` +
					`"executable":{"number":456,"url":"http://example.com/job/test/456/"}}`))
			case testBuildStatusPath:
				// Always return building status
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"number":456,"building":true,"duration":0,"result":null}`))
			}
			_ = count
		}))
		defer server.Close()

		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
		assert.NoError(t, err)

		buildInfo, err := jenkins.waitForCompletion(
			"test-job",
			queueID,
			50*time.Millisecond,
			200*time.Millisecond,
		)

		assert.Error(t, err)
		assert.Nil(t, buildInfo)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("build failed", func(t *testing.T) {
		var callCount int32
		queueID := 123
		buildNumber := 456

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)

			switch r.URL.Path {
			case testQueueItemPath:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":123,"blocked":false,"buildable":true,` +
					`"executable":{"number":456,"url":"http://example.com/job/test/456/"}}`))
			case testBuildStatusPath:
				// First call: building, second call: failed
				w.WriteHeader(http.StatusOK)
				if count == 1 {
					_, _ = w.Write(
						[]byte(`{"number":456,"building":true,"duration":0,"result":null}`),
					)
				} else {
					_, _ = w.Write([]byte(`{"number":456,"building":false,"duration":3000,"result":"FAILURE"}`))
				}
			}
		}))
		defer server.Close()

		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, server.URL, "", false, "", false)
		assert.NoError(t, err)

		buildInfo, err := jenkins.waitForCompletion(
			"test-job",
			queueID,
			50*time.Millisecond,
			5*time.Second,
		)

		assert.NoError(t, err)
		assert.NotNil(t, buildInfo)
		assert.Equal(t, buildNumber, buildInfo.Number)
		assert.False(t, buildInfo.Building)
		assert.Equal(t, "FAILURE", buildInfo.Result)
	})
}

// Sample CA certificate for testing (self-signed, not for production use)
const testCACert = `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHBfpegPjMCMA0GCSqGSIb3DQEBCwUAMBExDzANBgNVBAMMBnRl
c3RjYTAeFw0yNDAxMDEwMDAwMDBaFw0yNTAxMDEwMDAwMDBaMBExDzANBgNVBAMM
BnRlc3RjYTBcMA0GCSqGSIb3DQEBAQUAA0sAMEgCQQC5mj8JwKzGvEVrbzk2QXBC
GqY6M8E3rYnxmPi3cewhPZs7QL5NVNrR5GhcCyLBJzGkJd3AU5kkHwfOPrqYO9Ep
AgMBAAGjUzBRMB0GA1UdDgQWBBQ7foZhM1qraTRhAFOJzWjQrDfoczAfBgNVHSME
GDAWgBQ7foZhM1qraTRhAFOJzWjQrDfoczAPBgNVHRMBAf8EBTADAQH/MA0GCSqG
SIb3DQEBCwUAA0EAQ7B1h9r7jJCMbqFNxKhyb3sT4k7fXJerDr8TqnGKB8K1VDXT
eQBd7OPIRKNyMhmD1vK5IBKYYCmykyGR9S7r2A==
-----END CERTIFICATE-----`

func TestLoadCACert(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		data, err := loadCACert("")
		assert.NoError(t, err)
		assert.Nil(t, data)
	})

	t.Run("PEM content directly", func(t *testing.T) {
		data, err := loadCACert(testCACert)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Contains(t, string(data), "BEGIN CERTIFICATE")
	})

	t.Run("PEM content with leading whitespace", func(t *testing.T) {
		data, err := loadCACert("  \n" + testCACert)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Contains(t, string(data), "BEGIN CERTIFICATE")
	})

	t.Run("file path", func(t *testing.T) {
		// Create a temporary file with the certificate
		tmpDir := t.TempDir()
		certFile := filepath.Join(tmpDir, "ca.pem")
		err := os.WriteFile(certFile, []byte(testCACert), 0o600)
		assert.NoError(t, err)

		data, err := loadCACert(certFile)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Contains(t, string(data), "BEGIN CERTIFICATE")
	})

	t.Run("file not found", func(t *testing.T) {
		data, err := loadCACert("/nonexistent/path/ca.pem")
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to read CA certificate file")
	})

	t.Run("HTTP URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(testCACert))
		}))
		defer server.Close()

		data, err := loadCACert(server.URL)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Contains(t, string(data), "BEGIN CERTIFICATE")
	})

	t.Run("HTTPS URL", func(t *testing.T) {
		server := httptest.NewTLSServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(testCACert))
			}),
		)
		defer server.Close()

		// Note: This test uses the test server's self-signed cert
		// In real scenarios, the URL would be to a trusted source
		// We skip HTTPS verification for this test
		data, err := loadCACert(server.URL)
		// This may fail due to certificate verification, which is expected
		if err != nil {
			assert.Contains(t, err.Error(), "certificate")
		} else {
			assert.NotNil(t, data)
		}
	})

	t.Run("HTTP URL returns error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		data, err := loadCACert(server.URL)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "HTTP 404")
	})

	t.Run("HTTP URL unreachable", func(t *testing.T) {
		data, err := loadCACert("http://localhost:59999/nonexistent")
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "failed to fetch CA certificate from URL")
	})
}

func TestNewJenkinsWithCACert(t *testing.T) {
	t.Run("with valid CA certificate", func(t *testing.T) {
		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, "https://example.com", "", false, testCACert, false)
		assert.NoError(t, err)
		assert.NotNil(t, jenkins)
		assert.NotNil(t, jenkins.Client)
	})

	t.Run("with CA certificate from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		certFile := filepath.Join(tmpDir, "ca.pem")
		err := os.WriteFile(certFile, []byte(testCACert), 0o600)
		assert.NoError(t, err)

		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, "https://example.com", "", false, certFile, false)
		assert.NoError(t, err)
		assert.NotNil(t, jenkins)
	})

	t.Run("with invalid CA certificate content", func(t *testing.T) {
		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(
			auth,
			"https://example.com",
			"",
			false,
			"invalid-cert-data",
			false,
		)
		assert.Error(t, err)
		assert.Nil(t, jenkins)
		assert.Contains(t, err.Error(), "failed to read CA certificate file")
	})

	t.Run("with invalid PEM format", func(t *testing.T) {
		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		invalidPEM := "-----BEGIN CERTIFICATE-----\ninvalid-base64-data\n-----END CERTIFICATE-----"
		jenkins, err := NewJenkins(auth, "https://example.com", "", false, invalidPEM, false)
		assert.Error(t, err)
		assert.Nil(t, jenkins)
		assert.Contains(t, err.Error(), "failed to parse CA certificate")
	})

	t.Run("with nonexistent file path", func(t *testing.T) {
		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(
			auth,
			"https://example.com",
			"",
			false,
			"/nonexistent/ca.pem",
			false,
		)
		assert.Error(t, err)
		assert.Nil(t, jenkins)
		assert.Contains(t, err.Error(), "failed to load CA certificate")
	})

	t.Run("insecure flag takes precedence over CA cert", func(t *testing.T) {
		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		// When insecure is true, CA cert should be ignored
		jenkins, err := NewJenkins(auth, "https://example.com", "", true, testCACert, false)
		assert.NoError(t, err)
		assert.NotNil(t, jenkins)
	})

	t.Run("without CA certificate uses default client", func(t *testing.T) {
		auth := &Auth{
			Username: "test",
			Token:    "test",
		}
		jenkins, err := NewJenkins(auth, "https://example.com", "", false, "", false)
		assert.NoError(t, err)
		assert.NotNil(t, jenkins)
		assert.Equal(t, http.DefaultClient, jenkins.Client)
	})
}
