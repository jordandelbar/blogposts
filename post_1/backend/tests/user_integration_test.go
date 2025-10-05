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

func TestCreateUser(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)
	articleData := map[string]string{
		"name":     "John Doe",
		"email":    "john.doe@email.com",
		"password": "Pa55word!",
	}

	jsonData, err := json.Marshal(articleData)
	require.NoError(t, err)

	resp, err := http.Post(server_addr+"/v1/users", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestCreateUser_DuplicateEmail_ReturnsConflict(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)
	userData := map[string]string{
		"name":     "John Doe",
		"email":    "duplicate@example.com",
		"password": "Pa55word!",
	}

	jsonData, err := json.Marshal(userData)
	require.NoError(t, err)

	// Create user first time - should succeed
	resp, err := http.Post(server_addr+"/v1/users", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	// Try to create the same user again - should return conflict
	jsonData, err = json.Marshal(userData) // Re-marshal for fresh buffer
	require.NoError(t, err)

	resp2, err := http.Post(server_addr+"/v1/users", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	// Verify the JSON error response
	body, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "user already exists", errorResp.Error)
}

func TestActivateUser_InvalidToken_ReturnsNotFound(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() { cleanupDB(t) })

	server_addr := startTestServer(t)

	// Try to activate user with non-existent/invalid token
	activationData := map[string]string{
		"token": "INVALIDTOKEN123456789012345678", // 26 characters but non-existent
	}

	jsonData, err := json.Marshal(activationData)
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH", server_addr+"/v1/users/activate", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Verify the JSON error response
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "session not found", errorResp.Error)
}
