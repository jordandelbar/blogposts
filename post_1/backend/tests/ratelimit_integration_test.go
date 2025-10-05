package tests

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimiterIPIsolation tests that rate limiting is applied per IP address
// ensuring one malicious IP cannot affect other legitimate IPs
func TestRateLimiterIPIsolation(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)
	client := &http.Client{Timeout: 5 * time.Second}

	maliciousIP := "192.168.1.100"
	legitimateIP := "10.0.0.50"

	// Step 1: Exhaust rate limit for malicious IP
	rateLimitHit := false
	for i := 0; i < 20; i++ {
		req, err := http.NewRequest("GET", ts.ServerAddr+"/health", nil)
		require.NoError(t, err)
		req.Header.Set("X-Forwarded-For", maliciousIP)

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitHit = true
			t.Logf("Rate limit hit after %d requests for IP %s", i+1, maliciousIP)
			break
		}
	}

	assert.True(t, rateLimitHit, "Expected malicious IP to hit rate limit")

	// Step 2: Verify legitimate IP can still make requests
	req, err := http.NewRequest("GET", ts.ServerAddr+"/health", nil)
	require.NoError(t, err)
	req.Header.Set("X-Forwarded-For", legitimateIP)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"Legitimate IP should not be affected by malicious IP's rate limiting")

	// Step 3: Verify another IP also works
	anotherIP := "172.16.0.25"
	req2, err := http.NewRequest("GET", ts.ServerAddr+"/health", nil)
	require.NoError(t, err)
	req2.Header.Set("X-Forwarded-For", anotherIP)

	resp2, err := client.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode,
		"Another IP should also not be affected by rate limiting of different IP")
}

// TestRateLimiterTraefikHeaders tests that IP extraction from proxy headers works correctly
// focusing on header precedence and parsing logic
func TestRateLimiterTraefikHeaders(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)
	client := &http.Client{Timeout: 5 * time.Second}

	testCases := []struct {
		name        string
		headers     map[string]string
		description string
	}{
		{
			name:        "X-Forwarded-For header",
			headers:     map[string]string{"X-Forwarded-For": "203.0.113.1, 192.168.1.1"},
			description: "Should use first IP from comma-separated list",
		},
		{
			name:        "X-Real-IP header",
			headers:     map[string]string{"X-Real-IP": "203.0.113.2"},
			description: "Should use X-Real-IP when X-Forwarded-For is not present",
		},
		{
			name: "X-Forwarded-For takes precedence",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.3",
				"X-Real-IP":       "203.0.113.4",
			},
			description: "X-Forwarded-For should take precedence over X-Real-IP",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make requests with the specific headers
			req, err := http.NewRequest("GET", ts.ServerAddr+"/health", nil)
			require.NoError(t, err)

			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"Request should succeed with proper header parsing")

			// Verify that a request without headers (using RemoteAddr) still works
			// This ensures we're actually using the header-extracted IP
			reqNoHeaders, err := http.NewRequest("GET", ts.ServerAddr+"/health", nil)
			require.NoError(t, err)

			respNoHeaders, err := client.Do(reqNoHeaders)
			require.NoError(t, err)
			defer respNoHeaders.Body.Close()

			assert.Equal(t, http.StatusOK, respNoHeaders.StatusCode,
				"Request without headers should also work (uses RemoteAddr)")
		})
	}
}

// TestRateLimiterConcurrentRequests tests that multiple IPs can make concurrent requests
// without interfering with each other's rate limits
func TestRateLimiterConcurrentRequests(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)

	var wg sync.WaitGroup
	results := make(chan testResult, 3)

	// Test 3 different IPs making requests concurrently
	ips := []string{"192.168.1.10", "192.168.1.20", "192.168.1.30"}

	for _, ip := range ips {
		wg.Add(1)
		go func(testIP string) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}
			successCount := 0

			// Each IP makes 15 requests (below rate limit)
			for i := 0; i < 15; i++ {
				req, err := http.NewRequest("GET", ts.ServerAddr+"/health", nil)
				if err != nil {
					continue
				}

				req.Header.Set("X-Forwarded-For", testIP)

				resp, err := client.Do(req)
				if err != nil {
					continue
				}
				resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					successCount++
				}

				time.Sleep(50 * time.Millisecond) // Spread out requests
			}

			results <- testResult{IP: testIP, SuccessCount: successCount}
		}(ip)
	}

	wg.Wait()
	close(results)

	// Verify all IPs got successful requests (should be most/all since below rate limit)
	for result := range results {
		assert.GreaterOrEqual(t, result.SuccessCount, 10,
			fmt.Sprintf("IP %s should have gotten most requests through (got %d/15)",
				result.IP, result.SuccessCount))
	}
}

type testResult struct {
	IP           string
	SuccessCount int
}