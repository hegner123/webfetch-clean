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

	// Content should contain message about file
	if !strings.Contains(result.Content, "Output exceeded limit") {
		t.Error("processInput() should indicate output exceeded limit")
	}

	if !strings.Contains(result.Content, "Content written to file:") {
		t.Error("processInput() should indicate file was written")
	}

	// Extract filename from message
	var filename string
	parts := strings.Split(result.Content, "Content written to file: ")
	if len(parts) == 2 {
		filename = strings.TrimSpace(parts[1])
	}

	if filename == "" {
		t.Fatal("Failed to extract filename from result")
	}

	// Verify file was created
	defer os.Remove(filename)
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Expected file %s to be created, but got error: %v", filename, err)
	}

	// Verify file contains the cleaned content
	if !strings.Contains(string(fileContent), "very long paragraph") {
		t.Error("Output file should contain cleaned content")
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

	// Extract filename
	parts := strings.Split(result.Content, "Content written to file: ")
	if len(parts) != 2 {
		t.Fatal("Expected filename in result")
	}
	filename := strings.TrimSpace(parts[1])
	defer os.Remove(filename)

	// Verify filename format
	if !strings.HasSuffix(filename, "-cleaned.md") {
		t.Errorf("Filename should end with -cleaned.md, got: %s", filename)
	}
}
