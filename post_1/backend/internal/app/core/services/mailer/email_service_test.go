package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"personal_website/config"
	"personal_website/internal/app/core/domain"
	"sync"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
)

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

func TestSendContactEmail(t *testing.T) {
	mockSender := &MockEmailSender{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sender, _ := NewService(&config.SMTPConfig{
		Host:      memguard.NewBufferFromBytes([]byte("smtp.example.com")),
		Port:      memguard.NewBufferFromBytes([]byte("587")),
		Username:  memguard.NewBufferFromBytes([]byte("username")),
		Password:  memguard.NewBufferFromBytes([]byte("password")),
		Recipient: memguard.NewBufferFromBytes([]byte("recipient@example.com")),
	}, mockSender, logger)

	form := domain.ContactMessage{
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Message: "This is a test message.",
	}

	err := sender.SendContactEmail(context.Background(), form)
	assert.NoError(t, err)

	assert.Len(t, mockSender.Calls, 1)
	call := mockSender.Calls[0]
	assert.Equal(t, "smtp.example.com:587", call["host"])
	assert.Equal(t, "", call["from"])
	assert.Equal(t, []string{"recipient@example.com"}, call["to"])

	msgContent := string(call["msg"].([]byte))
	assert.Contains(t, msgContent, "Content-Type: text/html")
	assert.Contains(t, msgContent, "John Doe")
	assert.Contains(t, msgContent, "john.doe@example.com")
	assert.Contains(t, msgContent, "This is a test message.")
}

func TestSendActivationEmail(t *testing.T) {
	mockSender := &MockEmailSender{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	sender, _ := NewService(&config.SMTPConfig{
		Host:     memguard.NewBufferFromBytes([]byte("smtp.example.com")),
		Port:     memguard.NewBufferFromBytes([]byte("587")),
		Username: memguard.NewBufferFromBytes([]byte("username")),
		Password: memguard.NewBufferFromBytes([]byte("password")),
	}, mockSender, logger)

	activationToken := "test-activation-token-123"
	recipientEmail := "user@example.com"
	baseURL := "http://localhost:8080"

	err := sender.SendActivationEmail(context.Background(), activationToken, recipientEmail, baseURL)
	assert.NoError(t, err)

	assert.Len(t, mockSender.Calls, 1)
	call := mockSender.Calls[0]
	assert.Equal(t, "smtp.example.com:587", call["host"])
	assert.Equal(t, "", call["from"])
	assert.Equal(t, []string{recipientEmail}, call["to"])

	msgContent := string(call["msg"].([]byte))
	assert.Contains(t, msgContent, "Subject: Activate Your Account")
	assert.Contains(t, msgContent, "Content-Type: text/html")
	expectedURL := fmt.Sprintf("%s?token=%s", baseURL, activationToken)
	assert.Contains(t, msgContent, expectedURL)
}
