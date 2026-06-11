-- 000002_tag_set.up.sql
-- TagSet: 命名的标签组快速过滤

CREATE TABLE IF NOT EXISTS tag_set_group (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    name       TEXT NOT NULL,
    weight     INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tag_set_group_user_id ON tag_set_group(user_id);

CREATE TRIGGER IF NOT EXISTS trg_tag_set_group_updated_at
AFTER UPDATE ON tag_set_group
BEGIN
    UPDATE tag_set_group SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TABLE IF NOT EXISTS tag_set (
    id           TEXT NOT NULL PRIMARY KEY,
    user_id      INTEGER NOT NULL,
    group_id     TEXT,
    name         TEXT NOT NULL,
    tags_any     TEXT NOT NULL DEFAULT '[]',
    tags_all     TEXT NOT NULL DEFAULT '[]',
    tags_exclude TEXT NOT NULL DEFAULT '[]',
    weight       INTEGER NOT NULL DEFAULT 0,
    last_used_at DATETIME,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES tag_set_group(id) ON DELETE SET NULL,
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tag_set_user_id ON tag_set(user_id);
CREATE INDEX IF NOT EXISTS idx_tag_set_group_id ON tag_set(group_id);
CREATE INDEX IF NOT EXISTS idx_tag_set_weight ON tag_set(user_id, weight DESC);

CREATE TRIGGER IF NOT EXISTS trg_tag_set_updated_at
AFTER UPDATE ON tag_set
BEGIN
    UPDATE tag_set SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
