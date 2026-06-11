-- name: CreateResource :exec
INSERT INTO resource (id, memo_uuid, owner_user_id, filename, hash, size, mime_type, internal_path)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetResourceByID :one
SELECT id, memo_uuid, owner_user_id, filename, hash, size, mime_type, internal_path, created_at
FROM resource WHERE id = ? AND owner_user_id = ? LIMIT 1;

-- name: ListResourcesByMemoUUID :many
SELECT id, memo_uuid, owner_user_id, filename, hash, size, mime_type, internal_path, created_at
FROM resource WHERE memo_uuid = ? AND owner_user_id = ?;

-- name: ListAllResources :many
SELECT id, memo_uuid, owner_user_id, filename, hash, size, mime_type, internal_path, created_at
FROM resource WHERE owner_user_id = ?;

-- name: ListTrackedPaths :many
SELECT internal_path
FROM resource;

-- name: LinkResourceToMemo :exec
UPDATE resource SET memo_uuid = ? WHERE id = ? AND owner_user_id = ?;

-- name: ResourceExistsForOwner :one
SELECT COUNT(1)
FROM resource
WHERE id = ? AND owner_user_id = ?;

-- name: UnlinkMemoResourcesByOwner :execrows
UPDATE resource
SET memo_uuid = NULL
WHERE memo_uuid = ? AND owner_user_id = ?;

-- name: ListMemoResourcesBatch :many
SELECT id, memo_uuid, filename, hash, size, mime_type, internal_path, created_at
FROM resource
WHERE memo_uuid IN (sqlc.slice('memo_uuids'))
ORDER BY memo_uuid ASC, created_at ASC;

-- name: CleanupOrphanResources :execrows
DELETE FROM resource
WHERE memo_uuid IS NULL 
  AND id NOT IN (
      SELECT DISTINCT json_each.value
      FROM memo_history, json_each(memo_history.resource_ids)
  );
