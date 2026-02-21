package main

import (
	"strings"
	"testing"
)

func TestConvertToMarkdown_EmptyHTML(t *testing.T) {
	_, err := ConvertToMarkdown("")
	if err == nil {
		t.Fatal("Expected error for empty HTML, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

func TestConvertToMarkdown_BasicHTML(t *testing.T) {
	html := `<html><body><p>Hello world</p></body></html>`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(markdown, "Hello world") {
		t.Errorf("Expected markdown to contain 'Hello world', got: %s", markdown)
	}
}

func TestConvertToMarkdown_Headings(t *testing.T) {
	html := `
	<html>
		<body>
			<h1>Heading 1</h1>
			<h2>Heading 2</h2>
			<h3>Heading 3</h3>
		</body>
	</html>
	`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check for markdown heading syntax
	if !strings.Contains(markdown, "# Heading 1") && !strings.Contains(markdown, "#Heading 1") {
		t.Error("Expected H1 to be converted to markdown heading")
	}
	if !strings.Contains(markdown, "## Heading 2") && !strings.Contains(markdown, "##Heading 2") {
		t.Error("Expected H2 to be converted to markdown heading")
	}
	if !strings.Contains(markdown, "### Heading 3") && !strings.Contains(markdown, "###Heading 3") {
		t.Error("Expected H3 to be converted to markdown heading")
	}
}

func TestConvertToMarkdown_Links(t *testing.T) {
	html := `<html><body><a href="https://example.com">Example Link</a></body></html>`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check for markdown link syntax [text](url)
	if !strings.Contains(markdown, "[Example Link]") {
		t.Error("Expected link text to be in markdown format")
	}
	if !strings.Contains(markdown, "https://example.com") {
		t.Error("Expected link URL to be preserved")
	}
}

func TestConvertToMarkdown_Lists(t *testing.T) {
	html := `
	<html>
		<body>
			<ul>
				<li>Item 1</li>
				<li>Item 2</li>
			</ul>
			<ol>
				<li>Numbered 1</li>
				<li>Numbered 2</li>
			</ol>
		</body>
	</html>
	`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check for list items
	if !strings.Contains(markdown, "Item 1") || !strings.Contains(markdown, "Item 2") {
		t.Error("Expected list items to be preserved")
	}
	if !strings.Contains(markdown, "Numbered 1") || !strings.Contains(markdown, "Numbered 2") {
		t.Error("Expected numbered list items to be preserved")
	}

	// Check for markdown list syntax (- or * for unordered, numbers for ordered)
	hasListMarker := strings.Contains(markdown, "- Item") ||
		strings.Contains(markdown, "* Item") ||
		strings.Contains(markdown, "-Item") ||
		strings.Contains(markdown, "*Item")
	if !hasListMarker {
		t.Error("Expected unordered list to have markdown list markers")
	}
}

func TestConvertToMarkdown_Code(t *testing.T) {
	html := `<html><body><code>const x = 5;</code></body></html>`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(markdown, "const x = 5") {
		t.Errorf("Expected code content to be preserved, got: %s", markdown)
	}

	// Check for inline code backticks
	if !strings.Contains(markdown, "`") {
		t.Error("Expected code to have markdown backticks")
	}
}

func TestConvertToMarkdown_PreformattedCode(t *testing.T) {
	html := `
	<html>
		<body>
			<pre><code>function test() {
  return true;
}</code></pre>
		</body>
	</html>
	`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(markdown, "function test") {
		t.Error("Expected code content to be preserved")
	}

	// Check for code block markers (triple backticks or indentation)
	hasCodeBlock := strings.Contains(markdown, "```") || strings.HasPrefix(strings.TrimSpace(markdown), "    ")
	if !hasCodeBlock {
		t.Error("Expected preformatted code to have markdown code block syntax")
	}
}

func TestConvertToMarkdown_Images(t *testing.T) {
	html := `<html><body><img src="photo.jpg" alt="Photo"></body></html>`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check for markdown image syntax ![alt](url)
	if !strings.Contains(markdown, "![") {
		t.Error("Expected image to have markdown image syntax")
	}
	if !strings.Contains(markdown, "photo.jpg") {
		t.Error("Expected image URL to be preserved")
	}
}

func TestConvertToMarkdown_Paragraphs(t *testing.T) {
	html := `
	<html>
		<body>
			<p>First paragraph</p>
			<p>Second paragraph</p>
		</body>
	</html>
	`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(markdown, "First paragraph") {
		t.Error("Expected first paragraph to be preserved")
	}
	if !strings.Contains(markdown, "Second paragraph") {
		t.Error("Expected second paragraph to be preserved")
	}
}

func TestConvertToFormat_HTML(t *testing.T) {
	html := `<html><body><p>Test</p></body></html>`

	result, err := ConvertToFormat(html, "html")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != html {
		t.Error("Expected HTML format to return HTML unchanged")
	}
}

func TestConvertToFormat_Markdown(t *testing.T) {
	html := `<html><body><h1>Title</h1></body></html>`

	result, err := ConvertToFormat(html, "markdown")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should be converted to markdown
	if strings.Contains(result, "<h1>") {
		t.Error("Expected HTML tags to be converted to markdown")
	}
	if !strings.Contains(result, "Title") {
		t.Error("Expected content to be preserved")
	}
}

func TestConvertToFormat_InvalidFormat(t *testing.T) {
	html := `<html><body><p>Test</p></body></html>`

	_, err := ConvertToFormat(html, "pdf")
	if err == nil {
		t.Fatal("Expected error for unsupported format, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestConvertToFormat_EmptyHTML(t *testing.T) {
	_, err := ConvertToFormat("", "markdown")
	if err == nil {
		t.Fatal("Expected error for empty HTML, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

func TestConvertToMarkdown_ComplexHTML(t *testing.T) {
	html := `
	<html>
		<body>
			<h1>Main Title</h1>
			<p>This is a paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
			<ul>
				<li>First item</li>
				<li>Second item</li>
			</ul>
			<a href="https://example.com">Link</a>
			<code>inline code</code>
		</body>
	</html>
	`

	markdown, err := ConvertToMarkdown(html)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all content is preserved
	requiredContent := []string{"Main Title", "paragraph", "bold", "italic", "First item", "Second item", "Link", "inline code"}
	for _, content := range requiredContent {
		if !strings.Contains(markdown, content) {
			t.Errorf("Expected markdown to contain '%s'", content)
		}
	}
}
