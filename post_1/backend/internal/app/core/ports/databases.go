package ports

import (
	"context"
)

type PostgresDatabase interface {
	UserRepo() UserRepository
	PermissionRepo() PermissionRepository
	ArticleRepo() ArticleRepository
	Begin(ctx context.Context) (Transaction, error)
	Close()
}

type ValkeyDatabase interface {
	SessionRepo() SessionRepository
	Close()
}
