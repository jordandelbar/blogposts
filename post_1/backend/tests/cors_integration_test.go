package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedStatus int
		expectedCORS   bool
		preflightTest  bool
	}{
		{
			name:           "No origin header should work but no CORS headers",
			origin:         "",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedCORS:   false,
		},
		{
			name:           "Trusted origin should get CORS headers",
			origin:         "http://localhost:3000",
			method:         "GET", 
			expectedStatus: http.StatusOK,
			expectedCORS:   true,
		},
		{
			name:           "Another trusted origin should get CORS headers",
			origin:         "https://example.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedCORS:   true,
		},
		{
			name:           "Untrusted origin should not get CORS headers",
			origin:         "https://malicious-site.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedCORS:   false,
		},
		{
			name:           "Preflight request with trusted origin should succeed",
			origin:         "http://localhost:3000",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			expectedCORS:   true,
			preflightTest:  true,
		},
		{
			name:           "Preflight request with untrusted origin should not get CORS headers", 
			origin:         "https://malicious-site.com",
			method:         "OPTIONS",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCORS:   false,
			preflightTest:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.ServerAddr+"/health", nil)
			require.NoError(t, err)

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			if tt.preflightTest {
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "Content-Type")
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			
			// Check Vary header is always present
			assert.Equal(t, "Origin", resp.Header.Get("Vary"))

			// Check CORS headers
			corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
			if tt.expectedCORS {
				assert.Equal(t, tt.origin, corsOrigin, "Expected CORS origin to match request origin")
				
				if tt.preflightTest && tt.expectedStatus == http.StatusOK {
					assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
					assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
				}
			} else {
				assert.Empty(t, corsOrigin, "Expected no CORS origin header for untrusted/no origin")
			}
		})
	}
}

func TestCORSWithDifferentEndpoints(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)

	endpoints := []string{"/health", "/v1/articles"}
	
	for _, endpoint := range endpoints {
		t.Run("Endpoint "+endpoint, func(t *testing.T) {
			req, err := http.NewRequest("GET", ts.ServerAddr+endpoint, nil)
			require.NoError(t, err)
			req.Header.Set("Origin", "http://localhost:3000")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should get CORS headers for trusted origin
			assert.Equal(t, "http://localhost:3000", resp.Header.Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "Origin", resp.Header.Get("Vary"))
		})
	}
}