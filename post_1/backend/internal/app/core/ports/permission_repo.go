package ports

import (
	"context"
	"personal_website/internal/app/core/domain"
)

type PermissionRepository interface {
	GetPermissions(ctx context.Context, user *domain.User) (domain.Permissions, error)
}
