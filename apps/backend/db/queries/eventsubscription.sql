-- name: InsertEventSubscription :one
INSERT INTO event_subscriptions (id, event_id, user_id)
VALUES (
    sqlc.arg('id'),
    sqlc.arg('event_id'),
    sqlc.arg('user_id')
)
RETURNING id, event_id, user_id, created_at;

-- name: FindEventSubscriptionByID :one
SELECT id, event_id, user_id, created_at
FROM event_subscriptions
WHERE id = $1
LIMIT 1;

-- name: FindEventSubscriptionByEventAndUserID :one
SELECT id, event_id, user_id, created_at
FROM event_subscriptions
WHERE event_id = $1 AND user_id = $2
LIMIT 1;

-- name: ListEventSubscriptionsByUserID :many
SELECT id, event_id, user_id, created_at
FROM event_subscriptions
WHERE user_id = $1
ORDER BY created_at;

-- name: DeleteEventSubscription :exec
DELETE FROM event_subscriptions
WHERE id = $1;
