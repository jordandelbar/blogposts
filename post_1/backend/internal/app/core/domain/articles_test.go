package domain

import (
	"testing"
	"time"
)

func TestArticle_Fields(t *testing.T) {
	// Test that Article struct fields can be properly set and retrieved
	now := time.Now()
	publishedTime := now.Add(-time.Hour)
	deletedTime := now.Add(time.Hour)

	article := Article{
		ID:          123,
		Title:       "Test Article Title",
		Slug:        "test-article-title",
		Content:     "This is the content of the test article.",
		CreatedAt:   now,
		PublishedAt: publishedTime,
		DeletedAt:   deletedTime,
		IsPublished: true,
	}

	// Verify all fields are set correctly
	if article.ID != 123 {
		t.Errorf("Article.ID = %v, want %v", article.ID, 123)
	}

	if article.Title != "Test Article Title" {
		t.Errorf("Article.Title = %v, want %v", article.Title, "Test Article Title")
	}

	if article.Slug != "test-article-title" {
		t.Errorf("Article.Slug = %v, want %v", article.Slug, "test-article-title")
	}

	if article.Content != "This is the content of the test article." {
		t.Errorf("Article.Content = %v, want %v", article.Content, "This is the content of the test article.")
	}

	if !article.CreatedAt.Equal(now) {
		t.Errorf("Article.CreatedAt = %v, want %v", article.CreatedAt, now)
	}

	if !article.PublishedAt.Equal(publishedTime) {
		t.Errorf("Article.PublishedAt = %v, want %v", article.PublishedAt, publishedTime)
	}

	if !article.DeletedAt.Equal(deletedTime) {
		t.Errorf("Article.DeletedAt = %v, want %v", article.DeletedAt, deletedTime)
	}

	if !article.IsPublished {
		t.Errorf("Article.IsPublished = %v, want %v", article.IsPublished, true)
	}
}

func TestArticle_ZeroValues(t *testing.T) {
	// Test zero values behavior
	var article Article

	if article.ID != 0 {
		t.Errorf("Zero-value Article.ID = %v, want 0", article.ID)
	}

	if article.Title != "" {
		t.Errorf("Zero-value Article.Title = %v, want empty string", article.Title)
	}

	if article.Slug != "" {
		t.Errorf("Zero-value Article.Slug = %v, want empty string", article.Slug)
	}

	if article.Content != "" {
		t.Errorf("Zero-value Article.Content = %v, want empty string", article.Content)
	}

	if article.IsPublished {
		t.Error("Zero-value Article.IsPublished should be false")
	}

	// Test zero-value times
	zeroTime := time.Time{}
	if !article.CreatedAt.Equal(zeroTime) {
		t.Error("Zero-value Article.CreatedAt should be zero time")
	}

	if !article.PublishedAt.Equal(zeroTime) {
		t.Error("Zero-value Article.PublishedAt should be zero time")
	}

	if !article.DeletedAt.Equal(zeroTime) {
		t.Error("Zero-value Article.DeletedAt should be zero time")
	}
}

func TestArticle_IDType(t *testing.T) {
	// Test that ID is int32 type as specified
	article := Article{
		ID: 2147483647, // Max int32 value
	}

	if article.ID != 2147483647 {
		t.Errorf("Article.ID = %v, want %v", article.ID, 2147483647)
	}

	// Test negative ID (edge case)
	article.ID = -1
	if article.ID != -1 {
		t.Errorf("Article.ID = %v, want %v", article.ID, -1)
	}
}

func TestArticle_PublishingStates(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		isPublished bool
		publishedAt time.Time
		description string
	}{
		{
			name:        "published article",
			isPublished: true,
			publishedAt: now.Add(-time.Hour),
			description: "Article published 1 hour ago",
		},
		{
			name:        "unpublished draft",
			isPublished: false,
			publishedAt: time.Time{}, // zero time
			description: "Draft article not yet published",
		},
		{
			name:        "published article with future date",
			isPublished: true,
			publishedAt: now.Add(time.Hour),
			description: "Article marked published but with future publish date",
		},
		{
			name:        "unpublished with past date",
			isPublished: false,
			publishedAt: now.Add(-time.Hour),
			description: "Article with past publish date but marked unpublished",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := Article{
				ID:          1,
				Title:       "Test Article",
				IsPublished: tt.isPublished,
				PublishedAt: tt.publishedAt,
				CreatedAt:   now.Add(-2 * time.Hour),
			}

			// Verify the states are set correctly
			if article.IsPublished != tt.isPublished {
				t.Errorf("Article.IsPublished = %v, want %v", article.IsPublished, tt.isPublished)
			}

			if !article.PublishedAt.Equal(tt.publishedAt) {
				t.Errorf("Article.PublishedAt = %v, want %v", article.PublishedAt, tt.publishedAt)
			}
		})
	}
}

func TestArticle_DeletionState(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		deletedAt time.Time
		isDeleted bool
	}{
		{
			name:      "not deleted",
			deletedAt: time.Time{}, // zero time
			isDeleted: false,
		},
		{
			name:      "deleted in past",
			deletedAt: now.Add(-time.Hour),
			isDeleted: true,
		},
		{
			name:      "deleted in future",
			deletedAt: now.Add(time.Hour),
			isDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := Article{
				ID:        1,
				Title:     "Test Article",
				DeletedAt: tt.deletedAt,
				CreatedAt: now.Add(-2 * time.Hour),
			}

			// Check if deletion timestamp is set
			isDeleted := !article.DeletedAt.IsZero()
			if isDeleted != tt.isDeleted {
				t.Errorf("Article deletion state = %v, want %v", isDeleted, tt.isDeleted)
			}

			if !article.DeletedAt.Equal(tt.deletedAt) {
				t.Errorf("Article.DeletedAt = %v, want %v", article.DeletedAt, tt.deletedAt)
			}
		})
	}
}

func TestArticle_SlugHandling(t *testing.T) {
	tests := []struct {
		name  string
		title string
		slug  string
	}{
		{
			name:  "normal slug",
			title: "My Great Article",
			slug:  "my-great-article",
		},
		{
			name:  "empty slug",
			title: "Article Title",
			slug:  "",
		},
		{
			name:  "slug with numbers",
			title: "Article 123",
			slug:  "article-123",
		},
		{
			name:  "complex slug",
			title: "How to Build APIs in Go: A Complete Guide",
			slug:  "how-to-build-apis-in-go-a-complete-guide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := Article{
				ID:    1,
				Title: tt.title,
				Slug:  tt.slug,
			}

			if article.Title != tt.title {
				t.Errorf("Article.Title = %v, want %v", article.Title, tt.title)
			}

			if article.Slug != tt.slug {
				t.Errorf("Article.Slug = %v, want %v", article.Slug, tt.slug)
			}
		})
	}
}