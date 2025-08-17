-- name: CreateUser :one
INSERT INTO users (id, email, display_name, password)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET display_name = $2, password = $3, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteUser :one
UPDATE users SET user_status = 'inactive', updated_at = now() WHERE id = $1 RETURNING *;

-- name: PurgeInactiveUsers :exec
DELETE FROM users WHERE user_status = 'inactive' AND updated_at < NOW() - INTERVAL '30 days'; 