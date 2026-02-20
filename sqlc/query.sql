-- name: CreateToken :one
INSERT INTO file_tokens (token, file, expires, consumed)
VALUES (?, ?, ?, FALSE)
RETURNING *;

-- name: RedeemToken :one
-- Atomic single-use redemption: marks consumed and returns file path in one statement.
-- Prevents TOCTOU race where two concurrent requests could both redeem the same token.
-- Returns no rows if token is invalid, expired, or already consumed.
UPDATE file_tokens SET consumed = TRUE
WHERE token = ? AND consumed = FALSE AND expires > unixepoch('now')
RETURNING file;

-- name: DeleteExpired :exec
DELETE FROM file_tokens WHERE expires < unixepoch('now') OR consumed = TRUE;

-- name: CountActiveTokens :one
SELECT COUNT(*) FROM file_tokens WHERE consumed = FALSE AND expires > unixepoch('now');
