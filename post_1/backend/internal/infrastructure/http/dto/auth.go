package dto

import "time"

type ActivationToken struct {
	TokenPlaintext string `json:"token" validate:"required"`
}

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Success     bool      `json:"success"`
	AccessToken string    `json:"access_token"`
	Expiry      time.Time `json:"expiry"`
}

type RefreshTokenResponse struct {
	AccessToken string    `json:"access_token"`
	Expiry      time.Time `json:"expiry"`
}

type AuthStatusResponse struct {
	Authenticated bool       `json:"authenticated"`
	Expiry        *time.Time `json:"expiry,omitempty"`
	UserEmail     *string    `json:"user_email,omitempty"`
}
