package ports

import (
	"context"
	"personal_website/internal/app/core/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (int, error)
	ActivateUser(ctx context.Context, user *domain.User) error
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	DeactivateUser(ctx context.Context, id int) error
	DeleteUser(ctx context.Context, id int) error
}
