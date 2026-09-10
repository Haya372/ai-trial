-- name: InsertUser :one
INSERT INTO users (email, display_name, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;
