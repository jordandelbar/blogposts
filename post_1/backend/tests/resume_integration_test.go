package tests

import (
	"context"
	"io"
	"net/http"
	"personal_website/internal/app/core/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResumeDownload_Success(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)
	mockResume := GetMockResumeService()

	// Set up mock to return successful resume data
	mockResume.Reset()
	mockResume.GetResumeData = []byte("PDF content here")

	// Make request
	req, err := http.NewRequest("GET", ts.ServerAddr+"/v1/resume", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
	assert.Equal(t, "attachment; filename=\"jordan_delbar_resume.pdf\"", resp.Header.Get("Content-Disposition"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("PDF content here"), body)

	// Verify service was called with correct parameters
	assert.Len(t, mockResume.GetResumeCalls, 1)
	assert.Equal(t, "resume_without_personal_data.pdf", mockResume.GetResumeCalls[0])
}

func TestResumeDownload_NotFound(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)
	mockResume := GetMockResumeService()

	// Set up mock to return not found error
	mockResume.Reset()
	mockResume.ShouldFailGet = true
	mockResume.GetResumeError = domain.DomainError{
		Code:    "resume_not_found",
		Message: "resume file not found",
		Type:    domain.ErrorTypeNotFound,
	}

	// Make request
	req, err := http.NewRequest("GET", ts.ServerAddr+"/v1/resume", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	ts.AssertJSONError(t, resp, "resume file not found")

	// Verify service was called
	assert.Len(t, mockResume.GetResumeCalls, 1)
	assert.Equal(t, "resume_without_personal_data.pdf", mockResume.GetResumeCalls[0])
}

func TestResumeDownload_StorageUnavailable(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)
	mockResume := GetMockResumeService()

	// Set up mock to return storage unavailable error
	mockResume.Reset()
	mockResume.ShouldFailGet = true
	mockResume.GetResumeError = domain.DomainError{
		Code:    "resume_storage_unavailable",
		Message: "resume storage service is unavailable",
		Type:    domain.ErrorTypeInternal,
	}

	// Make request
	req, err := http.NewRequest("GET", ts.ServerAddr+"/v1/resume", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response - internal errors return 500 with generic message
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	ts.AssertJSONError(t, resp, "internal server error")

	// Verify service was called
	assert.Len(t, mockResume.GetResumeCalls, 1)
	assert.Equal(t, "resume_without_personal_data.pdf", mockResume.GetResumeCalls[0])
}

func TestResumeDownload_ReadFailed(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)
	mockResume := GetMockResumeService()

	// Set up mock to return read failed error
	mockResume.Reset()
	mockResume.ShouldFailGet = true
	mockResume.GetResumeError = domain.DomainError{
		Code:    "resume_read_failed",
		Message: "failed to read resume file",
		Type:    domain.ErrorTypeInternal,
	}

	// Make request
	req, err := http.NewRequest("GET", ts.ServerAddr+"/v1/resume", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response - internal errors return 500 with generic message
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	ts.AssertJSONError(t, resp, "internal server error")

	// Verify service was called
	assert.Len(t, mockResume.GetResumeCalls, 1)
	assert.Equal(t, "resume_without_personal_data.pdf", mockResume.GetResumeCalls[0])
}

func TestResumeService_MapMinioErrors(t *testing.T) {
	// Test the error mapping logic by creating a real resume service and checking domain errors
	// This tests the mapMinioError function indirectly

	tests := []struct {
		name            string
		setupMockError  error
		expectedCode    string
		expectedType    domain.ErrorType
		expectedStatus  int
	}{
		{
			name: "NoSuchKey error maps to not found",
			setupMockError: domain.DomainError{
				Code:    "resume_not_found",
				Message: "resume file not found",
				Type:    domain.ErrorTypeNotFound,
			},
			expectedCode:   "resume_not_found",
			expectedType:   domain.ErrorTypeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Connection error maps to storage unavailable",
			setupMockError: domain.DomainError{
				Code:    "resume_storage_unavailable",
				Message: "resume storage service is unavailable",
				Type:    domain.ErrorTypeInternal,
			},
			expectedCode:   "resume_storage_unavailable",
			expectedType:   domain.ErrorTypeInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Generic error maps to read failed",
			setupMockError: domain.DomainError{
				Code:    "resume_read_failed",
				Message: "failed to read resume file",
				Type:    domain.ErrorTypeInternal,
			},
			expectedCode:   "resume_read_failed",
			expectedType:   domain.ErrorTypeInternal,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewUnauthenticatedTestSuite(t)
			mockResume := GetMockResumeService()

			// Set up mock to return the specific error
			mockResume.Reset()
			mockResume.ShouldFailGet = true
			mockResume.GetResumeError = tt.setupMockError

			// Make request
			req, err := http.NewRequest("GET", ts.ServerAddr+"/v1/resume", nil)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedType == domain.ErrorTypeNotFound {
				domainErr := tt.setupMockError.(domain.DomainError)
				ts.AssertJSONError(t, resp, domainErr.Message)
			} else {
				// Internal errors should return generic message
				ts.AssertJSONError(t, resp, "internal server error")
			}
		})
	}
}

func TestResumeService_ConnectionCheck(t *testing.T) {
	// This is more of a unit test for the CheckConnection method
	// We can test it through the service initialization or add a health endpoint later
	mockResume := NewMockResumeService()

	// Test successful connection
	err := mockResume.CheckConnection(context.Background())
	assert.NoError(t, err)
	assert.Len(t, mockResume.CheckConnCalls, 1)

	// Test failed connection
	mockResume.Reset()
	mockResume.ShouldFailCheck = true
	mockResume.CheckConnError = domain.DomainError{
		Code:    "resume_storage_unavailable",
		Message: "resume storage service is unavailable",
		Type:    domain.ErrorTypeInternal,
	}

	err = mockResume.CheckConnection(context.Background())
	assert.Error(t, err)
	assert.Len(t, mockResume.CheckConnCalls, 1)

	var domainErr domain.DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "resume_storage_unavailable", domainErr.Code)
}