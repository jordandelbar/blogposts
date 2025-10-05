package ports

import (
	"context"
	"personal_website/internal/app/core/domain"
)

type SessionRepository interface {
	StoreSession(ctx context.Context, token string, scope domain.TokenScope, session *domain.Session) error
	GetSession(ctx context.Context, token string, scope domain.TokenScope) (*domain.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteAllSessionsForUser(ctx context.Context, userID int, scope domain.TokenScope) error
}
