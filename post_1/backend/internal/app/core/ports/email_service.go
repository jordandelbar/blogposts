package ports

import (
	"context"
	"net/smtp"
	"personal_website/internal/app/core/domain"
)

type EmailService interface {
	SendContactEmail(ctx context.Context, form domain.ContactMessage) error
	SendActivationEmail(ctx context.Context, activationToken string, recipientEmail string, baseURL string) error
	SendNewUserNotification(ctx context.Context, user *domain.User) error
}

type EmailSender interface {
	SendMail(host string, auth smtp.Auth, from string, to []string, msg []byte) error
}
