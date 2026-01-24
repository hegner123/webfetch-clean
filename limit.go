package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateOutputFilename creates a filename based on URL or file path.
// For URLs: https://example.com/foo/bar -> example-com-foo-bar.md
// For files: /path/to/file.html -> file-cleaned.md
func GenerateOutputFilename(url, filePath, format string) string {
	var basename string
	var ext string

	if format == "html" {
		ext = ".html"
	} else {
		ext = ".md"
	}

	if filePath != "" {
		// File input mode
		base := filepath.Base(filePath)
		extLen := len(filepath.Ext(base))
		if extLen > 0 {
			basename = base[:len(base)-extLen]
		} else {
			basename = base
		}
		return basename + "-cleaned" + ext
	}

	if url != "" {
		// URL input mode
		// Remove scheme
		basename = url
		basename = strings.ReplaceAll(basename, "https://", "")
		basename = strings.ReplaceAll(basename, "http://", "")
		// Replace special characters with hyphens
		basename = strings.ReplaceAll(basename, "/", "-")
		basename = strings.ReplaceAll(basename, ".", "-")
		basename = strings.ReplaceAll(basename, "?", "-")
		basename = strings.ReplaceAll(basename, "&", "-")
		basename = strings.ReplaceAll(basename, "=", "-")
		// Clean up multiple hyphens and trailing hyphens
		for strings.Contains(basename, "--") {
			basename = strings.ReplaceAll(basename, "--", "-")
		}
		basename = strings.Trim(basename, "-")
		return basename + ext
	}

	return "output" + ext
}

// GenerateUniqueFilename ensures the filename is unique by appending -1, -2, etc.
// if the file already exists. Returns the first non-existing filename and any error.
// Returns an error if unable to check file existence (permission denied, etc.).
func GenerateUniqueFilename(baseName string) (string, error) {
	// Check if file exists
	_, err := os.Stat(baseName)
	if err != nil {
		if os.IsNotExist(err) {
			return baseName, nil
		}
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

	// File exists, find unique name
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]

	// Try adding -1, -2, etc. until we find a unique name
	for counter := 1; counter <= 10000; counter++ {
		candidate := fmt.Sprintf("%s-%d%s", nameWithoutExt, counter, ext)
		_, err := os.Stat(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				return candidate, nil
			}
			return "", fmt.Errorf("failed to check file existence: %w", err)
		}
	}

	return "", fmt.Errorf("could not generate unique filename after 10000 attempts: %s", baseName)
}
