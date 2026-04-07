-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, expires_at, user_id, updated_at, created_at)
values ($1, $2, $3, NOW(), NOW())
RETURNING *;

-- name: GetRerfreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;


-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
set revoked_at = $2,
updated_at = NOW()
where token = $1;