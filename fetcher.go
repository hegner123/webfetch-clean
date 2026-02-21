package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultTimeout is the default HTTP request timeout in seconds
const DefaultTimeout = 30

// MaxResponseSize is the maximum allowed response body size (50 MB).
// This prevents memory exhaustion from unexpectedly large responses.
const MaxResponseSize = 50 * 1024 * 1024

// ValidateURLScheme checks that the URL has a valid HTTP or HTTPS scheme.
// Returns an error if the URL is malformed or uses an unsupported scheme.
func ValidateURLScheme(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		if scheme == "" {
			return fmt.Errorf("URL must include scheme (http:// or https://)")
		}
		return fmt.Errorf("unsupported URL scheme %q: only http and https are allowed", scheme)
	}

	return nil
}

// FetchURL fetches the content from the given URL with the specified timeout.
// Returns the HTML content as a string, the final URL after any redirects, and any error encountered.
func FetchURL(ctx context.Context, url string, timeout int) (content string, finalURL string, err error) {
	if url == "" {
		return "", "", fmt.Errorf("URL cannot be empty")
	}

	// Validate URL scheme (only http and https allowed)
	if err := ValidateURLScheme(url); err != nil {
		return "", "", err
	}

	// Use default timeout if not specified or invalid
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to identify ourselves
	req.Header.Set("User-Agent", fmt.Sprintf("webfetch-clean/%s (HTML cleaning tool)", Version))
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Capture final URL after any redirects
	finalURL = resp.Request.URL.String()

	// Check HTTP status code
	if resp.StatusCode >= 500 {
		return "", finalURL, fmt.Errorf("server error (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return "", finalURL, fmt.Errorf("page not found or forbidden (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return "", finalURL, fmt.Errorf("unexpected status code: HTTP %d", resp.StatusCode)
	}

	// Read response body with size limit to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", finalURL, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return "", finalURL, fmt.Errorf("no content received from URL")
	}

	// Check if we exceeded the size limit
	if len(body) > MaxResponseSize {
		return "", finalURL, fmt.Errorf("response too large: exceeds %d bytes limit", MaxResponseSize)
	}

	return string(body), finalURL, nil
}
