-- name: GetUserByRefreshToken :one
SELECT users.*
FROM users
LEFT JOIN refresh_tokens
ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1 
        AND refresh_tokens.revoked_at IS NULL 
        AND refresh_tokens.expires_at > NOW();
