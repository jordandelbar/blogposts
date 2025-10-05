package ports

import (
	"context"
	"personal_website/internal/app/core/domain"
)

type ArticleRepository interface {
	CreateArticle(ctx context.Context, article domain.Article) error
	GetArticleByID(ctx context.Context, id int32) (domain.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (domain.Article, error)
	UpdateArticle(ctx context.Context, article domain.Article) error
	PublishArticle(ctx context.Context, id int32) error
	UnpublishArticle(ctx context.Context, id int32) error
	ListArticles(ctx context.Context) ([]domain.Article, error)
	ListAllArticles(ctx context.Context) ([]domain.Article, error)
	ListDeletedArticles(ctx context.Context) ([]domain.Article, error)
	SoftDeleteArticle(ctx context.Context, id int32) error
	DeleteArticle(ctx context.Context, id int32) error
	RestoreArticle(ctx context.Context, id int32) error
}
