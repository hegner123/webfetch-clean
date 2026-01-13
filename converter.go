package main

import (
	"fmt"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// ConvertToFormat converts HTML to the specified output format (html or markdown).
// For "html", it returns the HTML as-is.
// For "markdown", it converts HTML to Markdown format.
func ConvertToFormat(html string, format string) (string, error) {
	if html == "" {
		return "", fmt.Errorf("HTML content cannot be empty")
	}

	switch format {
	case "html":
		// Return HTML as-is
		return html, nil

	case "markdown":
		// Convert HTML to Markdown
		return ConvertToMarkdown(html)

	default:
		return "", fmt.Errorf("unsupported format: %s (use 'html' or 'markdown')", format)
	}
}

// ConvertToMarkdown converts HTML content to Markdown format.
func ConvertToMarkdown(html string) (string, error) {
	if html == "" {
		return "", fmt.Errorf("HTML content cannot be empty")
	}

	// Create converter
	converter := md.NewConverter("", true, nil)

	// Convert HTML to Markdown
	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	return markdown, nil
}
