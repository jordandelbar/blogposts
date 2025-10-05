package postgres_adapter

import (
	"context"
	"database/sql"
	"errors"

	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"

	"github.com/lib/pq"
)

type articleAdapter struct {
	queries *sqlc.Queries
}

func NewArticleAdapter(queries *sqlc.Queries) *articleAdapter {
	return &articleAdapter{
		queries: queries,
	}
}

func (a *articleAdapter) CreateArticle(ctx context.Context, article domain.Article) error {
	err := a.queries.CreateArticle(ctx, sqlc.CreateArticleParams{
		Title:   article.Title,
		Slug:    article.Slug,
		Content: article.Content,
	})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return domain.ErrArticleAlreadyExists
			}
		}
		return domain.NewInternalError(err)
	}
	return nil
}

func (a *articleAdapter) GetArticleByID(ctx context.Context, id int32) (domain.Article, error) {
	row, err := a.queries.GetArticleById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrArticleNotFound
		}
		return domain.Article{}, domain.NewInternalError(err)
	}
	return a.sqlcRowToArticle(row.ID, row.Title, row.Slug, row.Content, row.CreatedAt, row.PublishedAt, row.IsPublished, sql.NullTime{}, sql.NullBool{}), nil
}

func (a *articleAdapter) GetArticleBySlug(ctx context.Context, slug string) (domain.Article, error) {
	row, err := a.queries.GetArticleBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrArticleNotFound
		}
		return domain.Article{}, domain.NewInternalError(err)
	}
	return a.sqlcRowToArticle(row.ID, row.Title, row.Slug, row.Content, row.CreatedAt, row.PublishedAt, row.IsPublished, sql.NullTime{}, sql.NullBool{}), nil
}

func (a *articleAdapter) UpdateArticle(ctx context.Context, article domain.Article) error {
	err := a.queries.UpdateArticle(ctx, sqlc.UpdateArticleParams{
		ID:      article.ID,
		Title:   article.Title,
		Slug:    article.Slug,
		Content: article.Content,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrArticleNotFound
		}
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return domain.ErrArticleAlreadyExists
			}
		}
		return domain.NewInternalError(err)
	}
	return nil
}

func (a *articleAdapter) PublishArticle(ctx context.Context, id int32) error {
	err := a.queries.PublishArticle(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrArticleNotFound
		}
		return domain.NewInternalError(err)
	}
	return nil
}

func (a *articleAdapter) UnpublishArticle(ctx context.Context, id int32) error {
	err := a.queries.UnpublishArticle(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrArticleNotFound
		}
		return domain.NewInternalError(err)
	}
	return nil
}

func (a *articleAdapter) ListArticles(ctx context.Context) ([]domain.Article, error) {
	rows, err := a.queries.ListArticles(ctx)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	articles := make([]domain.Article, 0, len(rows))
	for _, row := range rows {
		article := a.sqlcRowToArticle(row.ID, row.Title, row.Slug, "", sql.NullTime{}, row.PublishedAt, sql.NullBool{Valid: true, Bool: true}, sql.NullTime{}, sql.NullBool{})
		articles = append(articles, article)
	}
	return articles, nil
}

func (a *articleAdapter) ListAllArticles(ctx context.Context) ([]domain.Article, error) {
	rows, err := a.queries.ListAllArticles(ctx)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	articles := make([]domain.Article, 0, len(rows))
	for _, row := range rows {
		article := a.sqlcRowToArticle(row.ID, row.Title, row.Slug, "", row.CreatedAt, row.PublishedAt, row.IsPublished, row.UpdatedAt, sql.NullBool{})
		articles = append(articles, article)
	}
	return articles, nil
}

func (a *articleAdapter) ListDeletedArticles(ctx context.Context) ([]domain.Article, error) {
	rows, err := a.queries.ListDeletedArticles(ctx)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	articles := make([]domain.Article, 0, len(rows))
	for _, row := range rows {
		article := a.sqlcRowToArticle(row.ID, row.Title, row.Slug, "", row.CreatedAt, row.PublishedAt, row.IsPublished, row.UpdatedAt, sql.NullBool{Valid: true, Bool: true})
		if row.DeletedAt.Valid {
			article.DeletedAt = row.DeletedAt.Time
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (a *articleAdapter) SoftDeleteArticle(ctx context.Context, id int32) error {
	rowsAffected, err := a.queries.SoftDeleteArticle(ctx, id)
	if err != nil {
		return domain.NewInternalError(err)
	}
	if rowsAffected == 0 {
		return domain.ErrArticleNotFound
	}
	return nil
}

func (a *articleAdapter) DeleteArticle(ctx context.Context, id int32) error {
	rowsAffected, err := a.queries.DeleteArticle(ctx, id)
	if err != nil {
		return domain.NewInternalError(err)
	}

	if rowsAffected == 0 {
		// Check if article exists but is not soft-deleted
		article, err := a.queries.GetAllArticlesByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrArticleNotFound
			}
			return domain.NewInternalError(err)
		}

		// Article exists but is not soft-deleted (or deleted_at is NULL)
		if !article.IsDeleted.Bool || !article.DeletedAt.Valid {
			return domain.ErrArticleNotSoftDeleted
		}

		// This should not happen, but fallback to not found
		return domain.ErrArticleNotFound
	}

	return nil
}

func (a *articleAdapter) RestoreArticle(ctx context.Context, id int32) error {
	err := a.queries.RestoreArticle(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrArticleNotFound
		}
		return domain.NewInternalError(err)
	}
	return nil
}

func (a *articleAdapter) sqlcRowToArticle(id int32, title, slug, content string, createdAt, publishedAt sql.NullTime, isPublished sql.NullBool, updatedAt sql.NullTime, isDeleted sql.NullBool) domain.Article {
	article := domain.Article{
		ID:      id,
		Title:   title,
		Slug:    slug,
		Content: content,
	}

	if createdAt.Valid {
		article.CreatedAt = createdAt.Time
	}

	if publishedAt.Valid {
		article.PublishedAt = publishedAt.Time
	}

	if isPublished.Valid {
		article.IsPublished = isPublished.Bool
	}

	return article
}
