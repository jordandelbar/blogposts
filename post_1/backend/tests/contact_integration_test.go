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

func TestContactHandler_Success(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	contactData := ContactData()

	jsonData, err := json.Marshal(contactData)
	require.NoError(t, err)

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the JSON success response
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var successResp struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(body, &successResp)
	require.NoError(t, err)
	assert.Equal(t, "Message sent successfully!", successResp.Message)

	// Verify email was sent by checking mock calls
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	require.Len(t, calls, 1, "Expected exactly one email to be sent")

	call := calls[0]
	assert.Equal(t, "localhost:587", call["host"])
	assert.Equal(t, "", call["from"])
	assert.Equal(t, []string{"test@example.com"}, call["to"])

	// Verify email content
	msgContent := string(call["msg"].([]byte))
	assert.Contains(t, msgContent, "Content-Type: text/html")
	assert.Contains(t, msgContent, "John Doe")
	assert.Contains(t, msgContent, "john.doe@example.com")
	assert.Contains(t, msgContent, "This is a test message from the contact form")
	assert.Contains(t, msgContent, "Subject: New Contact Form Submission from John Doe")
}

func TestContactHandler_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	// Send invalid JSON
	invalidJSON := `{"name": "John", "email": "invalid-json"`

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBufferString(invalidJSON))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for invalid JSON")
}

func TestContactHandler_MissingName_ReturnsValidationError(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	contactData := map[string]string{
		"email":   "john.doe@example.com",
		"message": "This is a test message from the contact form.",
	}

	jsonData, err := json.Marshal(contactData)
	require.NoError(t, err)

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	suite.AssertValidationError(t, resp, "validation failed", map[string]string{
		"name": "name is required",
	})

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for validation error")
}

func TestContactHandler_InvalidEmail_ReturnsValidationError(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	contactData := map[string]string{
		"name":    "John Doe",
		"email":   "invalid-email-format",
		"message": "This is a test message from the contact form.",
	}

	jsonData, err := json.Marshal(contactData)
	require.NoError(t, err)

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	suite.AssertValidationError(t, resp, "validation failed", map[string]string{
		"email": "invalid email format",
	})

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for validation error")
}

func TestContactHandler_MessageTooShort_ReturnsValidationError(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	contactData := map[string]string{
		"name":    "John Doe",
		"email":   "john.doe@example.com",
		"message": "Short",
	}

	jsonData, err := json.Marshal(contactData)
	require.NoError(t, err)

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	suite.AssertValidationError(t, resp, "validation failed", map[string]string{
		"message": "message must be at least 10 characters",
	})

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for validation error")
}

func TestContactHandler_ScriptTagRejected_ReturnsValidationError(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	contactData := map[string]string{
		"name":    "John Doe",
		"email":   "john.doe@example.com",
		"message": "This message contains a <script>alert('XSS')</script> tag for testing.",
	}

	jsonData, err := json.Marshal(contactData)
	require.NoError(t, err)

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	suite.AssertValidationError(t, resp, "validation failed", map[string]string{
		"message": "message cannot contain script tags",
	})

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for script injection attempt")
}

func TestContactHandler_HTMLTagInName_ReturnsValidationError(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	contactData := map[string]string{
		"name":    "John <b>Doe</b>",
		"email":   "john.doe@example.com",
		"message": "This is a valid message without any HTML tags.",
	}

	jsonData, err := json.Marshal(contactData)
	require.NoError(t, err)

	resp, err := http.Post(suite.ServerAddr+"/v1/contact", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	suite.AssertValidationError(t, resp, "validation failed", map[string]string{
		"name": "name cannot contain HTML tags",
	})

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for HTML injection attempt")
}

func TestContactHandler_WrongHTTPMethod_ReturnsMethodNotAllowed(t *testing.T) {
	suite := NewUnauthenticatedTestSuite(t)

	// Reset mock email sender calls
	mockSender := GetMockEmailSender()
	mockSender.mu.Lock()
	mockSender.Calls = nil
	mockSender.mu.Unlock()

	// Try to GET the contact endpoint (should only accept POST)
	resp, err := http.Get(suite.ServerAddr + "/v1/contact")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	// Verify no email was sent
	mockSender.mu.Lock()
	calls := mockSender.Calls
	mockSender.mu.Unlock()

	assert.Len(t, calls, 0, "No email should have been sent for method not allowed")
}
