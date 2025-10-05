package email_sender

import "net/smtp"

type EmailSender struct{}

func (e *EmailSender) SendMail(host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(host, auth, from, to, msg)
}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}
