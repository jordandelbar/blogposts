package domain

// ContactMessage defines the data structure needed by the mailer service.
type ContactMessage struct {
	Name    string
	Email   string
	Message string
}
