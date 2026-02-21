package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchBrowser_EmptyURL(t *testing.T) {
	_, _, err := FetchBrowser(context.Background(), "", 30)
	if err == nil {
		t.Fatal("Expected error for empty URL, got nil")
	}
	if !strings.Contains(err.Error(), "URL cannot be empty") {
		t.Errorf("Expected 'URL cannot be empty' error, got: %v", err)
	}
}

func TestFetchBrowser_InvalidScheme(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "ftp scheme",
			url:  "ftp://example.com",
			want: "unsupported URL scheme",
		},
		{
			name: "file scheme",
			url:  "file:///etc/passwd",
			want: "unsupported URL scheme",
		},
		{
			name: "no scheme",
			url:  "example.com",
			want: "must include scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := FetchBrowser(context.Background(), tt.url, 30)
			if err == nil {
				t.Fatalf("Expected error for URL %q, got nil", tt.url)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Expected error containing %q, got: %v", tt.want, err)
			}
		})
	}
}

func TestFetchBrowser_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	// Server that delays forever
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(30 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, _, err := FetchBrowser(ctx, server.URL, 30)
	if err == nil {
		t.Fatal("Expected error from cancelled context, got nil")
	}
}

func TestFetchBrowser_SPAContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	// Serve a page where content is rendered by JavaScript
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>SPA Test</title></head>
<body>
<div id="app"></div>
<script>
document.getElementById('app').innerHTML = '<p>JavaScript rendered content</p>';
</script>
</body>
</html>`)
	}))
	defer server.Close()

	content, finalURL, err := FetchBrowser(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(content, "JavaScript rendered content") {
		t.Errorf("Expected JS-rendered content in output, got: %s", content)
	}

	if finalURL == "" {
		t.Error("Expected non-empty final URL")
	}
}

func TestFetchBrowser_DefaultTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><p>OK</p></body></html>`)
	}))
	defer server.Close()

	// timeout=0 should use DefaultTimeout
	content, _, err := FetchBrowser(context.Background(), server.URL, 0)
	if err != nil {
		t.Fatalf("Expected no error with default timeout, got: %v", err)
	}

	if !strings.Contains(content, "OK") {
		t.Errorf("Expected content to contain 'OK', got: %s", content)
	}
}
