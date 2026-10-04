-- name: ListEventsByUserAndDateRange :many
SELECT id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at
FROM events
WHERE user_id = sqlc.arg('user_id')
  AND end_at > sqlc.arg('start_date')
  AND start_at < sqlc.arg('end_date')
ORDER BY start_at;

-- name: ListEventsByIDsAndDateRange :many
-- Bulk counterpart of FindEventReadModelByID, used to resolve a user's
-- EventSubscription targets in one round trip instead of one query per row.
SELECT id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at
FROM events
WHERE id = ANY(sqlc.arg('ids')::uuid[])
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

-- name: FindEventByID :one
SELECT id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at
FROM events
WHERE id = $1
LIMIT 1;

-- name: FindEventReadModelByID :one
-- Read-side counterpart of FindEventByID (ADR-022 logical CQRS): same table,
-- kept as its own query so the read path can diverge independently later.
SELECT id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at
FROM events
WHERE id = $1
LIMIT 1;

-- name: UpdateEvent :one
UPDATE events
SET title = $2,
    description = $3,
    start_at = $4,
    end_at = $5,
    location = $6,
    url = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = $1;
