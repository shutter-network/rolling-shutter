-- name: InsertMeta :exec
INSERT INTO meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT value FROM meta_inf WHERE key = $1;

-- name: UpdateMeta :exec
UPDATE meta_inf SET value = $1 WHERE key = $2;