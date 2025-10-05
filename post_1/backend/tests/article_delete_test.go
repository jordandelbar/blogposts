package tests

import (
	"context"
	"fmt"
	"net/http"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSoftDeleteArticle(t *testing.T) {
	suite := NewTestSuite(t)

	ctx := context.Background()

	// Create a test article first
	createParams := sqlc.CreateArticleParams{
		Title:   "Original Title",
		Slug:    "original-slug",
		Content: "Original content.",
	}
	err := queries.CreateArticle(ctx, createParams)
	require.NoError(t, err)

	// Get the article ID by querying directly
	var articleID int32
	err = db.QueryRow("SELECT id FROM content.articles WHERE slug = $1", "original-slug").Scan(&articleID)
	require.NoError(t, err)

	url := fmt.Sprintf("/v1/articles/id/%d", articleID)
	resp, err := NewRequestWithAuthentication(t, "DELETE", suite.ServerAddr+url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSoftDeleteNonExistentArticle(t *testing.T) {
	suite := NewTestSuite(t)

	// Try to delete an article that doesn't exist
	nonExistentID := int32(99999)
	url := fmt.Sprintf("/v1/articles/id/%d", nonExistentID)
	resp, err := NewRequestWithAuthentication(t, "DELETE", suite.ServerAddr+url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	suite.AssertJSONError(t, resp, "article not found")
}

func TestHardDeleteArticle(t *testing.T) {
	suite := NewTestSuite(t)

	ctx := context.Background()

	// Create a test article first
	createParams := sqlc.CreateArticleParams{
		Title:   "Original Title",
		Slug:    "original-slug",
		Content: "Original content.",
	}
	err := queries.CreateArticle(ctx, createParams)
	require.NoError(t, err)

	// Get the article ID by querying directly
	var articleID int32
	err = db.QueryRow("SELECT id FROM content.articles WHERE slug = $1", "original-slug").Scan(&articleID)
	require.NoError(t, err)

	// First soft-delete the article (hard delete only works on soft-deleted articles)
	softDeleteURL := fmt.Sprintf("/v1/articles/id/%d", articleID)
	softResp, err := NewRequestWithAuthentication(t, "DELETE", suite.ServerAddr+softDeleteURL, suite.AuthToken, nil)
	require.NoError(t, err)
	defer softResp.Body.Close()
	require.Equal(t, http.StatusOK, softResp.StatusCode)

	url := fmt.Sprintf("/v1/articles/id/%d/permanent", articleID)
	resp, err := NewRequestWithAuthentication(t, "DELETE", suite.ServerAddr+url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHardDeleteNonExistentArticle(t *testing.T) {
	suite := NewTestSuite(t)

	// Try to delete an article that doesn't exist
	nonExistentID := int32(99999)
	url := fmt.Sprintf("/v1/articles/id/%d/permanent", nonExistentID)
	resp, err := NewRequestWithAuthentication(t, "DELETE", suite.ServerAddr+url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	suite.AssertJSONError(t, resp, "article not found")
}

func TestHardDeleteNonSoftDeletedArticle(t *testing.T) {
	suite := NewTestSuite(t)

	ctx := context.Background()

	// Create a test article (but don't soft-delete it)
	createParams := sqlc.CreateArticleParams{
		Title:   "Article Not Soft Deleted",
		Slug:    "article-not-soft-deleted",
		Content: "This article exists but is not soft deleted.",
	}
	err := queries.CreateArticle(ctx, createParams)
	require.NoError(t, err)

	// Get the article ID by querying directly
	var articleID int32
	err = db.QueryRow("SELECT id FROM content.articles WHERE slug = $1", "article-not-soft-deleted").Scan(&articleID)
	require.NoError(t, err)

	// Try to hard delete the article without soft deleting it first
	url := fmt.Sprintf("/v1/articles/id/%d/permanent", articleID)
	resp, err := NewRequestWithAuthentication(t, "DELETE", suite.ServerAddr+url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	suite.AssertJSONError(t, resp, "article must be soft deleted before permanent deletion")
}
