package postgres_adapter

import (
	"context"
	"database/sql"
	"errors"

	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"

	"github.com/lib/pq"
)

type userAdapter struct {
	queries *sqlc.Queries
}

func NewUserAdapter(queries *sqlc.Queries) *userAdapter {
	return &userAdapter{
		queries: queries,
	}
}

func (u *userAdapter) CreateUser(ctx context.Context, user domain.User) (int, error) {
	userId, err := u.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.Password.Hash(),
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return 0, domain.ErrUserAlreadyExists
			}
		}
		return 0, domain.NewInternalError(err)
	}
	return int(userId), err
}

func (u *userAdapter) ActivateUser(ctx context.Context, user *domain.User) error {
	err := u.queries.ActivateUser(ctx, int32(user.ID))
	if err != nil {
		return domain.NewInternalError(err)
	}
	return nil
}

func (u *userAdapter) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	row, err := u.queries.CheckUserExistsByEmail(ctx, email)
	if err != nil {
		return false, domain.NewInternalError(err)
	}

	return row, nil
}

func (u *userAdapter) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := u.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// In this case (authentication) we return invalid credential error message
			return domain.User{}, domain.ErrInvalidCredentials
		}
		return domain.User{}, domain.NewInternalError(err)
	}

	user := domain.User{
		ID:        int(row.ID),
		CreatedAt: row.CreatedAt.Time,
		Name:      row.Name,
		Email:     row.Email,
		Activated: row.Activated,
	}

	user.Password.SetHash(row.PasswordHash)

	return user, nil
}

func (u *userAdapter) DeactivateUser(ctx context.Context, id int) error {
	err := u.queries.DeactivateUser(ctx, int32(id))
	if err != nil {
		return domain.NewInternalError(err)
	}

	return nil
}

func (u *userAdapter) DeleteUser(ctx context.Context, id int) error {
	err := u.queries.DeleteUser(ctx, int32(id))
	if err != nil {
		return domain.NewInternalError(err)
	}

	return nil
}
