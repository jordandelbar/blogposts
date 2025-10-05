package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetArticleBySlug(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)

	ctx := context.Background()

	// Create and publish a test article
	createParams := sqlc.CreateArticleParams{
		Title:   "Published Test Article",
		Slug:    "published-test-article",
		Content: "This is published content.",
	}
	err := queries.CreateArticle(ctx, createParams)
	require.NoError(t, err)

	// Get the article ID by querying directly (since GetArticleBySlug only returns published articles)
	var articleID int32
	err = db.QueryRow("SELECT id FROM content.articles WHERE slug = $1", "published-test-article").Scan(&articleID)
	require.NoError(t, err)

	// Publish it
	err = queries.PublishArticle(ctx, articleID)
	require.NoError(t, err)

	// Test getting the article
	resp, err := http.Get(ts.ServerAddr + "/v1/articles/slug/published-test-article")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].(map[string]any)
	assert.Equal(t, "Published Test Article", data["title"])
	assert.Equal(t, "published-test-article", data["slug"])
	assert.Equal(t, "This is published content.", data["content"])
}

func TestGetArticleBySlugNotFound(t *testing.T) {
	ts := NewUnauthenticatedTestSuite(t)

	resp, err := http.Get(ts.ServerAddr + "/v1/articles/slug/non-existent-slug")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
