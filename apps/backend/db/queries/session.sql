-- name: InsertSession :one
INSERT INTO sessions (user_id, expires_at)
VALUES ($1, $2)
RETURNING *;

-- name: FindSessionByID :one
SELECT id, user_id, expires_at, created_at FROM sessions
WHERE id = $1
LIMIT 1;

-- name: FindActiveSessionByID :one
SELECT
    s.id,
    s.user_id,
    s.expires_at,
    s.created_at,
    u.email,
    u.display_name
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.id = $1 AND s.expires_at > NOW()
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= NOW();
