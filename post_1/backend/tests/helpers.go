package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/smtp"
	"personal_website/config"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/app/core/ports"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
	"sync"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, queries *sqlc.Queries, datastore ports.Datastore) string {
	ctx := context.Background()

	// Create user with hashed password directly in database
	var user domain.User
	user.Name = "Test User"
	user.Email = "test@example.com"
	err := user.Password.Set("TestPassword123!")
	require.NoError(t, err)

	userID, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.Password.Hash(),
	})
	require.NoError(t, err)

	// Activate the user
	err = queries.ActivateUser(ctx, userID)
	require.NoError(t, err)

	// Grant articles:read and articles:write permissions to the user
	err = queries.AddPermissionForUser(ctx, sqlc.AddPermissionForUserParams{
		UserID: int64(userID),
		Code:   "articles:read",
	})
	require.NoError(t, err)

	err = queries.AddPermissionForUser(ctx, sqlc.AddPermissionForUserParams{
		UserID: int64(userID),
		Code:   "articles:write",
	})
	require.NoError(t, err)

	// Create session in valkey instead of token in database
	token := domain.GenerateToken(int(userID), domain.ScopeAuthentication)

	// Create session with user permissions
	session := &domain.Session{
		UserID:      int(userID),
		Email:       user.Email,
		Permissions: domain.Permissions{"articles:read", "articles:write"},
		Activated:   true,
	}

	// Store session in valkey
	err = datastore.SessionRepo().StoreSession(ctx, token.Plaintext, domain.ScopeAuthentication, session)
	require.NoError(t, err)

	return token.Plaintext
}

func NewTestConfig(dsnParts map[string]string, port int) *config.Config {
	return &config.Config{
		Postgres: config.PostgresConfig{
			User:     memguard.NewBufferFromBytes([]byte(dsnParts["pg_user"])),
			Password: memguard.NewBufferFromBytes([]byte(dsnParts["pg_password"])),
			Database: memguard.NewBufferFromBytes([]byte(dsnParts["pg_dbname"])),
			Host:     memguard.NewBufferFromBytes([]byte(dsnParts["pg_host"])),
			Port:     memguard.NewBufferFromBytes([]byte(dsnParts["pg_port"])),
		},
		Valkey: config.ValkeyConfig{
			Host: memguard.NewBufferFromBytes([]byte(dsnParts["valkey_host"])),
			Port: memguard.NewBufferFromBytes([]byte(dsnParts["valkey_port"])),
		},
		Minio: config.MinioConfig{
			Endpoint:  memguard.NewBufferFromBytes([]byte("localhost:9000")),
			AccessKey: memguard.NewBufferFromBytes([]byte("test-access-key")),
			SecretKey: memguard.NewBufferFromBytes([]byte("test-secret-key")),
			Bucket:    memguard.NewBufferFromBytes([]byte("test-bucket")),
			UseSSL:    false,
		},
		SMTP: config.SMTPConfig{
			Host:      memguard.NewBufferFromBytes([]byte("localhost")),
			Port:      memguard.NewBufferFromBytes([]byte("587")),
			Username:  memguard.NewBufferFromBytes([]byte("test@example.com")),
			Password:  memguard.NewBufferFromBytes([]byte("testpass")),
			Recipient: memguard.NewBufferFromBytes([]byte("test@example.com")),
		},
		App: config.AppConfig{
			Environment: "test",
			Version:     "test",
			Port:        port,
			Cors: config.CORSConfig{
				TrustedOrigins: []string{"http://localhost:3000", "https://example.com"},
			},
			Limiter: config.LimiterConfig{
				Rps:     5,  // Low limit for easy testing
				Burst:   10, // Low burst for easy testing
				Enabled: true,
			},
		},
	}
}

type MockEmailSender struct {
	mu    sync.Mutex
	Calls []map[string]interface{}
}

func (m *MockEmailSender) SendMail(host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, map[string]interface{}{
		"host": host,
		"auth": auth,
		"from": from,
		"to":   to,
		"msg":  msg,
	})
	return nil
}

// TestSuite encapsulates common test setup
type TestSuite struct {
	ServerAddr string
	AuthToken  string
}

// ArticleData creates test article data
func ArticleData() map[string]string {
	return map[string]string{
		"title":   "Test Article",
		"slug":    "test-article",
		"content": "This is a test article content. This must be at least 50 character or it won't pass validation",
	}
}

