package postgres_adapter

import (
	"database/sql"
	"personal_website/internal/app/core/ports"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
)

type transaction struct {
	tx        *sql.Tx
	queries   *sqlc.Queries
	userRepo  ports.UserRepository
	tokenRepo ports.TokenRepository
}

func (t *transaction) UserRepo() ports.UserRepository   { return t.userRepo }
func (t *transaction) TokenRepo() ports.TokenRepository { return t.tokenRepo }
func (t *transaction) Commit() error                    { return t.tx.Commit() }
func (t *transaction) Rollback() error                  { return t.tx.Rollback() }
