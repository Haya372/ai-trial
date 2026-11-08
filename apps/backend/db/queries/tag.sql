-- name: InsertTag :one
INSERT INTO tags (id, user_id, name, color)
VALUES (sqlc.arg('id'), sqlc.arg('user_id'), sqlc.arg('name'), sqlc.arg('color'))
RETURNING id, user_id, name, color, created_at, updated_at;

-- name: FindTagByID :one
SELECT id, user_id, name, color, created_at, updated_at
FROM tags
WHERE id = $1
LIMIT 1;

-- name: UpdateTag :exec
UPDATE tags
SET name = $2,
    color = $3
WHERE id = $1;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1;

-- name: ExistsTagByName :one
-- LOWER(name) matches idx_tags_user_id_lower_name, so the check uses the
-- same case-insensitive comparison as the unique index.
SELECT EXISTS (
    SELECT 1
    FROM tags
    WHERE user_id = sqlc.arg('user_id')
      AND LOWER(name) = LOWER(sqlc.arg('name'))
      AND id <> sqlc.arg('exclude_id')
);

-- name: CountTagsByUserID :one
SELECT COUNT(*)
FROM tags
WHERE user_id = $1;

-- name: LockUserForUpdate :one
SELECT id
FROM users
WHERE id = $1
FOR UPDATE;

-- name: FindTagsByIDsForShare :many
SELECT id, user_id, name, color, created_at, updated_at
FROM tags
WHERE id = ANY(sqlc.arg('ids')::uuid[])
FOR SHARE;
