package ports

import "context"

type Datastore interface {
	UserRepo() UserRepository
	SessionRepo() SessionRepository
	PermissionRepo() PermissionRepository
	ArticleRepo() ArticleRepository
	Begin(ctx context.Context) (Transaction, error)
}
