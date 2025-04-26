-- name: CreateChirp :one
INSERT INTO chirp (id, created_at, updated_at, body, user_id)
VALUES (
   gen_random_uuid(),
   NOW(),
   NOW(),
   $1,
   $2
   )
   RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirp
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirp 
WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirp 
WHERE id = $1;
