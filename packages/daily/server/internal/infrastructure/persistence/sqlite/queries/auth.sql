-- name: CreateAuthUser :one
INSERT INTO auth_user (username, password_hash, role, status)
VALUES (?, ?, ?, ?)
RETURNING id, created_at, updated_at;

-- name: GetAuthUserByUsername :one
SELECT id, username, password_hash, role, status, created_at, updated_at
FROM auth_user
WHERE username = ?;

-- name: GetAuthUserByID :one
SELECT id, username, password_hash, role, status, created_at, updated_at
FROM auth_user
WHERE id = ?;

-- name: CountAuthUsers :one
SELECT COUNT(1) FROM auth_user;

-- name: DeleteAuthUserByID :execrows
DELETE FROM auth_user
WHERE id = ?;

-- name: CreateAuthSession :exec
INSERT INTO auth_session (
    id, user_id, session_token_hash, remember_token_hash,
    expires_at, remember_expires_at, user_agent, client_ip
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAuthSessionBySessionTokenHash :one
SELECT id, user_id, session_token_hash, remember_token_hash,
       expires_at, remember_expires_at, revoked_at,
       user_agent, client_ip, created_at, updated_at
FROM auth_session
WHERE session_token_hash = ?;

-- name: GetAuthSessionByRememberTokenHash :one
SELECT id, user_id, session_token_hash, remember_token_hash,
       expires_at, remember_expires_at, revoked_at,
       user_agent, client_ip, created_at, updated_at
FROM auth_session
WHERE remember_token_hash = ?;

-- name: UpdateAuthSessionTokensAndExpiry :execrows
UPDATE auth_session
SET session_token_hash = ?,
    remember_token_hash = ?,
    expires_at = ?,
    remember_expires_at = ?
WHERE id = ? AND revoked_at IS NULL;

-- name: RevokeAuthSessionByID :execrows
UPDATE auth_session
SET revoked_at = ?
WHERE id = ? AND revoked_at IS NULL;

-- name: DeleteExpiredAuthSessions :execrows
DELETE FROM auth_session
WHERE remember_expires_at <= ? OR revoked_at IS NOT NULL;

-- name: CreateAuthApiKey :exec
INSERT INTO auth_api_key (id, user_id, key_hash, label, expires_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetAuthApiKeyByKeyHash :one
SELECT id, user_id, key_hash, label, created_at, expires_at, last_used_at
FROM auth_api_key
WHERE key_hash = ?;

-- name: ListAuthApiKeysByUserID :many
SELECT id, user_id, key_hash, label, created_at, expires_at, last_used_at
FROM auth_api_key
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: UpdateAuthApiKeyLastUsedAt :execrows
UPDATE auth_api_key
SET last_used_at = ?
WHERE id = ?;

-- name: DeleteAuthApiKey :execrows
DELETE FROM auth_api_key
WHERE id = ? AND user_id = ?;

-- name: CreateAuthInvite :exec
INSERT INTO auth_invite (id, code_hash, role, expires_at, created_by)
VALUES (?, ?, ?, ?, ?);

-- name: GetAuthInviteByCodeHash :one
SELECT id, code_hash, role, expires_at, used_at, used_by, created_by, revoked_at, created_at
FROM auth_invite
WHERE code_hash = ?;

-- name: ListAuthInvitesByCreator :many
SELECT id, code_hash, role, expires_at, used_at, used_by, created_by, revoked_at, created_at
FROM auth_invite
WHERE created_by = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListActiveAuthInvitesByCreator :many
SELECT id, code_hash, role, expires_at, used_at, used_by, created_by, revoked_at, created_at
FROM auth_invite
WHERE created_by = ?
  AND used_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListUsedAuthInvitesByCreator :many
SELECT id, code_hash, role, expires_at, used_at, used_by, created_by, revoked_at, created_at
FROM auth_invite
WHERE created_by = ?
  AND used_at IS NOT NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListRevokedAuthInvitesByCreator :many
SELECT id, code_hash, role, expires_at, used_at, used_by, created_by, revoked_at, created_at
FROM auth_invite
WHERE created_by = ?
  AND revoked_at IS NOT NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListExpiredAuthInvitesByCreator :many
SELECT id, code_hash, role, expires_at, used_at, used_by, created_by, revoked_at, created_at
FROM auth_invite
WHERE created_by = ?
  AND used_at IS NULL
  AND revoked_at IS NULL
  AND expires_at <= ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountActiveAuthInvitesByCreatorAndRole :one
SELECT COUNT(1)
FROM auth_invite
WHERE created_by = ?
  AND role = ?
  AND used_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > ?;

-- name: MarkAuthInviteUsed :execrows
UPDATE auth_invite
SET used_at = ?, used_by = ?
WHERE id = ? AND used_at IS NULL AND revoked_at IS NULL;

-- name: RevokeAuthInvite :execrows
UPDATE auth_invite
SET revoked_at = ?
WHERE id = ? AND revoked_at IS NULL;
