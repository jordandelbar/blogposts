-- name: ListAllArticles :many
SELECT
  id,
  title,
  slug,
  published_at,
  is_published,
  created_at,
  updated_at
FROM content.articles
WHERE is_deleted = false OR is_deleted IS NULL
ORDER BY created_at DESC;

-- name: GetArticleById :one
SELECT
  id,
  title,
  slug,
  content,
  created_at,
  published_at,
  is_published
FROM content.articles
WHERE id = $1
    AND (is_deleted = false OR is_deleted IS NULL);

-- name: GetAllArticlesByID :one
SELECT
  id,
  title,
  slug,
  content,
  created_at,
  published_at,
  is_published,
  is_deleted,
  deleted_at
FROM content.articles
WHERE id = $1;

-- name: CreateArticle :exec
INSERT INTO content.articles (
    title,
    slug,
    content
) VALUES ($1, $2, $3);

-- name: UpdateArticle :exec
UPDATE content.articles
SET
    title = $2,
    slug = $3,
    content = $4,
    updated_at = now()
WHERE id = $1;

-- name: PublishArticle :exec
UPDATE content.articles
SET
    is_published = true,
    published_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: UnpublishArticle :exec
UPDATE content.articles
SET
    is_published = false,
    published_at = NULL,
    updated_at = now()
WHERE id = $1;

-- name: SoftDeleteArticle :execrows
UPDATE content.articles
SET
    is_deleted = true,
    is_published = false,
    deleted_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: DeleteArticle :execrows
DELETE FROM content.articles
WHERE id = $1
    AND is_deleted = true
    AND deleted_at IS NOT NULL;

-- name: RestoreArticle :exec
UPDATE content.articles
SET
    is_deleted = false,
    deleted_at = NULL,
    updated_at = now()
WHERE id = $1;

-- name: ListDeletedArticles :many
SELECT
  id,
  title,
  slug,
  published_at,
  is_published,
  created_at,
  updated_at,
  deleted_at
FROM content.articles
WHERE is_deleted = true
ORDER BY deleted_at DESC;
