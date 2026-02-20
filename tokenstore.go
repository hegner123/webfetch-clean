package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/hegner123/webfetch-clean/db"
	_ "modernc.org/sqlite"
)

// schemaSQL is the DDL used to initialize the token database.
const schemaSQL = `CREATE TABLE IF NOT EXISTS file_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    file TEXT NOT NULL,
    expires INTEGER NOT NULL,
    consumed BOOLEAN NOT NULL DEFAULT FALSE
);`

// TokenStore manages single-use file access tokens backed by SQLite.
type TokenStore struct {
	queries *db.Queries
	sqlDB   *sql.DB
}

// NewTokenStore opens a SQLite database at dbPath, enables WAL mode and busy
// timeout, runs the schema migration, and returns a ready TokenStore.
func NewTokenStore(dbPath string) (*TokenStore, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open token database: %w", err)
	}

	// WAL mode allows concurrent reads during writes
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Busy timeout: retry for 5 seconds instead of returning SQLITE_BUSY
	if _, err := sqlDB.Exec("PRAGMA busy_timeout=5000"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// Run schema migration
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &TokenStore{
		queries: db.New(sqlDB),
		sqlDB:   sqlDB,
	}, nil
}

// CreateFileToken generates a single-use UUID token for filePath that expires
// after ttl. The file must exist and be readable.
func (ts *TokenStore) CreateFileToken(filePath string, ttl time.Duration) (string, error) {
	// Validate file exists and is readable
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("file not accessible: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	token := generateUUID()
	expires := time.Now().Add(ttl).Unix()

	_, err = ts.queries.CreateToken(context.Background(), db.CreateTokenParams{
		Token:   token,
		File:    filePath,
		Expires: expires,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

// RedeemToken atomically marks the token as consumed and returns the file path.
// Returns an error if the token is not found, already consumed, or expired.
func (ts *TokenStore) RedeemToken(token string) (string, error) {
	filePath, err := ts.queries.RedeemToken(context.Background(), token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("token not found, expired, or already consumed")
		}
		return "", fmt.Errorf("failed to redeem token: %w", err)
	}
	return filePath, nil
}

// Cleanup deletes expired and consumed token rows.
func (ts *TokenStore) Cleanup() error {
	return ts.queries.DeleteExpired(context.Background())
}

// Close closes the underlying SQLite connection.
func (ts *TokenStore) Close() error {
	return ts.sqlDB.Close()
}

// generateUUID returns a RFC 4122 v4 UUID string using crypto/rand.
func generateUUID() string {
	var uuid [16]byte
	// crypto/rand.Read always returns len(p) and a nil error on supported platforms
	_, err := rand.Read(uuid[:])
	if err != nil {
		// This should never happen on modern systems
		panic(fmt.Sprintf("crypto/rand.Read failed: %v", err))
	}
	// Set version 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// Set variant bits
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
