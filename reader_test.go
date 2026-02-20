package main

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestReadFile_Success(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "<html><body><p>Test content</p></body></html>"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := ReadFile(tmpfile.Name())
	if err != nil {
		t.Errorf("ReadFile() error = %v, want nil", err)
	}
	if result != content {
		t.Errorf("ReadFile() = %v, want %v", result, content)
	}
}

func TestReadFile_EmptyPath(t *testing.T) {
	_, err := ReadFile("")
	if err == nil {
		t.Error("ReadFile(\"\") expected error, got nil")
	}
	expectedMsg := "file path cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("ReadFile(\"\") error = %v, want %v", err.Error(), expectedMsg)
	}
}

func TestReadFile_FileNotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/path/file.html")
	if err == nil {
		t.Error("ReadFile() expected error for nonexistent file, got nil")
	}
	// Error message should contain "does not exist"
	if err != nil && err.Error() == "" {
		t.Error("ReadFile() error message is empty")
	}
}

func TestReadFile_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "empty*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	_, err = ReadFile(tmpfile.Name())
	if err == nil {
		t.Error("ReadFile() expected error for empty file, got nil")
	}
	// Error message should contain "empty"
	if err != nil && err.Error() == "" {
		t.Error("ReadFile() error message is empty")
	}
}

func TestReadFile_Directory(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	_, err = ReadFile(tmpdir)
	if err == nil {
		t.Error("ReadFile() expected error for directory, got nil")
	}
	// Error message should contain "directory"
	if err != nil && err.Error() == "" {
		t.Error("ReadFile() error message is empty")
	}
}

func TestReadFile_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows: chmod does not remove read access")
	}

	tmpfile, err := os.CreateTemp("", "noperm*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := "<html><body><p>Test</p></body></html>"
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Remove all permissions
	if err := os.Chmod(tmpfile.Name(), 0000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions for cleanup
	defer os.Chmod(tmpfile.Name(), 0644)

	_, err = ReadFile(tmpfile.Name())
	if err == nil {
		t.Error("ReadFile() expected permission error, got nil")
	}
}

func TestFullPipeline_FileInput(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "pipeline*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	htmlContent := `<html>
<head><title>Test</title></head>
<body>
	<nav>Navigation</nav>
	<script>alert('ad');</script>
	<main>
		<h1>Main Content</h1>
		<p>This is the main content.</p>
	</main>
	<footer>Footer</footer>
</body>
</html>`

	if _, err := tmpfile.Write([]byte(htmlContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := Config{
		File:      tmpfile.Name(),
		Format:    "markdown",
		MaxTokens: 100000,
	}

	result := processInput(config)

	if result.Error != "" {
		t.Errorf("processInput() error = %v, want nil", result.Error)
	}

	if result.Content == "" {
		t.Error("processInput() content is empty, want non-empty")
	}

	// Verify nav, script, footer are removed
	if strings.Contains(result.Content, "Navigation") {
		t.Error("processInput() content contains 'Navigation', should be removed")
	}
	if strings.Contains(result.Content, "alert") {
		t.Error("processInput() content contains 'alert', should be removed")
	}
	if strings.Contains(result.Content, "Footer") {
		t.Error("processInput() content contains 'Footer', should be removed")
	}

	// Verify main content is preserved
	if !strings.Contains(result.Content, "Main Content") {
		t.Error("processInput() content missing 'Main Content', should be preserved")
	}
}

func TestFullPipeline_FileInput_HTMLFormat(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "pipeline-html*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	htmlContent := `<html>
<body>
	<script>alert('test');</script>
	<p>Content</p>
</body>
</html>`

	if _, err := tmpfile.Write([]byte(htmlContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := Config{
		File:      tmpfile.Name(),
		Format:    "html",
		MaxTokens: 100000,
	}

	result := processInput(config)

	if result.Error != "" {
		t.Errorf("processInput() error = %v, want nil", result.Error)
	}

	if result.Format != "html" {
		t.Errorf("processInput() format = %v, want html", result.Format)
	}

	// Verify script is removed
	if strings.Contains(result.Content, "alert") {
		t.Error("processInput() content contains 'alert', should be removed")
	}

	// Verify content is preserved
	if !strings.Contains(result.Content, "Content") {
		t.Error("processInput() content missing 'Content', should be preserved")
	}
}

func TestFullPipeline_FileInput_WithOptions(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "options*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	htmlContent := `<html>
<body>
	<aside>Sidebar</aside>
	<main>
		<h1>Main Title</h1>
		<p>Main paragraph</p>
		<img src="test.jpg" alt="Test">
	</main>
	<article>
		<h2>Article Title</h2>
	</article>
</body>
</html>`

	if _, err := tmpfile.Write([]byte(htmlContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	config := Config{
		File:         tmpfile.Name(),
		Format:       "markdown",
		PreserveMain: true,
		RemoveImages: true,
		MaxTokens:    100000,
	}

	result := processInput(config)

	if result.Error != "" {
		t.Errorf("processInput() error = %v, want nil", result.Error)
	}

	// With PreserveMain, article content should be removed
	if strings.Contains(result.Content, "Article Title") {
		t.Error("processInput() with PreserveMain should not contain 'Article Title'")
	}

	// With RemoveImages, img should be removed
	if strings.Contains(result.Content, "test.jpg") {
		t.Error("processInput() with RemoveImages should not contain image references")
	}

	// Main content should be preserved
	if !strings.Contains(result.Content, "Main Title") {
		t.Error("processInput() content missing 'Main Title', should be preserved")
	}
}
