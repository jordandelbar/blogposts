package postgres_adapter

import (
	"context"
	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
)

type permissionAdapter struct {
	queries *sqlc.Queries
}

func NewPermissionAdapter(queries *sqlc.Queries) *permissionAdapter {
	return &permissionAdapter{
		queries: queries,
	}
}

func (p *permissionAdapter) GetPermissions(ctx context.Context, user *domain.User) (domain.Permissions, error) {
	return p.queries.GetPermissions(ctx, int32(user.ID))
}
