package main

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateOutputFilename_URL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		format   string
		expected string
	}{
		{
			name:     "simple URL with markdown",
			url:      "https://example.com",
			format:   "markdown",
			expected: "example-com.md",
		},
		{
			name:     "URL with path",
			url:      "https://example.com/foo/bar",
			format:   "markdown",
			expected: "example-com-foo-bar.md",
		},
		{
			name:     "URL with HTML format",
			url:      "https://example.com/test",
			format:   "html",
			expected: "example-com-test.html",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://example.com/foo/",
			format:   "markdown",
			expected: "example-com-foo.md",
		},
		{
			name:     "URL with query params",
			url:      "https://example.com/search?q=test&page=1",
			format:   "markdown",
			expected: "example-com-search-q-test-page-1.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateOutputFilename(tt.url, "", tt.format)
			if result != tt.expected {
				t.Errorf("GenerateOutputFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateOutputFilename_File(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		format   string
		expected string
	}{
		{
			name:     "simple filename",
			filePath: "test.html",
			format:   "markdown",
			expected: "test-cleaned.md",
		},
		{
			name:     "file with path",
			filePath: "/path/to/file.html",
			format:   "markdown",
			expected: "file-cleaned.md",
		},
		{
			name:     "file with HTML format",
			filePath: "myfile.html",
			format:   "html",
			expected: "myfile-cleaned.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateOutputFilename("", tt.filePath, tt.format)
			if result != tt.expected {
				t.Errorf("GenerateOutputFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessInput_OutputLimit_NotExceeded(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "small*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "<html><body><p>Small content</p></body></html>"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := Config{
		File:      tmpfile.Name(),
		Format:    "markdown",
		MaxTokens: 100000, // Much larger than content
	}

	result := processInput(config)

	if result.Error != "" {
		t.Errorf("processInput() error = %v, want nil", result.Error)
	}

	// Content should be returned directly
	if !strings.Contains(result.Content, "Small content") {
		t.Error("processInput() should return content directly when under limit")
	}

	// Should NOT contain "Output exceeded limit"
	if strings.Contains(result.Content, "Output exceeded limit") {
		t.Error("processInput() should not write to file when under limit")
	}
}

func TestProcessInput_OutputLimit_Exceeded(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "large*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create large content (> 300 bytes for 100 token limit)
	largeContent := "<html><body>"
	for i := 0; i < 100; i++ {
		largeContent += "<p>This is a very long paragraph with lots of content to exceed the token limit. "
		largeContent += "We need to make sure this content is large enough to trigger file writing.</p>"
	}
	largeContent += "</body></html>"

	if _, err := tmpfile.Write([]byte(largeContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := Config{
		File:      tmpfile.Name(),
		Format:    "markdown",
		MaxTokens: 100, // Very small limit (300 bytes)
	}

	result := processInput(config)

	if result.Error != "" {
		t.Errorf("processInput() error = %v, want nil", result.Error)
	}

	// processInput should signal over-limit without doing I/O
	if !result.OverLimit {
		t.Error("processInput() should set OverLimit = true when content exceeds limit")
	}

	if result.TokenCount <= 0 {
		t.Error("processInput() should set TokenCount > 0 when over limit")
	}

	if result.RawContent == "" {
		t.Error("processInput() should set RawContent when over limit")
	}

	if !strings.Contains(result.RawContent, "very long paragraph") {
		t.Error("processInput() RawContent should contain the cleaned content")
	}

	// Content should be empty — callers handle file-write
	if result.Content != "" {
		t.Errorf("processInput() Content should be empty when over limit, got: %s", result.Content)
	}
}

func TestProcessInput_OutputLimit_URLFilename(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create content that exceeds limit
	content := "<html><body>"
	for i := 0; i < 50; i++ {
		content += "<p>Content paragraph " + string(rune('0'+i%10)) + " with some text.</p>"
	}
	content += "</body></html>"

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := Config{
		File:      tmpfile.Name(),
		Format:    "markdown",
		MaxTokens: 50, // Small limit
	}

	result := processInput(config)

	if result.Error != "" {
		t.Errorf("processInput() error = %v, want nil", result.Error)
	}

	// processInput should signal over-limit
	if !result.OverLimit {
		t.Error("processInput() should set OverLimit = true")
	}

	if result.RawContent == "" {
		t.Error("processInput() should set RawContent when over limit")
	}
}
