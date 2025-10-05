package ports

import (
	"context"
	"personal_website/internal/app/core/domain"
)

type TokenRepository interface {
	CreateToken(ctx context.Context, token *domain.Token) error
	DeleteAllTokenForUser(ctx context.Context, token *domain.Token) error
}
