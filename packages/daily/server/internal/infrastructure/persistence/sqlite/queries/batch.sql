-- name: BatchPrecheck :many
SELECT memo_uuid FROM memo WHERE memo_uuid IN (sqlc.slice('uuids')) AND owner_user_id = ?;

-- name: BatchArchive :many
UPDATE memo SET row_status = 'archived', updated_at = CURRENT_TIMESTAMP
WHERE memo_uuid IN (sqlc.slice('uuids')) AND owner_user_id = ?
RETURNING memo_uuid;

-- name: BatchDelete :many
DELETE FROM memo WHERE memo_uuid IN (sqlc.slice('uuids')) AND owner_user_id = ?
RETURNING memo_uuid;

-- name: BatchSaveTags :exec
INSERT OR IGNORE INTO tag (name)
SELECT value FROM json_each(sqlc.arg('tags'));

-- name: BatchTagAdd :exec
INSERT OR IGNORE INTO memo_tag (memo_uuid, tag_name)
SELECT u.value, t.value
FROM json_each(sqlc.arg('uuids')) AS u
CROSS JOIN json_each(sqlc.arg('tags')) AS t;

-- name: BatchTagRemove :exec
DELETE FROM memo_tag
WHERE memo_uuid IN (sqlc.slice('uuids')) AND tag_name IN (sqlc.slice('tags'));
