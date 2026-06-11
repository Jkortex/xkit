-- name: CreateMemo :one
INSERT INTO memo (memo_uuid, owner_user_id, content, row_status, expires_at, search_text)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING memo_uuid, content, row_status, expires_at, created_at, updated_at;

-- name: GetMemoByUUID :one
SELECT memo_uuid, content, row_status, expires_at, created_at, updated_at
FROM memo
WHERE memo_uuid = ? AND owner_user_id = ? LIMIT 1;

-- name: SearchMemos :many
SELECT m.memo_uuid, m.content, m.row_status, m.expires_at, m.created_at, m.updated_at
FROM memo m
JOIN memo_fts f ON m.memo_uuid = f.memo_uuid
WHERE f.content MATCH ?
AND m.owner_user_id = ?
AND m.row_status = ?
ORDER BY f.rank
LIMIT ? OFFSET ?;

-- name: ListAllMemos :many
SELECT memo_uuid, content, row_status, expires_at, created_at, updated_at
FROM memo 
WHERE row_status = ? AND owner_user_id = ?
ORDER BY created_at DESC 
LIMIT ? OFFSET ?;

-- name: ListMemosByTag :many
SELECT m.memo_uuid, m.content, m.row_status, m.expires_at, m.created_at, m.updated_at
FROM memo m
JOIN memo_tag mt ON m.memo_uuid = mt.memo_uuid
WHERE mt.tag_name = ?
AND m.owner_user_id = ?
AND m.row_status = ?
ORDER BY m.created_at DESC
LIMIT ? OFFSET ?;

-- name: ArchiveMemo :exec
UPDATE memo SET row_status = 'archived' WHERE memo_uuid = ? AND owner_user_id = ?;

-- name: ArchiveExpiredMemos :exec
UPDATE memo 
SET row_status = 'archived', updated_at = CURRENT_TIMESTAMP 
WHERE expires_at IS NOT NULL 
AND expires_at < CURRENT_TIMESTAMP 
AND row_status = 'normal';

-- name: UpdateMemoContent :exec
UPDATE memo 
SET content = ?
WHERE memo_uuid = ? AND owner_user_id = ?;

-- name: GetRandomMemo :one
SELECT memo_uuid, content, row_status, expires_at, created_at, updated_at
FROM memo 
WHERE row_status = 'normal' AND owner_user_id = ?
ORDER BY random() 
LIMIT 1;

-- name: CountMemos :one
SELECT COUNT(*) FROM memo WHERE row_status = 'normal' AND owner_user_id = ?;

-- name: CountTags :one
SELECT COUNT(DISTINCT tag_name) FROM memo_tag mt
JOIN memo m ON mt.memo_uuid = m.memo_uuid
WHERE m.owner_user_id = ? AND m.row_status = 'normal';

-- name: ListTagsWithCount :many
SELECT mt.tag_name as name, COUNT(mt.memo_uuid) as count
FROM memo_tag mt
JOIN memo m ON mt.memo_uuid = m.memo_uuid
WHERE m.owner_user_id = ? AND m.row_status = 'normal'
GROUP BY mt.tag_name
ORDER BY count DESC, name ASC;

-- name: GetDailyHeatmap :many
SELECT strftime('%Y-%m-%d', created_at) AS date, COUNT(*) AS count
FROM memo
WHERE row_status = 'normal' AND owner_user_id = ?
GROUP BY date
ORDER BY date ASC;

-- name: CountResources :one
SELECT COUNT(*) FROM resource WHERE owner_user_id = ?;

-- name: GetDailyStats :many
SELECT strftime('%Y-%m-%d', created_at) as date, COUNT(*) as count
FROM memo
WHERE row_status = 'normal' AND owner_user_id = ?
GROUP BY date
ORDER BY date DESC
LIMIT 365;

-- name: SaveTag :exec
INSERT INTO tag (name) VALUES (?) ON CONFLICT (name) DO NOTHING;

