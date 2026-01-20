-- name: DeleteChirpWithID :exec
DELETE FROM chirps
WHERE id = $1;