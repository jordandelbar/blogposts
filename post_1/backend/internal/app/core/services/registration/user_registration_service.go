package registration

import (
	"context"
	"fmt"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/app/core/ports"
)

type userService struct {
	emailService ports.EmailService
	datastore    ports.Datastore
}

func NewUserService(emailService ports.EmailService, datastore ports.Datastore) *userService {
	return &userService{
		emailService: emailService,
		datastore:    datastore,
	}
}

func (u *userService) RegisterUser(ctx context.Context, user domain.User, activationURL string) error {
	tx, err := u.datastore.Begin(ctx)
	if err != nil {
		return domain.NewInternalError(err)
	}
	defer tx.Rollback()

	userID, err := tx.UserRepo().CreateUser(ctx, user)
	if err != nil {
		return err
	}

	token := domain.GenerateToken(userID, domain.ScopeActivation)

	session := &domain.Session{
		UserID:      userID,
		Email:       user.Email,
		Permissions: domain.Permissions{},
		Activated:   false,
	}

	if err := u.datastore.SessionRepo().StoreSession(ctx, token.Plaintext, domain.ScopeActivation, session); err != nil {
		return domain.NewInternalError(err)
	}

	if err := tx.Commit(); err != nil {
		return domain.NewInternalError(err)
	}

	if err := u.emailService.SendActivationEmail(ctx, token.Plaintext, user.Email, activationURL); err != nil {
		return err
	}

	return nil
}

func (u *userService) ActivateUser(ctx context.Context, tokenPlaintext string) (*domain.User, error) {
	tx, err := u.datastore.Begin(ctx)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}
	defer tx.Rollback()

	// Get session to verify activation token and get user ID
	session, err := u.datastore.SessionRepo().GetSession(ctx, tokenPlaintext, domain.ScopeActivation)
	if err != nil {
		return nil, err
	}

	// Get user by email (we need the full user object for activation and email notification)
	user, err := tx.UserRepo().GetUserByEmail(ctx, session.Email)
	if err != nil {
		return nil, err
	}

	if err := tx.UserRepo().ActivateUser(ctx, &user); err != nil {
		return nil, err
	}

	// Delete all activation sessions for this user
	if err := u.datastore.SessionRepo().DeleteAllSessionsForUser(ctx, session.UserID, domain.ScopeActivation); err != nil {
		return nil, domain.NewInternalError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, domain.NewInternalError(err)
	}

	if err := u.emailService.SendNewUserNotification(ctx, &user); err != nil {
		fmt.Printf("user activated but failed to send notification email: %v\n", err)
	}

	return &user, nil
}
