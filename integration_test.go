package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFullPipeline_FetchCleanConvert tests the complete pipeline:
// Fetch URL -> Clean HTML -> Convert to Markdown
func TestFullPipeline_FetchCleanConvert(t *testing.T) {
	// Create test server with complex HTML
	html := `
    <html>
        <head>
            <title>Test Page</title>
            <script>tracking();</script>
            <style>.foo { color: red; }</style>
        </head>
        <body>
            <nav>Navigation</nav>
            <aside class="sidebar">Sidebar content</aside>
            <div class="advertisement">Ad content</div>
            <main>
                <h1>Main Title</h1>
                <p>This is the main content.</p>
                <ul>
                    <li>Item 1</li>
                    <li>Item 2</li>
                </ul>
            </main>
            <footer>Footer</footer>
        </body>
    </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Step 1: Fetch
	fetched, _, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Step 2: Clean
	cleaned, err := CleanHTML(fetched, false, false, false, "clean")
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Verify cleaning removed unwanted elements
	if strings.Contains(cleaned, "<script") {
		t.Error("Script should be removed")
	}
	if strings.Contains(cleaned, "<style") {
		t.Error("Style should be removed")
	}
	if strings.Contains(cleaned, "<nav") {
		t.Error("Nav should be removed")
	}
	if strings.Contains(cleaned, "advertisement") {
		t.Error("Ad should be removed")
	}
	if strings.Contains(cleaned, "Sidebar content") {
		t.Error("Sidebar should be removed")
	}
	if strings.Contains(cleaned, "Footer") {
		t.Error("Footer should be removed")
	}

	// Verify main content is preserved
	if !strings.Contains(cleaned, "Main Title") {
		t.Error("Main title should be preserved")
	}
	if !strings.Contains(cleaned, "main content") {
		t.Error("Main content should be preserved")
	}

	// Step 3: Convert to Markdown
	markdown, err := ConvertToMarkdown(cleaned)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify markdown conversion
	if !strings.Contains(markdown, "Main Title") {
		t.Error("Title should be in markdown")
	}
	if !strings.Contains(markdown, "main content") {
		t.Error("Content should be in markdown")
	}
	if !strings.Contains(markdown, "Item 1") {
		t.Error("List items should be in markdown")
	}

	// Should not contain HTML tags in markdown output
	htmlTags := []string{"<div", "<main", "<p>", "</p>"}
	for _, tag := range htmlTags {
		if strings.Contains(markdown, tag) {
			t.Errorf("Markdown should not contain HTML tag: %s", tag)
		}
	}
}

// TestFullPipeline_PreserveMainOnly tests the pipeline with preserveMainOnly option
func TestFullPipeline_PreserveMainOnly(t *testing.T) {
	html := `
    <html>
        <body>
            <header>Header content</header>
            <nav>Navigation</nav>
            <main>
                <h1>Article Title</h1>
                <p>Article content</p>
            </main>
            <aside>Sidebar</aside>
            <footer>Footer</footer>
        </body>
    </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Fetch
	fetched, _, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Clean with preserveMainOnly
	cleaned, err := CleanHTML(fetched, true, false, false, "clean")
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Should only have main content
	if !strings.Contains(cleaned, "Article Title") {
		t.Error("Article title should be preserved")
	}
	if !strings.Contains(cleaned, "Article content") {
		t.Error("Article content should be preserved")
	}

	// Should not have other content
	if strings.Contains(cleaned, "Header content") {
		t.Error("Header should be removed")
	}
	if strings.Contains(cleaned, "Navigation") {
		t.Error("Navigation should be removed")
	}
	if strings.Contains(cleaned, "Sidebar") {
		t.Error("Sidebar should be removed")
	}

	// Convert to markdown
	markdown, err := ConvertToMarkdown(cleaned)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if !strings.Contains(markdown, "Article Title") {
		t.Error("Title should be in final markdown")
	}
}

// TestFullPipeline_RemoveImages tests the pipeline with removeImages option
func TestFullPipeline_RemoveImages(t *testing.T) {
	html := `
    <html>
        <body>
            <h1>Gallery</h1>
            <img src="photo1.jpg" alt="Photo 1">
            <p>Description</p>
            <img src="photo2.jpg" alt="Photo 2">
        </body>
    </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Fetch
	fetched, _, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Clean with removeImages
	cleaned, err := CleanHTML(fetched, false, true, false, "clean")
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Should not have images
	if strings.Contains(cleaned, "<img") {
		t.Error("Images should be removed")
	}

	// Should have text content
	if !strings.Contains(cleaned, "Gallery") {
		t.Error("Text content should be preserved")
	}
	if !strings.Contains(cleaned, "Description") {
		t.Error("Text content should be preserved")
	}

	// Convert to markdown
	markdown, err := ConvertToMarkdown(cleaned)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Markdown should not have image syntax
	if strings.Contains(markdown, "![") {
		t.Error("Markdown should not contain image syntax when images removed")
	}
}

// TestFullPipeline_HTMLOutput tests the pipeline with HTML output format
func TestFullPipeline_HTMLOutput(t *testing.T) {
	html := `
    <html>
        <head><script>alert('hi');</script></head>
        <body>
            <h1>Title</h1>
            <p>Content</p>
        </body>
    </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Fetch
	fetched, _, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Clean
	cleaned, err := CleanHTML(fetched, false, false, false, "clean")
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Convert to HTML format (should return as-is)
	output, err := ConvertToFormat(cleaned, "html")
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Should still be HTML
	if !strings.Contains(output, "<h1") {
		t.Error("HTML output should contain HTML tags")
	}
	if !strings.Contains(output, "Title") {
		t.Error("HTML output should contain content")
	}
	if strings.Contains(output, "<script") {
		t.Error("HTML output should have cleaned content (no script)")
	}
}

