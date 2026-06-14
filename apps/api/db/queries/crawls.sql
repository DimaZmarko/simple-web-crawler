-- name: CreateCrawl :one
INSERT INTO crawls (seed_url, max_depth, max_pages, status)
VALUES ($1, $2, $3, 'queued')
RETURNING id, seed_url, max_depth, max_pages, status, created_at, updated_at;

-- name: GetCrawlByID :one
SELECT id, seed_url, max_depth, max_pages, status, created_at, updated_at
FROM crawls
WHERE id = $1;

-- name: ListCrawlsFirstPage :many
SELECT id, seed_url, max_depth, max_pages, status, created_at, updated_at
FROM crawls
ORDER BY created_at DESC, id DESC
LIMIT $1;

-- name: ListCrawlsAfterCursor :many
SELECT id, seed_url, max_depth, max_pages, status, created_at, updated_at
FROM crawls
WHERE (created_at, id) < (sqlc.arg(cursor_created_at)::timestamptz, sqlc.arg(cursor_id)::uuid)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit);
