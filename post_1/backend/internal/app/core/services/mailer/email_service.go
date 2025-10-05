package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"personal_website/config"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/app/core/ports"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

type EmailService struct {
	emailSender ports.EmailSender
	config      *config.SMTPConfig
	logger      *slog.Logger
	templates   *template.Template
}

func NewService(cfg *config.SMTPConfig, emailSender ports.EmailSender, logger *slog.Logger) (*EmailService, error) {
	templates, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		logger.Error("Failed to parse email templates", "error", err)
		return &EmailService{}, err
	}

	return &EmailService{
		config:      cfg,
		emailSender: emailSender,
		logger:      logger,
		templates:   templates,
	}, nil
}

type smtpConfig struct {
	host           string
	port           string
	username       string
	password       string
	serverAddr     string
	verifiedSender string
}

func (s *EmailService) setupSMTP() (*smtpConfig, error) {
	smtpHost := s.config.Host.String()
	smtpPort := s.config.Port.String()
	username := s.config.Username.String()
	password := s.config.Password.String()

	if smtpHost == "" || smtpPort == "" || username == "" || password == "" {
		return nil, fmt.Errorf("missing SMTP configuration: host=%s, port=%s, username=%s",
			smtpHost, smtpPort, username)
	}

	return &smtpConfig{
		host:           smtpHost,
		port:           smtpPort,
		username:       username,
		password:       password,
		serverAddr:     smtpHost + ":" + smtpPort,
		verifiedSender: "",
	}, nil
}

func (s *EmailService) buildHTMLEmail(to, subject, htmlContent string) []byte {
	return []byte("From: " + "" + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		htmlContent)
}

func (s *EmailService) SendContactEmail(ctx context.Context, form domain.ContactMessage) error {
	smtpCfg, err := s.setupSMTP()
	if err != nil {
		s.logger.Error("SMTP configuration incomplete", "error", err)
		return domain.ErrEmailConfigurationMissing
	}

	recipient := s.config.Recipient.String()
	if recipient == "" {
		err := fmt.Errorf("missing recipient configuration")
		s.logger.Error("Recipient configuration incomplete", "error", err)
		return domain.ErrEmailConfigurationMissing
	}

	s.logger.Info("Preparing to send contact email",
		"server", smtpCfg.serverAddr,
		"recipient", recipient,
		"sender_email", form.Email,
	)

	templateData := struct {
		Name      string
		Email     string
		Message   string
		Timestamp string
	}{
		Name:      form.Name,
		Email:     form.Email,
		Message:   form.Message,
		Timestamp: time.Now().Format("2006-01-02 15:04:05 MST"),
	}

	var htmlBuffer bytes.Buffer
	if s.templates != nil {
		err = s.templates.ExecuteTemplate(&htmlBuffer, "contact.html", templateData)
		if err != nil {
			s.logger.Error("Failed to execute contact template", "error", err)
			return domain.ErrEmailTemplateFailed
		}
	} else {
		return domain.ErrEmailTemplateFailed
	}

	subject := fmt.Sprintf("New Contact Form Submission from %s", form.Name)
	msg := s.buildHTMLEmail(recipient, subject, htmlBuffer.String())

	auth := smtp.PlainAuth("", smtpCfg.username, smtpCfg.password, smtpCfg.host)
	s.logger.Info("Attempting to send contact email", "server", smtpCfg.serverAddr, "to", recipient)

	err = s.emailSender.SendMail(smtpCfg.serverAddr, auth, smtpCfg.verifiedSender, []string{recipient}, msg)
	if err != nil {
		s.logger.Error("Failed to send contact email via SMTP",
			"error", err,
			"server", smtpCfg.serverAddr,
			"from", smtpCfg.verifiedSender,
			"reply_to", form.Email,
			"recipient", recipient,
		)
		return domain.ErrEmailSendFailed
	}

	s.logger.Info("Contact email sent successfully", "recipient", recipient, "from", smtpCfg.verifiedSender, "reply_to", form.Email)
	return nil
}

