package mappers

import (
	"personal_website/internal/app/core/domain"
	"personal_website/internal/infrastructure/http/dto"
)

func ArticleRequestToDomain(req dto.ArticleRequest) domain.Article {
	return domain.Article{
		Title:   req.Title,
		Slug:    req.Slug,
		Content: req.Content,
	}
}

func ArticleRequestToDomainWithID(req dto.ArticleRequest, id int32) domain.Article {
	return domain.Article{
		ID:      id,
		Title:   req.Title,
		Slug:    req.Slug,
		Content: req.Content,
	}
}

func ArticleToPreview(article domain.Article) dto.ArticlePreview {
	preview := dto.ArticlePreview{
		ID:          article.ID,
		Title:       article.Title,
		Slug:        article.Slug,
		IsPublished: article.IsPublished,
	}

	if !article.CreatedAt.IsZero() {
		createdAt := article.CreatedAt.Format("2006-01-02")
		preview.Created_at = &createdAt
	}

	if !article.PublishedAt.IsZero() {
		publishedAt := article.PublishedAt.Format("2006-01-02")
		preview.Published_at = &publishedAt
	}

	if !article.DeletedAt.IsZero() {
		deletedAt := article.DeletedAt.Format("2006-01-02")
		preview.Deleted_at = &deletedAt
	}

	return preview
}

func ArticlesToPreviews(articles []domain.Article) []dto.ArticlePreview {
	previews := make([]dto.ArticlePreview, len(articles))
	for i, article := range articles {
		previews[i] = ArticleToPreview(article)
	}
	return previews
}

func ArticleToResponse(article domain.Article) dto.ArticleResponse {
	response := dto.ArticleResponse{
		ID:          article.ID,
		Title:       article.Title,
		Slug:        article.Slug,
		Content:     article.Content,
		IsPublished: article.IsPublished,
	}

	if !article.CreatedAt.IsZero() {
		createdAt := article.CreatedAt.Format("2006-01-02")
		response.Created_at = &createdAt
	}

	if !article.PublishedAt.IsZero() {
		publishedAt := article.PublishedAt.Format("2006-01-02")
		response.Published_at = &publishedAt
	}

	if !article.DeletedAt.IsZero() {
		deletedAt := article.DeletedAt.Format("2006-01-02")
		response.Deleted_at = &deletedAt
	}

	return response
}
