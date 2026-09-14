-- name: ListEventsByUserAndDateRange :many
SELECT id, user_id, title, description, start_at, end_at, created_at, updated_at
FROM events
WHERE user_id = sqlc.arg('user_id')
  AND end_at > sqlc.arg('start_date')
  AND start_at < sqlc.arg('end_date')
ORDER BY start_at;
