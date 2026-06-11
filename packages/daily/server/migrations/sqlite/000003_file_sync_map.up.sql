-- 000003_file_sync_map.up.sql
-- Table to map local file paths to their synced memo UUIDs and content hashes
CREATE TABLE IF NOT EXISTS workspace_file_sync (
    file_path    TEXT NOT NULL PRIMARY KEY,
    memo_uuid    TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (memo_uuid) REFERENCES memo(memo_uuid) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workspace_file_sync_memo_uuid ON workspace_file_sync(memo_uuid);
