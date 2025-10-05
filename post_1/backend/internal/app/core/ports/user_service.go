package ports

import (
	"context"
	"personal_website/internal/app/core/domain"
)

// UserService encapsulates business logic for user operations with transaction support
type UserService interface {
	// RegisterUser handles the complete user registration process:
	// - Creates user in database
	// - Generates activation token
	// - Stores activation token
	// - Sends activation email
	RegisterUser(ctx context.Context, user domain.User, activationURL string) error

	// ActivateUser handles the complete user activation process:
	// - Validates activation token
	// - Activates user account
	// - Deletes all activation tokens for the user
	// - Sends new user notification email to admin
	ActivateUser(ctx context.Context, tokenPlaintext string) (*domain.User, error)
}
