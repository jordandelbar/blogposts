package docs

import (
	"personal_website/internal/app/core/domain"
	"time"
)

// ArticleResponse represents an article in API responses
// swagger:model ArticleResponse
type ArticleResponse struct {
	// The unique identifier of the article
	// required: true
	// example: 1
	ID int32 `json:"id"`

	// The title of the article
	// required: true
	// example: "My First Blog Post"
	Title string `json:"title"`

	// The URL-friendly slug for the article
	// required: true
	// example: "my-first-blog-post"
	Slug string `json:"slug"`

	// The article content in markdown format
	// required: true
	// example: "# Welcome\n\nThis is my first blog post..."
	Content string `json:"content"`

	// When the article was created
	// required: true
	// example: "2023-01-15T10:30:00Z"
	CreatedAt time.Time `json:"created_at"`

	// When the article was published
	// example: "2023-01-15T12:00:00Z"
	PublishedAt time.Time `json:"published_at"`

	// When the article was deleted (soft delete)
	// example: "2023-01-20T15:30:00Z"
	DeletedAt time.Time `json:"deleted_at"`

	// Whether the article is published
	// required: true
	// example: true
	IsPublished bool `json:"is_published"`
}

// CreateArticleRequest represents the request body for creating an article
// swagger:model CreateArticleRequest
type CreateArticleRequest struct {
	// The title of the article
	// required: true
	// example: "My New Article"
	Title string `json:"title" validate:"required,min=1,max=200"`

	// The URL-friendly slug for the article (auto-generated if not provided)
	// example: "my-new-article"
	Slug string `json:"slug" validate:"omitempty,min=1,max=200"`

	// The article content in markdown format
	// required: true
	// example: "# Introduction\n\nThis is the content of my article..."
	Content string `json:"content" validate:"required,min=1"`
}

// UpdateArticleRequest represents the request body for updating an article
// swagger:model UpdateArticleRequest
type UpdateArticleRequest struct {
	// The title of the article
	// example: "Updated Article Title"
	Title string `json:"title" validate:"omitempty,min=1,max=200"`

	// The URL-friendly slug for the article
	// example: "updated-article-title"
	Slug string `json:"slug" validate:"omitempty,min=1,max=200"`

	// The article content in markdown format
	// example: "# Updated Content\n\nThis is the updated content..."
	Content string `json:"content" validate:"omitempty,min=1"`
}

// ContactRequest represents a contact form submission
// swagger:model ContactRequest
type ContactRequest struct {
	// The name of the person contacting
	// required: true
	// example: "John Doe"
	Name string `json:"name" validate:"required,min=1,max=100"`

	// The email address of the person contacting
	// required: true
	// example: "john.doe@example.com"
	Email string `json:"email" validate:"required,email"`

	// The subject of the message
	// required: true
	// example: "Inquiry about your services"
	Subject string `json:"subject" validate:"required,min=1,max=200"`

	// The message content
	// required: true
	// example: "Hello, I'm interested in learning more about your services..."
	Message string `json:"message" validate:"required,min=1,max=2000"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	// The error message
	Error string `json:"error" example:"Invalid request parameters"`
	// Request correlation ID for tracking
	CorrelationID string `json:"correlation_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// SuccessResponse represents a generic success response
// swagger:model SuccessResponse
type SuccessResponse struct {
	// Success message
	// required: true
	// example: "Operation completed successfully"
	Message string `json:"message"`

	// Request correlation ID for tracking
	// example: "123e4567-e89b-12d3-a456-426614174000"
	CorrelationID string `json:"correlation_id,omitempty"`
}

// ArticleListResponse represents a list of articles
// swagger:model ArticleListResponse
type ArticleListResponse struct {
	// List of articles
	// required: true
	Articles []ArticleResponse `json:"articles"`

	// Total count of articles
	// required: true
	// example: 25
	Total int `json:"total"`
}

// Helper function to convert domain article to response model
func ToArticleResponse(article domain.Article) ArticleResponse {
	return ArticleResponse{
		ID:          article.ID,
		Title:       article.Title,
		Slug:        article.Slug,
		Content:     article.Content,
		CreatedAt:   article.CreatedAt,
		PublishedAt: article.PublishedAt,
		DeletedAt:   article.DeletedAt,
		IsPublished: article.IsPublished,
	}
}

// Helper function to convert multiple domain articles to response models
func ToArticleListResponse(articles []domain.Article) ArticleListResponse {
	responses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		responses[i] = ToArticleResponse(article)
	}
	return ArticleListResponse{
		Articles: responses,
		Total:    len(responses),
	}
}
