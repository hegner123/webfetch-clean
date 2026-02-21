package main

import (
	"fmt"
	"os"
)

// ReadFile reads content from a local file.
// Returns the file content as a string and any error encountered.
func ReadFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Check file info
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist: %s", path)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Reject directories
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Check for empty content
	if len(data) == 0 {
		return "", fmt.Errorf("file is empty: %s", path)
	}

	return string(data), nil
}
