-- name: ListArticles :many
SELECT
  id,
  title,
  slug,
  published_at
FROM content.articles
WHERE is_published = true
    AND is_deleted = false
ORDER BY published_at DESC;

-- name: GetArticleBySlug :one
SELECT
  id,
  title,
  slug,
  content,
  created_at,
  published_at,
  is_published
FROM content.articles
WHERE slug = $1
    AND is_published = true
    AND is_deleted = false;
