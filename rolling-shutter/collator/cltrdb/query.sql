-- name: InsertMeta :exec
INSERT INTO collator.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT * FROM collator.meta_inf WHERE key = $1;
