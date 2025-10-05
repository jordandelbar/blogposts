package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"personal_website/internal/infrastructure/adapters/repository/postgres/sqlc"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateArticle(t *testing.T) {
	suite := NewTestSuite(t)

	resp, err := suite.POST(t, "/v1/articles", ArticleData())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateArticle_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	suite := NewTestSuite(t)

	testCases := []struct {
		name        string
		jsonBody    string
		expectedMsg string
	}{
		{
			name:        "malformed JSON",
			jsonBody:    `{"title": "Test", "slug": "test", "content": "test"`, // missing closing brace
			expectedMsg: "body contains badly-formed JSON",
		},
		{
			name:        "empty body",
			jsonBody:    "",
			expectedMsg: "body must not be empty",
		},
		{
			name:        "unknown field",
			jsonBody:    `{"title": "Test", "slug": "test", "content": "test content", "unknown": "field"}`,
			expectedMsg: "body contains unknown key",
		},
		{
			name:        "wrong type",
			jsonBody:    `{"title": 123, "slug": "test", "content": "test content"}`, // title should be string
			expectedMsg: "body contains incorrect JSON type for field \"title\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := NewRequestWithAuthentication(t, "POST", suite.ServerAddr+"/v1/articles", suite.AuthToken, []byte(tc.jsonBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			suite.AssertJSONError(t, resp, tc.expectedMsg)
		})
	}
}

func TestCreateArticleValidation(t *testing.T) {
	suite := NewTestSuite(t)

	testCases := []struct {
		name         string
		data         map[string]string
		expectedCode int
		expectedMsg  string
	}{
		{
			name: "missing title",
			data: map[string]string{
				"slug":    "test-slug",
				"content": "This is valid content that is definitely longer than 50 characters as required by validation",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "missing slug",
			data: map[string]string{
				"title":   "Valid Title",
				"content": "This is valid content that is definitely longer than 50 characters as required by validation",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "missing content",
			data: map[string]string{
				"title": "Valid Title",
				"slug":  "valid-slug",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "title too short",
			data: map[string]string{
				"title":   "Hi", // less than 5 chars
				"slug":    "valid-slug",
				"content": "This is valid content that is definitely longer than 50 characters as required by validation",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "title too long",
			data: map[string]string{
				"title":   "This title is way too long and exceeds the 200 character limit set by the validation rules. It just keeps going and going and going until it definitely exceeds the maximum length allowed for article titles in this system which is exactly 200 characters according to the validation rules we have defined",
				"slug":    "valid-slug",
				"content": "This is valid content that is definitely longer than 50 characters as required by validation",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "slug too short",
			data: map[string]string{
				"title":   "Valid Title",
				"slug":    "hi", // less than 3 chars
				"content": "This is valid content that is definitely longer than 50 characters as required by validation",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "slug invalid format",
			data: map[string]string{
				"title":   "Valid Title",
				"slug":    "invalid slug with spaces!", // not alphanum_hyphen
				"content": "This is valid content that is definitely longer than 50 characters as required by validation",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
		{
			name: "content too short",
			data: map[string]string{
				"title":   "Valid Title",
				"slug":    "valid-slug",
				"content": "Short", // less than 50 chars
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "validation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.POST(t, "/v1/articles", tc.data)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			// For some specific cases, check field validation errors
			switch tc.name {
			case "missing title":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"title": "title is required",
				})
			case "missing slug":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"slug": "slug is required",
				})
			case "missing content":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"content": "content is required",
				})
			case "title too short":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"title": "title must be at least 5 characters",
				})
			case "title too long":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"title": "title must not exceed 200 characters",
				})
			case "slug too short":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"slug": "slug must be at least 3 characters",
				})
			case "slug invalid format":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"slug": "slug must contain only letters, numbers and hyphens",
				})
			case "content too short":
				suite.AssertValidationError(t, resp, tc.expectedMsg, map[string]string{
					"content": "content must be at least 50 characters",
				})
			default:
				suite.AssertValidationError(t, resp, tc.expectedMsg)
			}
		})
	}
}

func TestUpdateArticle(t *testing.T) {
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

	// Update the article
	updateData := map[string]string{
		"title":   "Updated Title",
		"slug":    "updated-slug",
		"content": "Updated content. It needs to be at least 50 characters!",
	}

	jsonData, err := json.Marshal(updateData)
	require.NoError(t, err)

	url := fmt.Sprintf("%s/v1/articles/id/%d", suite.ServerAddr, articleID)
	resp, err := NewRequestWithAuthentication(t, "PUT", url, suite.AuthToken, jsonData)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPublishArticle(t *testing.T) {
	suite := NewTestSuite(t)

	ctx := context.Background()

	// Create a test article first
	createParams := sqlc.CreateArticleParams{
		Title:   "Article to Publish",
		Slug:    "article-to-publish",
		Content: "Content to publish.",
	}
	err := queries.CreateArticle(ctx, createParams)
	require.NoError(t, err)

	// Get the article ID by querying directly
	var articleID int32
	err = db.QueryRow("SELECT id FROM content.articles WHERE slug = $1", "article-to-publish").Scan(&articleID)
	require.NoError(t, err)

	// Publish the article with authentication
	url := fmt.Sprintf("%s/v1/articles/id/%d/publish", suite.ServerAddr, articleID)
	resp, err := NewRequestWithAuthentication(t, "PATCH", url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the article is now published and can be retrieved
	getResp, err := http.Get(suite.ServerAddr + "/v1/articles/slug/article-to-publish")
	require.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
}

func TestUnpublishArticle(t *testing.T) {
	suite := NewTestSuite(t)

	ctx := context.Background()

	// Create a test article first
	createParams := sqlc.CreateArticleParams{
		Title:   "Article to Publish",
		Slug:    "article-to-publish",
		Content: "Content to publish.",
	}
	err := queries.CreateArticle(ctx, createParams)
	require.NoError(t, err)

	// Get the article ID by querying directly
	var articleID int32
	err = db.QueryRow("SELECT id FROM content.articles WHERE slug = $1", "article-to-publish").Scan(&articleID)
	require.NoError(t, err)

	// Unpublish the article with authentication
	url := fmt.Sprintf("%s/v1/articles/id/%d/unpublish", suite.ServerAddr, articleID)
	resp, err := NewRequestWithAuthentication(t, "PATCH", url, suite.AuthToken, nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateArticle_DuplicateSlug_ReturnsConflict(t *testing.T) {
	suite := NewTestSuite(t)

	articleData := map[string]string{
		"title":   "First Article",
		"slug":    "duplicate-slug",
		"content": "This is the first article with this slug. It must be at least 50 characters long.",
	}

	// Create article first time - should succeed
	resp, err := suite.POST(t, "/v1/articles", articleData)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Try to create article with same slug - should return conflict
	articleData2 := map[string]string{
		"title":   "Second Article",
		"slug":    "duplicate-slug", // Same slug as first article
		"content": "This is the second article trying to use the same slug. Must be 50+ chars.",
	}

	resp2, err := suite.POST(t, "/v1/articles", articleData2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
	suite.AssertJSONError(t, resp2, "article already exists")
}