func (s *EmailService) SendActivationEmail(ctx context.Context, activationToken string, recipientEmail string, baseURL string) error {
	smtpCfg, err := s.setupSMTP()
	if err != nil {
		s.logger.Error("SMTP configuration incomplete", "error", err)
		return domain.ErrEmailConfigurationMissing
	}

	s.logger.Info("Preparing to send activation email",
		"server", smtpCfg.serverAddr,
		"recipient", recipientEmail,
		"base_url", baseURL,
	)

	activationURL := fmt.Sprintf("%s?token=%s", baseURL, activationToken)

	templateData := struct {
		Token         string
		Email         string
		ActivationURL string
	}{
		Token:         activationToken,
		Email:         recipientEmail,
		ActivationURL: activationURL,
	}

	var htmlBuffer bytes.Buffer
	if s.templates != nil {
		err = s.templates.ExecuteTemplate(&htmlBuffer, "activation.html", templateData)
		if err != nil {
			s.logger.Error("Failed to execute activation template", "error", err)
			return domain.ErrEmailTemplateFailed
		}
	} else {
		return domain.ErrEmailTemplateFailed
	}

	subject := "Activate Your Account"
	msg := s.buildHTMLEmail(recipientEmail, subject, htmlBuffer.String())

	auth := smtp.PlainAuth("", smtpCfg.username, smtpCfg.password, smtpCfg.host)
	s.logger.Info("Attempting to send activation email", "server", smtpCfg.serverAddr, "to", recipientEmail)

	err = s.emailSender.SendMail(smtpCfg.serverAddr, auth, smtpCfg.verifiedSender, []string{recipientEmail}, msg)
	if err != nil {
		s.logger.Error("Failed to send activation email via SMTP",
			"error", err,
			"server", smtpCfg.serverAddr,
			"from", smtpCfg.verifiedSender,
			"recipient", recipientEmail,
		)
		return domain.ErrEmailSendFailed
	}

	s.logger.Info("Activation email sent successfully", "recipient", recipientEmail, "from", smtpCfg.verifiedSender, "activation_url", activationURL)
	return nil
}

func (s *EmailService) SendNewUserNotification(ctx context.Context, user *domain.User) error {
	smtpCfg, err := s.setupSMTP()
	if err != nil {
		s.logger.Error("SMTP configuration incomplete", "error", err)
		return domain.ErrEmailConfigurationMissing
	}

	recipient := s.config.Recipient.String()
	if recipient == "" {
		err := fmt.Errorf("missing recipient configuration")
		s.logger.Error("Recipient configuration incomplete", "error", err)
		return domain.ErrEmailConfigurationMissing
	}

	s.logger.Info("Preparing to send new user notification email",
		"server", smtpCfg.serverAddr,
		"recipient", recipient,
		"new_user_email", user.Email,
	)

	templateData := struct {
		Username         string
		Email            string
		RegistrationDate string
		NotificationTime string
	}{
		Username:         user.Name,
		Email:            user.Email,
		RegistrationDate: user.CreatedAt.Format("2006-01-02 15:04:05 MST"),
		NotificationTime: time.Now().Format("2006-01-02 15:04:05 MST"),
	}

	var htmlBuffer bytes.Buffer
	if s.templates != nil {
		err = s.templates.ExecuteTemplate(&htmlBuffer, "new_user.html", templateData)
		if err != nil {
			s.logger.Error("Failed to execute new user template", "error", err)
			return domain.ErrEmailTemplateFailed
		}
	} else {
		return domain.ErrEmailTemplateFailed
	}

	subject := fmt.Sprintf("New User Activated: %s", user.Name)
	msg := s.buildHTMLEmail(recipient, subject, htmlBuffer.String())

	auth := smtp.PlainAuth("", smtpCfg.username, smtpCfg.password, smtpCfg.host)
	s.logger.Info("Attempting to send new user notification email", "server", smtpCfg.serverAddr, "to", recipient)

	err = s.emailSender.SendMail(smtpCfg.serverAddr, auth, smtpCfg.verifiedSender, []string{recipient}, msg)
	if err != nil {
		s.logger.Error("Failed to send new user notification email via SMTP",
			"error", err,
			"server", smtpCfg.serverAddr,
			"from", smtpCfg.verifiedSender,
			"recipient", recipient,
		)
		return domain.ErrEmailSendFailed
	}

	s.logger.Info("New user notification email sent successfully", "recipient", recipient, "from", smtpCfg.verifiedSender, "new_user", user.Email)
	return nil
}
