-- name: ListEventsWithTagsByUserAndDateRange :many
-- Unions the user's own events (with tag joins) with subscribed events
-- (SPEC-004). Tags are LEFT JOINed so untagged events produce one NULL-tag
-- row; the Go side aggregates rows into EventReadModel per event (ADR-022).
-- UNION ALL is safe because a user can never subscribe to their own event.
SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.created_at, e.updated_at,
       FALSE AS is_subscribed, NULL::uuid AS subscription_id,
       et.tag_id, et.position, t.name AS tag_name, t.color AS tag_color
FROM events e
LEFT JOIN event_tags et ON et.event_id = e.id
LEFT JOIN tags t ON t.id = et.tag_id
WHERE e.user_id = sqlc.arg('user_id')
  AND e.end_at > sqlc.arg('start_date')
  AND e.start_at < sqlc.arg('end_date')

UNION ALL

SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.created_at, e.updated_at,
       TRUE AS is_subscribed, es.id AS subscription_id,
       NULL::uuid, NULL::smallint, NULL::text, NULL::text
FROM event_subscriptions es
JOIN events e ON e.id = es.event_id
WHERE es.user_id = sqlc.arg('user_id')
  AND e.end_at > sqlc.arg('start_date')
  AND e.start_at < sqlc.arg('end_date')

ORDER BY start_at, subscription_id NULLS FIRST, position NULLS FIRST;

-- name: ListEventsWithTagsByUserAndDateRangeFilteredByTag :many
-- Returns only events owned by the user that carry a specific tag, together
-- with ALL their tags (not just the filter tag). Subscribed events are
-- excluded (SPEC-008: tagId filter applies to own events only).
SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.created_at, e.updated_at,
       FALSE AS is_subscribed, NULL::uuid AS subscription_id,
       et.tag_id, et.position, t.name AS tag_name, t.color AS tag_color
FROM events e
JOIN event_tags et_filter ON et_filter.event_id = e.id
                         AND et_filter.tag_id = sqlc.arg('tag_id')
LEFT JOIN event_tags et ON et.event_id = e.id
LEFT JOIN tags t ON t.id = et.tag_id
WHERE e.user_id = sqlc.arg('user_id')
  AND e.end_at > sqlc.arg('start_date')
  AND e.start_at < sqlc.arg('end_date')
ORDER BY e.start_at, et.position;

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

-- name: FindEventWithTagsByID :many
-- One row per tag (or one row with NULL tag columns when the event has
-- none). Reading the event and its tags in one statement gives them the same
-- snapshot, so a concurrent edit cannot be seen half applied.
SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.created_at, e.updated_at, et.tag_id, et.position
FROM events e
LEFT JOIN event_tags et ON et.event_id = e.id
WHERE e.id = $1
ORDER BY et.position;

-- name: FindEventReadModelByID :one
-- Read-side counterpart of FindEventWithTagsByID (ADR-022 logical CQRS): same table,
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

-- name: ListEventTagsByEventID :many
SELECT tag_id, position
FROM event_tags
WHERE event_id = $1
ORDER BY position;

-- name: InsertEventTag :exec
INSERT INTO event_tags (event_id, tag_id, position)
VALUES ($1, $2, $3);

-- name: DeleteEventTagsByEventID :exec
DELETE FROM event_tags
WHERE event_id = $1;

-- name: FindEventByIDForUpdate :one
SELECT id, user_id, title, description, start_at, end_at, location, url, created_at, updated_at
FROM events
WHERE id = $1
FOR UPDATE;
