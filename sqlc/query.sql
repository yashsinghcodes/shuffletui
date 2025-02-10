-- name: GetAllUsers :many
SELECT * FROM sessions;

-- name: InsertSession :one
INSERT INTO sessions (
    sshkey, username
) VALUES (
    ?, ?
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM sessions WHERE sshkey=?;

-- name: GetApiKey :one
SELECT apikey FROM sessions WHERE sshkey=?;