-- name: LinkMemoTag :exec
INSERT INTO memo_tag (memo_uuid, tag_name) VALUES (?, ?);

-- name: CleanupOrphanTags :exec
DELETE FROM tag 
WHERE name NOT IN (SELECT DISTINCT tag_name FROM memo_tag);

-- name: UpdateMemoRowStatus :execrows
UPDATE memo
SET row_status = ?
WHERE memo_uuid = ?;

-- name: DeleteMemoByUUID :execrows
DELETE FROM memo
WHERE memo_uuid = ? AND owner_user_id = ?;

-- name: ArchiveExpiredMemosBefore :execrows
UPDATE memo
SET row_status = 'archived', updated_at = ?
WHERE expires_at IS NOT NULL
  AND expires_at < ?
  AND row_status = 'normal';

-- name: UpdateMemoContentAndExpires :execrows
UPDATE memo
SET content = ?,
    expires_at = ?,
    search_text = ?
WHERE memo_uuid = ? AND owner_user_id = ?;

-- name: DeleteMemoTagsByMemoUUID :exec
DELETE FROM memo_tag
WHERE memo_uuid = ?;

-- name: GetMemoTags :many
SELECT tag_name
FROM memo_tag
WHERE memo_uuid = ?
ORDER BY tag_name ASC;

-- name: GetMemoTagsBatch :many
SELECT memo_uuid, tag_name
FROM memo_tag
WHERE memo_uuid IN (sqlc.slice('memo_uuids'))
ORDER BY memo_uuid ASC, tag_name ASC;

-- name: TagExists :one
SELECT COUNT(1)
FROM tag
WHERE name = ?;

-- name: ListExistingTagsByNames :many
SELECT name
FROM tag
WHERE name IN (sqlc.slice('tags'));

-- name: CountAffectedMemosByTags :one
SELECT COUNT(DISTINCT mt.memo_uuid)
FROM memo_tag mt
JOIN memo m ON mt.memo_uuid = m.memo_uuid
WHERE mt.tag_name IN (sqlc.slice('tags'))
  AND m.owner_user_id = ?;

-- name: DeleteDuplicateMemoTagLinksByOwner :execrows
DELETE FROM memo_tag
WHERE memo_tag.tag_name = ?
  AND memo_tag.memo_uuid IN (
      SELECT mt2.memo_uuid
      FROM memo_tag mt2
      JOIN memo m2 ON mt2.memo_uuid = m2.memo_uuid
      WHERE mt2.tag_name = ?
        AND m2.owner_user_id = ?
  );

-- name: MoveMemoTagLinksByOwner :execrows
UPDATE memo_tag
SET tag_name = ?
WHERE tag_name = ?
  AND memo_uuid IN (
      SELECT memo_uuid
      FROM memo
      WHERE owner_user_id = ?
  );

-- name: DeleteTagByName :execrows
DELETE FROM tag
WHERE name = ?;

-- name: UpsertTagAlias :exec
INSERT INTO tag_alias(alias_name, canonical_name)
VALUES(?, ?)
ON CONFLICT(alias_name) DO UPDATE
SET canonical_name = excluded.canonical_name;

-- name: DeleteTagAliasByName :execrows
DELETE FROM tag_alias
WHERE alias_name = ?;

-- name: ListTagAliases :many
SELECT alias_name, canonical_name
FROM tag_alias
ORDER BY alias_name ASC;

-- name: GetCanonicalTagAlias :one
SELECT canonical_name
FROM tag_alias
WHERE alias_name = ?
LIMIT 1;

-- name: AppendTagGovernanceAudit :exec
INSERT INTO tag_governance_audit(action, summary, affected_memos)
VALUES(?, ?, ?);

-- name: ListTagAudits :many
SELECT action, summary, affected_memos, created_at
FROM tag_governance_audit
WHERE (? = '' OR action = ?)
ORDER BY created_at DESC, id DESC
LIMIT ?;