// TestFullPipeline_ComplexPage tests a realistic complex page
func TestFullPipeline_ComplexPage(t *testing.T) {
	html := `
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="UTF-8">
            <title>Complex Blog Post</title>
            <script src="analytics.js"></script>
            <script>trackPageView();</script>
            <style>
                body { font-family: Arial; }
                .ad { background: yellow; }
            </style>
            <link rel="stylesheet" href="styles.css">
        </head>
        <body>
            <header class="site-header">
                <div class="logo">Logo</div>
                <nav class="main-nav">
                    <a href="/">Home</a>
                    <a href="/about">About</a>
                </nav>
            </header>

            <div class="ad-banner">
                <img src="ad.jpg" alt="Advertisement">
            </div>

            <aside class="sidebar">
                <div class="widget">
                    <h3>Recent Posts</h3>
                    <ul>
                        <li><a href="/post1">Post 1</a></li>
                    </ul>
                </div>
                <div class="advertisement">
                    <script>showAd();</script>
                </div>
            </aside>

            <main class="content">
                <article>
                    <h1>How to Clean HTML</h1>
                    <p class="meta">Posted on January 13, 2026</p>

                    <p>This is the main article content. It contains <strong>important information</strong> about cleaning HTML.</p>

                    <h2>Why Clean HTML?</h2>
                    <p>Here are the reasons:</p>
                    <ul>
                        <li>Remove clutter</li>
                        <li>Improve readability</li>
                        <li>Better processing</li>
                    </ul>

                    <h2>Code Example</h2>
                    <pre><code>const clean = (html) => {
  return html.replace(/<script.*?<\/script>/g, '');
};</code></pre>

                    <p>For more information, visit <a href="https://example.com">our website</a>.</p>

                    <img src="diagram.png" alt="Diagram showing the process">
                </article>
            </main>

            <div class="social-share">
                <button onclick="share()">Share</button>
            </div>

            <div class="comments-section">
                <h3>Comments</h3>
                <div class="comment">Comment 1</div>
            </div>

            <footer class="site-footer">
                <p>&copy; 2026 Example Site</p>
                <div class="footer-links">
                    <a href="/privacy">Privacy</a>
                    <a href="/terms">Terms</a>
                </div>
            </footer>

            <div class="cookie-notice">
                This site uses cookies.
            </div>

            <div class="modal" id="signup-modal">
                <form>Subscribe to newsletter</form>
            </div>
        </body>
    </html>
    `

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Fetch
	fetched, _, err := FetchURL(context.Background(), server.URL, 30)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Clean
	cleaned, err := CleanHTML(fetched, false, false, false, "clean")
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	// Verify unwanted elements are removed
	// Note: We check for script/style tags, not the literal string "<script"
	// because the content includes a code example with that text
	unwantedContent := []string{
		"<script src=", "analytics.js", "trackPageView",
		"<style>", "ad-banner", "Advertisement", "sidebar", "Recent Posts",
		"social-share", "Share", "comments-section", "Comment 1",
		"site-footer", "cookie-notice", "modal", "Subscribe",
	}

	for _, content := range unwantedContent {
		if strings.Contains(cleaned, content) {
			t.Errorf("Unwanted content should be removed: %s", content)
		}
	}

	// Verify main content is preserved
	wantedContent := []string{
		"How to Clean HTML",
		"main article content",
		"important information",
		"Why Clean HTML",
		"Remove clutter",
		"Improve readability",
		"Better processing",
		"Code Example",
		"const clean",
		"example.com",
	}

	for _, content := range wantedContent {
		if !strings.Contains(cleaned, content) {
			t.Errorf("Wanted content should be preserved: %s", content)
		}
	}

	// Convert to markdown
	markdown, err := ConvertToMarkdown(cleaned)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify markdown has main content
	for _, content := range wantedContent {
		if !strings.Contains(markdown, content) {
			t.Errorf("Markdown should contain: %s", content)
		}
	}

	// Verify markdown syntax
	markdownFeatures := []string{"#", "**", "`", "-", "["}
	foundFeatures := 0
	for _, feature := range markdownFeatures {
		if strings.Contains(markdown, feature) {
			foundFeatures++
		}
	}

	if foundFeatures < 3 {
		t.Error("Expected markdown to have proper markdown formatting")
	}
}
