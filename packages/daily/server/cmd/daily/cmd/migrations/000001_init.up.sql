-- 000001_init.up.sql
-- Daily Server SQLite Initial Schema

-- 1. 核心笔记表
CREATE TABLE IF NOT EXISTS memo (
    memo_uuid    TEXT NOT NULL PRIMARY KEY,
    content      TEXT NOT NULL,
    row_status   TEXT NOT NULL DEFAULT 'normal',
    owner_user_id INTEGER NOT NULL DEFAULT 0,
    expires_at   DATETIME,
    search_text  TEXT, -- 专门用于全文检索的聚合文本
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_memo_owner_user_id ON memo(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_memo_row_status_created_at ON memo(row_status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memo_row_status_updated_at ON memo(row_status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_memo_expires_at ON memo(expires_at) WHERE expires_at IS NOT NULL;

-- 自动更新 updated_at 的触发器
CREATE TRIGGER IF NOT EXISTS trg_memo_updated_at
AFTER UPDATE ON memo
BEGIN
    UPDATE memo SET updated_at = CURRENT_TIMESTAMP WHERE memo_uuid = NEW.memo_uuid;
END;

-- 2. 标签表
CREATE TABLE IF NOT EXISTS tag (
    name       TEXT NOT NULL PRIMARY KEY,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. 笔记-标签关联表
CREATE TABLE IF NOT EXISTS memo_tag (
    memo_uuid TEXT NOT NULL,
    tag_name  TEXT NOT NULL,
    PRIMARY KEY (memo_uuid, tag_name),
    FOREIGN KEY (memo_uuid) REFERENCES memo(memo_uuid) ON DELETE CASCADE,
    FOREIGN KEY (tag_name) REFERENCES tag(name) ON DELETE CASCADE
);

-- 4. 标签别名
CREATE TABLE IF NOT EXISTS tag_alias (
    alias_name     TEXT NOT NULL PRIMARY KEY,
    canonical_name TEXT NOT NULL,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (canonical_name) REFERENCES tag(name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tag_alias_canonical_name ON tag_alias(canonical_name);

-- 5. 标签治理审计
CREATE TABLE IF NOT EXISTS tag_governance_audit (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    action         TEXT NOT NULL,
    summary        TEXT NOT NULL,
    affected_memos INTEGER NOT NULL DEFAULT 0,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tag_governance_audit_created_at ON tag_governance_audit(created_at DESC);

-- 6. 资源附件表
CREATE TABLE IF NOT EXISTS resource (
    id            TEXT NOT NULL PRIMARY KEY,
    memo_uuid     TEXT,
    owner_user_id INTEGER NOT NULL DEFAULT 0,
    filename      TEXT NOT NULL,
    hash          TEXT NOT NULL,
    size          INTEGER NOT NULL,
    mime_type     TEXT NOT NULL,
    internal_path TEXT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (memo_uuid) REFERENCES memo(memo_uuid) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_resource_owner_user_id ON resource(owner_user_id);

-- 7. 笔记历史快照
CREATE TABLE IF NOT EXISTS memo_history (
    id            TEXT NOT NULL PRIMARY KEY,
    memo_uuid     TEXT NOT NULL,
    owner_user_id INTEGER NOT NULL,
    content       TEXT NOT NULL,
    tags          TEXT NOT NULL DEFAULT '[]', -- JSON 存储为 TEXT
    resource_ids  TEXT NOT NULL DEFAULT '[]',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (memo_uuid) REFERENCES memo(memo_uuid) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_memo_history_memo_uuid_created_at ON memo_history(memo_uuid, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memo_history_owner_user_id ON memo_history(owner_user_id);

-- 8. 用户表
CREATE TABLE IF NOT EXISTS auth_user (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'member',
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS trg_auth_user_updated_at
AFTER UPDATE ON auth_user
BEGIN
    UPDATE auth_user SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 9. 会话表
CREATE TABLE IF NOT EXISTS auth_session (
    id                   TEXT NOT NULL PRIMARY KEY,
    user_id              INTEGER NOT NULL,
    session_token_hash   TEXT NOT NULL UNIQUE,
    remember_token_hash  TEXT NOT NULL UNIQUE,
    expires_at           DATETIME NOT NULL,
    remember_expires_at  DATETIME NOT NULL,
    revoked_at           DATETIME,
    user_agent           TEXT NOT NULL DEFAULT '',
    client_ip            TEXT NOT NULL DEFAULT '',
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auth_session_user_id ON auth_session(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_session_expires_at ON auth_session(expires_at);
CREATE INDEX IF NOT EXISTS idx_auth_session_remember_expires ON auth_session(remember_expires_at);

CREATE TRIGGER IF NOT EXISTS trg_auth_session_updated_at
AFTER UPDATE ON auth_session
BEGIN
    UPDATE auth_session SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 10. 邀请码表
CREATE TABLE IF NOT EXISTS auth_invite (
    id         TEXT NOT NULL PRIMARY KEY,
    code_hash  TEXT NOT NULL UNIQUE,
    role       TEXT NOT NULL DEFAULT 'member',
    expires_at DATETIME NOT NULL,
    used_at    DATETIME,
    used_by    INTEGER,
    created_by INTEGER,
    revoked_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (used_by) REFERENCES auth_user(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES auth_user(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_invite_expires_at ON auth_invite(expires_at);
CREATE INDEX IF NOT EXISTS idx_auth_invite_revoked_at ON auth_invite(revoked_at);

-- 11. API Key 表
CREATE TABLE IF NOT EXISTS auth_api_key (
    id           TEXT NOT NULL PRIMARY KEY,
    user_id      INTEGER NOT NULL,
    key_hash     TEXT NOT NULL UNIQUE,
    label        TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at   DATETIME,
    last_used_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES auth_user(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_auth_api_key_user_id ON auth_api_key(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_api_key_hash ON auth_api_key(key_hash);
CREATE INDEX IF NOT EXISTS idx_auth_api_key_expires_at ON auth_api_key(expires_at);

-- 12. FTS5 虚拟表与同步触发器
CREATE VIRTUAL TABLE IF NOT EXISTS memo_fts USING fts5(
    memo_uuid UNINDEXED,
    content,
    tokenize='unicode61'
);

CREATE TRIGGER IF NOT EXISTS trg_memo_fts_insert AFTER INSERT ON memo
BEGIN
    INSERT INTO memo_fts(memo_uuid, content) VALUES (NEW.memo_uuid, NEW.search_text);
END;

CREATE TRIGGER IF NOT EXISTS trg_memo_fts_delete AFTER DELETE ON memo
BEGIN
    DELETE FROM memo_fts WHERE memo_uuid = OLD.memo_uuid;
END;

CREATE TRIGGER IF NOT EXISTS trg_memo_fts_update AFTER UPDATE ON memo
BEGIN
    DELETE FROM memo_fts WHERE memo_uuid = OLD.memo_uuid;
    INSERT INTO memo_fts(memo_uuid, content) VALUES (NEW.memo_uuid, NEW.search_text);
END;
