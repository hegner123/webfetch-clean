package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchURL_ValidURL(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent is set
		if !strings.Contains(r.Header.Get("User-Agent"), "webfetch-clean") {
			t.Errorf("Expected User-Agent to contain 'webfetch-clean', got: %s", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	// Test fetching
	content, finalURL, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(content, "Test content") {
		t.Errorf("Expected content to contain 'Test content', got: %s", content)
	}

	if finalURL != server.URL {
		t.Errorf("Expected finalURL to be %s, got: %s", server.URL, finalURL)
	}
}

func TestFetchURL_EmptyURL(t *testing.T) {
	_, _, err := FetchURL(context.Background(), "", 30)
	if err == nil {
		t.Fatal("Expected error for empty URL, got nil")
	}

	if !strings.Contains(err.Error(), "URL cannot be empty") {
		t.Errorf("Expected 'URL cannot be empty' error, got: %v", err)
	}
}

func TestFetchURL_404Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, _, err := FetchURL(context.Background(), server.URL, 30)
	if err == nil {
		t.Fatal("Expected error for 404 status, got nil")
	}

	if !strings.Contains(err.Error(), "not found or forbidden") {
		t.Errorf("Expected '404' error message, got: %v", err)
	}
}

func TestFetchURL_500Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, _, err := FetchURL(context.Background(), server.URL, 30)
	if err == nil {
		t.Fatal("Expected error for 500 status, got nil")
	}

	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' message, got: %v", err)
	}
}

func TestFetchURL_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Don't write any content
	}))
	defer server.Close()

	_, _, err := FetchURL(context.Background(), server.URL, 30)
	if err == nil {
		t.Fatal("Expected error for empty content, got nil")
	}

	if !strings.Contains(err.Error(), "no content received") {
		t.Errorf("Expected 'no content received' error, got: %v", err)
	}
}

func TestFetchURL_Timeout(t *testing.T) {
	// Create a server that sleeps longer than the timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.Write([]byte("too late"))
	}))
	defer server.Close()

	// Use a short timeout
	_, _, err := FetchURL(context.Background(), server.URL, 1)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// The error should mention timeout or deadline
	errMsg := err.Error()
	if !strings.Contains(errMsg, "timeout") && !strings.Contains(errMsg, "deadline") && !strings.Contains(errMsg, "context") {
		t.Errorf("Expected timeout-related error, got: %v", err)
	}
}

func TestFetchURL_InvalidURL(t *testing.T) {
	_, _, err := FetchURL(context.Background(), "not-a-valid-url", 30)
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}

func TestFetchURL_Headers(t *testing.T) {
	// Test that proper headers are set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check all expected headers
		headers := map[string]string{
			"User-Agent":      "webfetch-clean",
			"Accept":          "text/html",
			"Accept-Language": "en-US",
		}

		for header, expectedSubstring := range headers {
			value := r.Header.Get(header)
			if !strings.Contains(value, expectedSubstring) {
				t.Errorf("Expected %s header to contain '%s', got: %s", header, expectedSubstring, value)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	_, _, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestFetchURL_RedirectTracking(t *testing.T) {
	// Create a server that redirects
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/original" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Final content</body></html>"))
	}))
	defer redirectServer.Close()

	// Test fetching with redirect
	content, finalURL, err := FetchURL(context.Background(), redirectServer.URL+"/original", 30)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(content, "Final content") {
		t.Errorf("Expected content to contain 'Final content', got: %s", content)
	}

	expectedFinalURL := redirectServer.URL + "/final"
	if finalURL != expectedFinalURL {
		t.Errorf("Expected finalURL to be %s, got: %s", expectedFinalURL, finalURL)
	}
}

func TestFetchURL_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte("too late"))
	}))
	defer server.Close()

	// Create a context that we cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, _, err := FetchURL(ctx, server.URL, 30)
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	// The error should indicate context was canceled
	errMsg := err.Error()
	if !strings.Contains(errMsg, "context canceled") && !strings.Contains(errMsg, "cancel") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

// Tests for ValidateURLScheme function
func TestValidateURLScheme_ValidHTTP(t *testing.T) {
	err := ValidateURLScheme("http://example.com")
	if err != nil {
		t.Errorf("Expected no error for http:// URL, got: %v", err)
	}
}

func TestValidateURLScheme_ValidHTTPS(t *testing.T) {
	err := ValidateURLScheme("https://example.com")
	if err != nil {
		t.Errorf("Expected no error for https:// URL, got: %v", err)
	}
}

func TestValidateURLScheme_FTPRejected(t *testing.T) {
	err := ValidateURLScheme("ftp://example.com")
	if err == nil {
		t.Fatal("Expected error for ftp:// URL, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Errorf("Expected 'unsupported URL scheme' error, got: %v", err)
	}
}

func TestValidateURLScheme_FileRejected(t *testing.T) {
	err := ValidateURLScheme("file:///etc/passwd")
	if err == nil {
		t.Fatal("Expected error for file:// URL, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Errorf("Expected 'unsupported URL scheme' error, got: %v", err)
	}
}

func TestValidateURLScheme_NoScheme(t *testing.T) {
	err := ValidateURLScheme("example.com")
	if err == nil {
		t.Fatal("Expected error for URL without scheme, got nil")
	}
	if !strings.Contains(err.Error(), "must include scheme") {
		t.Errorf("Expected 'must include scheme' error, got: %v", err)
	}
}

func TestValidateURLScheme_JavascriptRejected(t *testing.T) {
	err := ValidateURLScheme("javascript:alert('xss')")
	if err == nil {
		t.Fatal("Expected error for javascript: URL, got nil")
	}
}

func TestValidateURLScheme_DataRejected(t *testing.T) {
	err := ValidateURLScheme("data:text/html,<script>alert('xss')</script>")
	if err == nil {
		t.Fatal("Expected error for data: URL, got nil")
	}
}

// Tests for default timeout behavior
func TestFetchURL_DefaultTimeout(t *testing.T) {
	requestReceived := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived <- true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	// Use timeout of 0 - should use default (30s)
	_, _, err := FetchURL(context.Background(), server.URL, 0)
	if err != nil {
		t.Fatalf("Expected no error with default timeout, got: %v", err)
	}

	select {
	case <-requestReceived:
		// Good - request was made
	default:
		t.Fatal("Request was not made to server")
	}
}

func TestFetchURL_NegativeTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	// Use negative timeout - should use default (30s)
	_, _, err := FetchURL(context.Background(), server.URL, -5)
	if err != nil {
		t.Fatalf("Expected no error with negative timeout (should use default), got: %v", err)
	}
}

// Tests for response size limit
func TestFetchURL_ResponseSizeLimit(t *testing.T) {
	// Create a server that returns more than MaxResponseSize
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write slightly more than MaxResponseSize (50MB + 1KB)
		// We'll simulate this by writing a lot of data
		largeData := make([]byte, MaxResponseSize+1024)
		for i := range largeData {
			largeData[i] = 'x'
		}
		w.Write(largeData)
	}))
	defer server.Close()

	_, _, err := FetchURL(context.Background(), server.URL, 60)
	if err == nil {
		t.Fatal("Expected error for oversized response, got nil")
	}

	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' error, got: %v", err)
	}
}

// Test that URL scheme validation is called during FetchURL
func TestFetchURL_RejectsInvalidScheme(t *testing.T) {
	_, _, err := FetchURL(context.Background(), "ftp://example.com/file.txt", 30)
	if err == nil {
		t.Fatal("Expected error for ftp:// URL, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Errorf("Expected 'unsupported URL scheme' error, got: %v", err)
	}
}
