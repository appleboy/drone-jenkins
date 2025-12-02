package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
	jenkins := NewJenkins(auth, "http://example.com", "", false, false)

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
	jenkins := NewJenkins(auth, "example.com", "", false, false)

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
	jenkins := NewJenkins(auth, server.URL, "remote-token", false, false)

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
			jenkins := NewJenkins(auth, server.URL, "", false, false)

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
			jenkins := NewJenkins(auth, server.URL, "", false, false)

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
			jenkins := NewJenkins(auth, server.URL, "", false, false)

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
		jenkins := NewJenkins(auth, server.URL, "", false, false)

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
		jenkins := NewJenkins(auth, server.URL, "", false, false)

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
		jenkins := NewJenkins(auth, server.URL, "", false, false)

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
		jenkins := NewJenkins(auth, server.URL, "", false, false)

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
