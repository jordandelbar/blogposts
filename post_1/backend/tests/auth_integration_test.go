package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationToken_ValidCredentials_ReturnsToken(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	_ = createTestUser(t, queries, datastore)

	authData := map[string]string{
		"email":    "test@example.com",
		"password": "TestPassword123!",
	}

	jsonData, err := json.Marshal(authData)
	require.NoError(t, err)

	resp, err := http.Post(server_addr+"/v1/auth/token", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Verify response contains authentication token
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var authResp struct {
		AuthenticationToken string `json:"authentication_token"`
		Expiry              string `json:"expiry"`
	}
	err = json.Unmarshal(body, &authResp)
	require.NoError(t, err)

	assert.NotEmpty(t, authResp.AuthenticationToken)
	assert.NotEmpty(t, authResp.Expiry)
}

func TestAuthenticationToken_InvalidEmail_ReturnsUnauthorized(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	authData := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "SomePassword123!",
	}

	jsonData, err := json.Marshal(authData)
	require.NoError(t, err)

	resp, err := http.Post(server_addr+"/v1/auth/token", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid credentials", errorResp.Error)
}

func TestAuthenticationToken_InvalidPassword_ReturnsUnauthorized(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	_ = createTestUser(t, queries, datastore)

	authData := map[string]string{
		"email":    "test@example.com",
		"password": "WrongPassword123!",
	}

	jsonData, err := json.Marshal(authData)
	require.NoError(t, err)

	resp, err := http.Post(server_addr+"/v1/auth/token", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid credentials", errorResp.Error)
}

func TestAuthenticationToken_MissingFields_ReturnsValidationError(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	tests := []struct {
		name     string
		authData map[string]string
		field    string
		error    string
	}{
		{
			name:     "missing email",
			authData: map[string]string{"password": "ValidPassword123!"},
			field:    "email",
			error:    "email is required",
		},
		{
			name:     "missing password",
			authData: map[string]string{"email": "user@example.com"},
			field:    "password",
			error:    "password is required",
		},
		{
			name:     "empty email",
			authData: map[string]string{"email": "", "password": "ValidPassword123!"},
			field:    "email",
			error:    "email is required",
		},
		{
			name:     "empty password",
			authData: map[string]string{"email": "user@example.com", "password": ""},
			field:    "password",
			error:    "password is required",
		},
		{
			name:     "invalid email format",
			authData: map[string]string{"email": "invalid-email", "password": "ValidPassword123!"},
			field:    "email",
			error:    "invalid email format",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.authData)
			require.NoError(t, err)

			resp, err := http.Post(server_addr+"/v1/auth/token", "application/json", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var validationResp struct {
				Error            string              `json:"error"`
				ValidationErrors map[string][]string `json:"validation_errors"`
			}
			err = json.Unmarshal(body, &validationResp)
			require.NoError(t, err)

			assert.Contains(t, validationResp.Error, "validation failed")
			assert.NotEmpty(t, validationResp.ValidationErrors, "Expected validation_errors to be present")

			fieldErrors, exists := validationResp.ValidationErrors[tc.field]
			require.True(t, exists, "Expected validation error for field '%s'", tc.field)
			require.NotEmpty(t, fieldErrors, "Expected validation errors for field '%s' to not be empty", tc.field)

			// Check if any error message contains the expected text
			found := false
			for _, errMsg := range fieldErrors {
				if assert.Contains(t, errMsg, tc.error) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected field '%s' to have error containing '%s', got: %v", tc.field, tc.error, fieldErrors)
		})
	}
}

func TestAuthenticationToken_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	// Send malformed JSON
	malformedJSON := `{"email": "test@example.com", "password": "incomplete`

	resp, err := http.Post(server_addr+"/v1/auth/token", "application/json", bytes.NewBufferString(malformedJSON))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Error, "body contains badly-formed JSON")
}

func TestAuthenticationToken_EmptyBody_ReturnsBadRequest(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	// Send empty body
	resp, err := http.Post(server_addr+"/v1/auth/token", "application/json", bytes.NewBuffer(nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp.Error, "body must not be empty")
}
