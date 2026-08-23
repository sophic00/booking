-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    full_name,
    phone,
    role
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUserProfile :one
UPDATE users
SET full_name = $2,
    phone = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

