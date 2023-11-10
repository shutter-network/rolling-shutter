-- name: InsertMeta :exec
INSERT INTO meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT value FROM meta_inf WHERE key = $1;
