-- name: UpdateUserPwdEmailByToken :one
UPDATE users
SET email = $1, hashed_password = $2
WHERE id = $3  --ID get from checking JWT Token before parsing 
RETURNING email;