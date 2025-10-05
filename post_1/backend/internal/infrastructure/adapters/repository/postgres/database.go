package postgres_adapter

import (
	"context"
	"database/sql"
	"personal_website/config"
	"personal_website/internal/app/core/ports"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
)

type database struct {
	db             *sql.DB
	queries        *sqlc.Queries
	articleRepo    ports.ArticleRepository
	userRepo       ports.UserRepository
	permissionRepo ports.PermissionRepository
}

func NewDatabase(cfg *config.PostgresConfig) (*database, error) {
	db, err := NewConnection(cfg)
	if err != nil {
		return &database{}, err
	}
	queries := sqlc.New(db)

	return &database{
		db:             db,
		queries:        queries,
		articleRepo:    NewArticleAdapter(queries),
		userRepo:       NewUserAdapter(queries),
		permissionRepo: NewPermissionAdapter(queries),
	}, nil
}

func (d *database) UserRepo() ports.UserRepository             { return d.userRepo }
func (d *database) ArticleRepo() ports.ArticleRepository       { return d.articleRepo }
func (d *database) PermissionRepo() ports.PermissionRepository { return d.permissionRepo }

func (d *database) Begin(ctx context.Context) (ports.Transaction, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	qtx := sqlc.New(tx)
	return &transaction{
		tx:       tx,
		queries:  qtx,
		userRepo: NewUserAdapter(qtx),
	}, nil
}

func (d *database) Close() {
	d.db.Close()
}
