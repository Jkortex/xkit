-- name: CreateMemoHistory :exec
INSERT INTO memo_history (id, memo_uuid, owner_user_id, content, tags, resource_ids)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListMemoHistory :many
SELECT id, memo_uuid, content, tags, resource_ids, created_at
FROM memo_history
WHERE memo_uuid = ? AND owner_user_id = ?
ORDER BY created_at DESC;

-- name: GetMemoHistoryByID :one
SELECT id, memo_uuid, content, tags, resource_ids, created_at
FROM memo_history
WHERE id = ? AND owner_user_id = ?;

-- name: DeleteOldMemoHistory :execrows
DELETE FROM memo_history
WHERE memo_history.memo_uuid = ? AND memo_history.id NOT IN (
    SELECT h.id FROM memo_history h
    WHERE h.memo_uuid = ?
    ORDER BY h.created_at DESC
    LIMIT ?
);
