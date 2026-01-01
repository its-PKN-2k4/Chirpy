-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;