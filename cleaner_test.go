package main

import (
	"strings"
	"testing"
)

func TestCleanHTML_EmptyHTML(t *testing.T) {
	_, err := CleanHTML("", false, false, false)
	if err == nil {
		t.Fatal("Expected error for empty HTML, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

func TestCleanHTML_RemoveScriptStyleNav(t *testing.T) {
	html := `
	<html>
		<head><title>Test</title></head>
		<script>alert('hello');</script>
		<style>.foo { color: red; }</style>
		<body>
			<nav>Navigation</nav>
			<main>Content</main>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not contain script, style, or nav
	if strings.Contains(cleaned, "<script") {
		t.Error("Expected <script> to be removed")
	}
	if strings.Contains(cleaned, "<style") {
		t.Error("Expected <style> to be removed")
	}
	if strings.Contains(cleaned, "<nav") {
		t.Error("Expected <nav> to be removed")
	}
	if strings.Contains(cleaned, "<head") {
		t.Error("Expected <head> to be removed")
	}

	// Should still contain main content
	if !strings.Contains(cleaned, "Content") {
		t.Error("Expected main content to be preserved")
	}
}

func TestCleanHTML_RemoveAds(t *testing.T) {
	html := `
	<html>
		<body>
			<div class="advertisement">Ad here</div>
			<div id="ad-banner">Banner ad</div>
			<div class="ad-container">Another ad</div>
			<p>Real content</p>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not contain ad-related content
	if strings.Contains(cleaned, "Ad here") {
		t.Error("Expected advertisement div to be removed")
	}
	if strings.Contains(cleaned, "Banner ad") {
		t.Error("Expected ad-banner div to be removed")
	}
	if strings.Contains(cleaned, "Another ad") {
		t.Error("Expected ad-container div to be removed")
	}

	// Should preserve real content
	if !strings.Contains(cleaned, "Real content") {
		t.Error("Expected real content to be preserved")
	}
}

func TestCleanHTML_RemoveClutter(t *testing.T) {
	html := `
	<html>
		<body>
			<aside>Sidebar</aside>
			<footer>Footer</footer>
			<div class="sidebar">Side content</div>
			<div class="popup">Popup</div>
			<div class="modal">Modal</div>
			<div class="cookie-notice">Cookie notice</div>
			<main>Main content</main>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not contain clutter
	if strings.Contains(cleaned, "Sidebar") {
		t.Error("Expected aside to be removed")
	}
	if strings.Contains(cleaned, "Footer") {
		t.Error("Expected footer to be removed")
	}
	if strings.Contains(cleaned, "Side content") {
		t.Error("Expected sidebar div to be removed")
	}
	if strings.Contains(cleaned, "Popup") {
		t.Error("Expected popup to be removed")
	}
	if strings.Contains(cleaned, "Modal") {
		t.Error("Expected modal to be removed")
	}
	if strings.Contains(cleaned, "Cookie notice") {
		t.Error("Expected cookie notice to be removed")
	}

	// Should preserve main content
	if !strings.Contains(cleaned, "Main content") {
		t.Error("Expected main content to be preserved")
	}
}

func TestCleanHTML_RemoveIframes(t *testing.T) {
	html := `
	<html>
		<body>
			<iframe src="tracking.html"></iframe>
			<p>Content</p>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if strings.Contains(cleaned, "<iframe") {
		t.Error("Expected iframe to be removed")
	}

	if !strings.Contains(cleaned, "Content") {
		t.Error("Expected content to be preserved")
	}
}

func TestCleanHTML_PreserveMainOnly(t *testing.T) {
	html := `
	<html>
		<body>
			<header>Header</header>
			<main>Main content</main>
			<footer>Footer</footer>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, true, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only contain main content
	if !strings.Contains(cleaned, "Main content") {
		t.Error("Expected main content to be preserved")
	}

	// Should not contain header or footer
	if strings.Contains(cleaned, "Header") {
		t.Error("Expected header to be removed when preserveMainOnly=true")
	}
	if strings.Contains(cleaned, "Footer") {
		t.Error("Expected footer to be removed when preserveMainOnly=true")
	}
}

func TestCleanHTML_PreserveArticle(t *testing.T) {
	html := `
	<html>
		<body>
			<header>Header</header>
			<article>Article content</article>
			<footer>Footer</footer>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, true, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should contain article content
	if !strings.Contains(cleaned, "Article content") {
		t.Error("Expected article content to be preserved")
	}

	// Should not contain header or footer
	if strings.Contains(cleaned, "Header") {
		t.Error("Expected header to be removed when preserveMainOnly=true")
	}
	if strings.Contains(cleaned, "Footer") {
		t.Error("Expected footer to be removed when preserveMainOnly=true")
	}
}

func TestCleanHTML_RemoveImages(t *testing.T) {
	html := `
	<html>
		<body>
			<img src="photo.jpg" alt="Photo">
			<p>Text content</p>
			<img src="logo.png">
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, true, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not contain images
	if strings.Contains(cleaned, "<img") {
		t.Error("Expected images to be removed when removeImages=true")
	}

	// Should preserve text content
	if !strings.Contains(cleaned, "Text content") {
		t.Error("Expected text content to be preserved")
	}
}

func TestCleanHTML_KeepImages(t *testing.T) {
	html := `
	<html>
		<body>
			<img src="photo.jpg" alt="Photo">
			<p>Text content</p>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should contain images when removeImages=false
	if !strings.Contains(cleaned, "<img") {
		t.Error("Expected images to be preserved when removeImages=false")
	}

	// Should have src and alt attributes (semantic attributes)
	if !strings.Contains(cleaned, `src="photo.jpg"`) {
		t.Error("Expected img src attribute to be preserved")
	}
	if !strings.Contains(cleaned, `alt="Photo"`) {
		t.Error("Expected img alt attribute to be preserved")
	}
}

func TestCleanHTML_StripAttributes(t *testing.T) {
	html := `
	<html>
		<body>
			<div class="container" id="main" style="color: red;" data-foo="bar">
				<a href="https://example.com" class="link" onclick="alert()">Link</a>
			</div>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not contain non-semantic attributes
	if strings.Contains(cleaned, `class=`) {
		t.Error("Expected class attributes to be removed")
	}
	if strings.Contains(cleaned, `id=`) {
		t.Error("Expected id attributes to be removed")
	}
	if strings.Contains(cleaned, `style=`) {
		t.Error("Expected style attributes to be removed")
	}
	if strings.Contains(cleaned, `data-`) {
		t.Error("Expected data attributes to be removed")
	}
	if strings.Contains(cleaned, `onclick=`) {
		t.Error("Expected onclick attributes to be removed")
	}

	// Should preserve href (semantic attribute)
	if !strings.Contains(cleaned, `href="https://example.com"`) {
		t.Error("Expected href attribute to be preserved")
	}

	// Should preserve content
	if !strings.Contains(cleaned, "Link") {
		t.Error("Expected link text to be preserved")
	}
}

func TestCleanHTML_PreserveSemanticElements(t *testing.T) {
	html := `
	<html>
		<body>
			<h1>Heading 1</h1>
			<h2>Heading 2</h2>
			<p>Paragraph</p>
			<ul><li>List item</li></ul>
			<ol><li>Ordered item</li></ol>
			<code>code</code>
			<pre>preformatted</pre>
			<table><tr><td>Cell</td></tr></table>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should preserve all semantic elements
	semanticElements := []string{
		"<h1", "<h2", "<p", "<ul", "<ol", "<li",
		"<code", "<pre", "<table", "<tr", "<td",
		"Heading 1", "Heading 2", "Paragraph", "List item",
		"Ordered item", "code", "preformatted", "Cell",
	}

	for _, element := range semanticElements {
		if !strings.Contains(cleaned, element) {
			t.Errorf("Expected semantic element/content '%s' to be preserved", element)
		}
	}
}

func TestCleanHTML_AdDetectionLimitations(t *testing.T) {
	// Test documents known limitations with ad detection.
	// The ad detection is intentionally aggressive - it matches patterns like "ad-", "-ad-", etc.
	// This means some legitimate content with "ad" in class names may be removed.
	// Examples: "thread-card" matches "-ad-", "reader-mode" matches "-ad-"
	//
	// This is acceptable as:
	// 1. It's better to be overly aggressive in removing ads
	// 2. Most legitimate content doesn't have "ad" surrounded by dashes
	// 3. The main semantic content is usually not affected
	html := `
	<html>
		<body>
			<div class="advertisement">Real ad</div>
			<div class="ad-banner">Banner ad</div>
			<div class="content">Main content</div>
		</body>
	</html>
	`

	cleaned, err := CleanHTML(html, false, false, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify ads are removed
	if strings.Contains(cleaned, "Real ad") {
		t.Error("Expected advertisement to be removed")
	}
	if strings.Contains(cleaned, "Banner ad") {
		t.Error("Expected ad-banner to be removed")
	}

	// Verify main content is preserved
	if !strings.Contains(cleaned, "Main content") {
		t.Error("Expected main content to be preserved")
	}
}
