package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchURL fetches the content from the given URL with the specified timeout.
// Returns the HTML content as a string and any error encountered.
func FetchURL(url string, timeout int) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to identify ourselves
	req.Header.Set("User-Agent", "webfetch-clean/1.0 (HTML cleaning tool)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode >= 500 {
		return "", fmt.Errorf("server error (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("page not found or forbidden (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: HTTP %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return "", fmt.Errorf("no content received from URL")
	}

	return string(body), nil
}
