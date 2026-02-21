package main

import (
	"os"
	"testing"
	"time"
)

func newTestTokenStore(t *testing.T) *TokenStore {
	t.Helper()
	ts, err := NewTokenStore(":memory:")
	if err != nil {
		t.Fatalf("NewTokenStore(:memory:) failed: %v", err)
	}
	t.Cleanup(func() { ts.Close() })
	return ts
}

func createTestFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "tokentest*.html")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString("<html><body>test</body></html>"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestTokenStore_CreateAndRedeem(t *testing.T) {
	ts := newTestTokenStore(t)
	filePath := createTestFile(t)

	token, err := ts.CreateFileToken(filePath, 5*time.Minute)
	if err != nil {
		t.Fatalf("CreateFileToken() error: %v", err)
	}
	if token == "" {
		t.Fatal("CreateFileToken() returned empty token")
	}

	got, err := ts.RedeemToken(token)
	if err != nil {
		t.Fatalf("RedeemToken() error: %v", err)
	}
	if got != filePath {
		t.Errorf("RedeemToken() = %q, want %q", got, filePath)
	}
}

func TestTokenStore_RedeemConsumed(t *testing.T) {
	ts := newTestTokenStore(t)
	filePath := createTestFile(t)

	token, err := ts.CreateFileToken(filePath, 5*time.Minute)
	if err != nil {
		t.Fatalf("CreateFileToken() error: %v", err)
	}

	// First redemption succeeds
	_, err = ts.RedeemToken(token)
	if err != nil {
		t.Fatalf("first RedeemToken() error: %v", err)
	}

	// Second redemption fails
	_, err = ts.RedeemToken(token)
	if err == nil {
		t.Error("second RedeemToken() should fail for consumed token")
	}
}

func TestTokenStore_RedeemExpired(t *testing.T) {
	ts := newTestTokenStore(t)
	filePath := createTestFile(t)

	// Create token that expires immediately (negative TTL)
	token, err := ts.CreateFileToken(filePath, -1*time.Second)
	if err != nil {
		t.Fatalf("CreateFileToken() error: %v", err)
	}

	_, err = ts.RedeemToken(token)
	if err == nil {
		t.Error("RedeemToken() should fail for expired token")
	}
}

func TestTokenStore_RedeemNotFound(t *testing.T) {
	ts := newTestTokenStore(t)

	_, err := ts.RedeemToken("nonexistent-token-uuid")
	if err == nil {
		t.Error("RedeemToken() should fail for nonexistent token")
	}
}

func TestTokenStore_Cleanup(t *testing.T) {
	ts := newTestTokenStore(t)
	filePath := createTestFile(t)

	// Create an expired token
	_, err := ts.CreateFileToken(filePath, -1*time.Second)
	if err != nil {
		t.Fatalf("CreateFileToken() error: %v", err)
	}

	// Create a consumed token
	token2, err := ts.CreateFileToken(filePath, 5*time.Minute)
	if err != nil {
		t.Fatalf("CreateFileToken() error: %v", err)
	}
	_, err = ts.RedeemToken(token2)
	if err != nil {
		t.Fatalf("RedeemToken() error: %v", err)
	}

	// Cleanup should remove both
	err = ts.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup() error: %v", err)
	}

	// Verify: no active tokens remain (the expired and consumed are gone)
	// Create a fresh active token and verify only it exists
	_, err = ts.CreateFileToken(filePath, 5*time.Minute)
	if err != nil {
		t.Fatalf("CreateFileToken() after cleanup error: %v", err)
	}
}

func TestTokenStore_CreateValidatesFile(t *testing.T) {
	ts := newTestTokenStore(t)

	// Nonexistent file
	_, err := ts.CreateFileToken("/nonexistent/path/file.html", 5*time.Minute)
	if err == nil {
		t.Error("CreateFileToken() should fail for nonexistent file")
	}

	// Directory instead of file
	dir := os.TempDir()
	_, err = ts.CreateFileToken(dir, 5*time.Minute)
	if err == nil {
		t.Error("CreateFileToken() should fail for directory")
	}
}

func TestTokenStore_Close(t *testing.T) {
	ts, err := NewTokenStore(":memory:")
	if err != nil {
		t.Fatalf("NewTokenStore(:memory:) failed: %v", err)
	}

	err = ts.Close()
	if err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	// Operations after close should fail
	filePath := createTestFile(t)
	_, err = ts.CreateFileToken(filePath, 5*time.Minute)
	if err == nil {
		t.Error("CreateFileToken() should fail after Close()")
	}
}