// UserData creates test user data
func UserData() map[string]string {
	return map[string]string{
		"name":     "John Doe",
		"email":    "john.doe@email.com",
		"password": "Pa55word!",
	}
}

// ContactData creates test contact form data
func ContactData() map[string]string {
	return map[string]string{
		"name":    "John Doe",
		"email":   "john.doe@example.com",
		"message": "This is a test message from the contact form. It needs to be at least 10 characters long.",
	}
}

// Global mock email sender instance
var testMockEmailSender *MockEmailSender

// GetMockEmailSender returns the global mock email sender instance
func GetMockEmailSender() *MockEmailSender {
	return testMockEmailSender
}

// MockResumeService implements ports.ResumeService for testing
type MockResumeService struct {
	mu              sync.Mutex
	GetResumeCalls  []string
	CheckConnCalls  []bool
	ShouldFailGet   bool
	ShouldFailCheck bool
	GetResumeData   []byte
	GetResumeError  error
	CheckConnError  error
}

func NewMockResumeService() *MockResumeService {
	return &MockResumeService{
		GetResumeData: []byte("mock pdf data"),
	}
}

func (m *MockResumeService) GetResume(ctx context.Context, objectName string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetResumeCalls = append(m.GetResumeCalls, objectName)

	if m.ShouldFailGet {
		return nil, m.GetResumeError
	}

	return m.GetResumeData, nil
}

func (m *MockResumeService) CheckConnection(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CheckConnCalls = append(m.CheckConnCalls, true)

	if m.ShouldFailCheck {
		return m.CheckConnError
	}

	return nil
}

func (m *MockResumeService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetResumeCalls = []string{}
	m.CheckConnCalls = []bool{}
	m.ShouldFailGet = false
	m.ShouldFailCheck = false
	m.GetResumeError = nil
	m.CheckConnError = nil
}

// Global mock resume service instance
var testMockResumeService *MockResumeService

// GetMockResumeService returns the global mock resume service instance
func GetMockResumeService() *MockResumeService {
	return testMockResumeService
}

// POST makes authenticated POST request
func (ts *TestSuite) POST(t *testing.T, path string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return NewRequestWithAuthentication(t, "POST", ts.ServerAddr+path, ts.AuthToken, jsonData)
}

// GET makes authenticated GET request
func (ts *TestSuite) GET(t *testing.T, path string) (*http.Response, error) {
	return NewRequestWithAuthentication(t, "GET", ts.ServerAddr+path, ts.AuthToken, nil)
}

// DELETE makes authenticated DELETE request
func (ts *TestSuite) DELETE(t *testing.T, path string) (*http.Response, error) {
	return NewRequestWithAuthentication(t, "DELETE", ts.ServerAddr+path, ts.AuthToken, nil)
}

func NewRequestWithAuthentication(t *testing.T, method string, route string, authToken string, payload []byte) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewBuffer(payload)
	}

	req, err := http.NewRequest(method, route, body)
	require.NoError(t, err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// AssertJSONError parses JSON error response and checks message
func (ts *TestSuite) AssertJSONError(t *testing.T, resp *http.Response, expectedMsg string) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var errorResp struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp.Error, expectedMsg)
}

// AssertValidationError parses JSON validation error response and checks message and field errors
func (ts *TestSuite) AssertValidationError(t *testing.T, resp *http.Response, expectedMsg string, expectedFieldErrors ...map[string]string) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var validationResp struct {
		Error            string              `json:"error"`
		ValidationErrors map[string][]string `json:"validation_errors"`
	}
	err = json.Unmarshal(body, &validationResp)
	require.NoError(t, err)

	assert.Contains(t, validationResp.Error, expectedMsg)
	assert.NotEmpty(t, validationResp.ValidationErrors, "Expected validation_errors to be present")

	// If specific field errors are provided, check them
	if len(expectedFieldErrors) > 0 {
		expectedFields := expectedFieldErrors[0]
		for field, expectedError := range expectedFields {
			fieldErrors, exists := validationResp.ValidationErrors[field]
			require.True(t, exists, "Expected validation error for field '%s'", field)
			require.NotEmpty(t, fieldErrors, "Expected validation errors for field '%s' to not be empty", field)

			// Check if any error message contains the expected text
			found := false
			for _, errMsg := range fieldErrors {
				if assert.ObjectsAreEqual(expectedError, errMsg) ||
					(expectedError != "" && assert.Contains(t, errMsg, expectedError)) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected field '%s' to have error containing '%s', got: %v", field, expectedError, fieldErrors)
		}
	}
}
