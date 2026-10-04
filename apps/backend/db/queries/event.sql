-- name: ListEventsByUserAndDateRange :many
-- Unions the user's own events with events they added via EventSubscription
-- (SPEC-004), so the calendar listing read model is built in one query
-- instead of the application layer joining across repositories (ADR-022
-- logical CQRS: the query side owns its own read model). UNION ALL is safe
-- because a user can never subscribe to their own event, so the two halves
-- never overlap.
SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.created_at, e.updated_at, FALSE AS is_subscribed, NULL::uuid AS subscription_id
FROM events e
WHERE e.user_id = sqlc.arg('user_id')
  AND e.end_at > sqlc.arg('start_date')
  AND e.start_at < sqlc.arg('end_date')

UNION ALL

SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.created_at, e.updated_at, TRUE AS is_subscribed, es.id AS subscription_id
FROM event_subscriptions es
JOIN events e ON e.id = es.event_id
WHERE es.user_id = sqlc.arg('user_id')
  AND e.end_at > sqlc.arg('start_date')
  AND e.start_at < sqlc.arg('end_date')

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
