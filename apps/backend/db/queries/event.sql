-- name: ListEventsByUserAndDateRange :many
SELECT id, user_id, title, description, start_at, end_at, created_at, updated_at
FROM events
WHERE user_id = sqlc.arg('user_id')
  AND end_at > sqlc.arg('start_date')
  AND start_at < sqlc.arg('end_date')
ORDER BY start_at;

-- name: InsertEvent :one
INSERT INTO events (id, user_id, title, description, start_at, end_at, location, url)
VALUES (
    sqlc.arg('id'),
    sqlc.arg('user_id'),
    sqlc.arg('title'),
    sqlc.arg('description'),
    sqlc.arg('start_at'),
    sqlc.arg('end_at'),
    sqlc.arg('location'),
    sqlc.arg('url')
)
RETURNING id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at;
