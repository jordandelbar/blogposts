package datastore_adapter

import (
	"context"
	"personal_website/internal/app/core/ports"
)

type Datastore struct {
	postgresDB ports.PostgresDatabase
	valkeyDB   ports.ValkeyDatabase
}

func NewDatastore(postgresDB ports.PostgresDatabase, valkeyDB ports.ValkeyDatabase) *Datastore {
	return &Datastore{
		postgresDB: postgresDB,
		valkeyDB:   valkeyDB,
	}
}

func (d *Datastore) UserRepo() ports.UserRepository {
	return d.postgresDB.UserRepo()
}

func (d *Datastore) ArticleRepo() ports.ArticleRepository {
	return d.postgresDB.ArticleRepo()
}

func (d *Datastore) PermissionRepo() ports.PermissionRepository {
	return d.postgresDB.PermissionRepo()
}

func (d *Datastore) SessionRepo() ports.SessionRepository {
	return d.valkeyDB.SessionRepo()
}

func (d *Datastore) Begin(ctx context.Context) (ports.Transaction, error) {
	return d.postgresDB.Begin(ctx)
}
