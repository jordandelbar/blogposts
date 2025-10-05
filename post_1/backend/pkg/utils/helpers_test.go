package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           any
		expectedStatus int
		expectedBody   string
		expectedError  bool
	}{
		// Valid inputs
		{
			name:           "simple string",
			status:         200,
			data:           "hello world",
			expectedStatus: 200,
			expectedBody:   `"hello world"`,
			expectedError:  false,
		},
		{
			name:           "simple object",
			status:         200,
			data:           map[string]string{"message": "success"},
			expectedStatus: 200,
			expectedBody:   `{"message":"success"}`,
			expectedError:  false,
		},
		{
			name:           "array of strings",
			status:         200,
			data:           []string{"apple", "banana", "cherry"},
			expectedStatus: 200,
			expectedBody:   `["apple","banana","cherry"]`,
			expectedError:  false,
		},
		{
			name:           "number",
			status:         200,
			data:           42,
			expectedStatus: 200,
			expectedBody:   `42`,
			expectedError:  false,
		},
		{
			name:           "boolean true",
			status:         200,
			data:           true,
			expectedStatus: 200,
			expectedBody:   `true`,
			expectedError:  false,
		},
		{
			name:           "boolean false",
			status:         200,
			data:           false,
			expectedStatus: 200,
			expectedBody:   `false`,
			expectedError:  false,
		},
		{
			name:           "null value",
			status:         200,
			data:           nil,
			expectedStatus: 200,
			expectedBody:   `null`,
			expectedError:  false,
		},
		{
			name:           "empty object",
			status:         200,
			data:           map[string]any{},
			expectedStatus: 200,
			expectedBody:   `{}`,
			expectedError:  false,
		},
		{
			name:           "empty array",
			status:         200,
			data:           []any{},
			expectedStatus: 200,
			expectedBody:   `[]`,
			expectedError:  false,
		},
		{
			name:           "nested object",
			status:         201,
			data:           map[string]any{"user": map[string]any{"id": 1, "name": "John"}},
			expectedStatus: 201,
			expectedBody:   `{"user":{"id":1,"name":"John"}}`,
			expectedError:  false,
		},
		{
			name:           "envelope pattern",
			status:         200,
			data:           Envelope{"data": "test", "success": true},
			expectedStatus: 200,
			expectedBody:   `{"data":"test","success":true}`,
			expectedError:  false,
		},
		{
			name:           "status 404",
			status:         404,
			data:           map[string]string{"error": "not found"},
			expectedStatus: 404,
			expectedBody:   `{"error":"not found"}`,
			expectedError:  false,
		},
		{
			name:           "status 500",
			status:         500,
			data:           map[string]string{"error": "internal server error"},
			expectedStatus: 500,
			expectedBody:   `{"error":"internal server error"}`,
			expectedError:  false,
		},
		{
			name:   "complex nested structure",
			status: 200,
			data: map[string]any{
				"articles": []map[string]any{
					{"id": 1, "title": "First Post", "published": true},
					{"id": 2, "title": "Second Post", "published": false},
				},
				"meta": map[string]any{
					"total": 2,
					"page":  1,
				},
			},
			expectedStatus: 200,
			expectedBody:   `{"articles":[{"id":1,"published":true,"title":"First Post"},{"id":2,"published":false,"title":"Second Post"}],"meta":{"page":1,"total":2}}`,
			expectedError:  false,
		},

		// Invalid inputs that should cause JSON marshal errors
		{
			name:           "function (non-serializable)",
			status:         200,
			data:           func() {},
			expectedStatus: 0, // Status won't be set due to error
			expectedBody:   "",
			expectedError:  true,
		},
		{
			name:           "channel (non-serializable)",
			status:         200,
			data:           make(chan int),
			expectedStatus: 0,
			expectedBody:   "",
			expectedError:  true,
		},
		{
			name:           "complex number (non-serializable)",
			status:         200,
			data:           complex(1, 2),
			expectedStatus: 0,
			expectedBody:   "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := WriteJSON(w, tt.status, tt.data)

			if tt.expectedError {
				if err == nil {
					t.Errorf("WriteJSON() expected an error, but got none")
				}
				return // Skip other checks if we expected an error
			}

			// Check for unexpected error
			if err != nil {
				t.Errorf("WriteJSON() unexpected error: %v", err)
				return
			}

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("WriteJSON() status = %d; expected %d", w.Code, tt.expectedStatus)
			}

			// Check Content-Type header
			expectedContentType := "application/json"
			if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
				t.Errorf("WriteJSON() Content-Type = %q; expected %q", contentType, expectedContentType)
			}

			// Check body content
			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("WriteJSON() body = %q; expected %q", body, tt.expectedBody)
			}

			// Verify the JSON is valid by unmarshaling it
			if !tt.expectedError && body != "" {
				var result any
				if err := json.Unmarshal([]byte(body), &result); err != nil {
					t.Errorf("WriteJSON() produced invalid JSON: %v", err)
				}
			}
		})
	}
}

// Test WriteJSON with a custom ResponseWriter that fails on Write
func TestWriteJSONWriteError(t *testing.T) {
	// Create a ResponseWriter that fails on Write
	w := &failingResponseWriter{}

	data := map[string]string{"test": "data"}
	err := WriteJSON(w, 200, data)

	if err == nil {
		t.Errorf("WriteJSON() expected an error when Write fails, but got none")
	}

	expectedError := "write failed"
	if err.Error() != expectedError {
		t.Errorf("WriteJSON() error = %q; expected %q", err.Error(), expectedError)
	}
}

// failingResponseWriter is a mock ResponseWriter that fails on Write
type failingResponseWriter struct {
	header http.Header
}

func (f *failingResponseWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (f *failingResponseWriter) WriteHeader(statusCode int) {
	// Do nothing
}
