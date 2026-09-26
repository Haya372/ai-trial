-- name: InsertEventShare :one
INSERT INTO event_shares (id, event_id, token_hash, expires_at)
VALUES (
    sqlc.arg('id'),
    sqlc.arg('event_id'),
    sqlc.arg('token_hash'),
    sqlc.arg('expires_at')
)
RETURNING id, event_id, token_hash, expires_at, created_at;

-- name: FindEventShareByTokenHash :one
SELECT id, event_id, token_hash, expires_at, created_at
FROM event_shares
WHERE token_hash = $1
LIMIT 1;
