-- name: ListEventsByUserAndDateRange :many
SELECT id, user_id, title, description, start_at, end_at, created_at, updated_at
FROM events
WHERE user_id = $1
  AND start_at < $3
  AND end_at > $2
ORDER BY start_at;
