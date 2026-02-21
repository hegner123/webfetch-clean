# webfetch-clean Test Sites

This directory contains HTML test sites designed to stress test the webfetch-clean tool under various conditions. Each site tests different aspects of the HTML cleaning pipeline.

## Test Sites Overview

### 1. heavy-ads/
**Purpose:** Test ad removal effectiveness

**Contains:**
- 20+ ad elements with various class/id patterns
- Inline ads mixed with content
- Native advertising disguised as content
- Banner ads, sidebar ads, footer ads
- Popup modal ads
- Cookie banners
- Tracking iframes

**What to verify:**
- All ad-related elements are removed
- Actual article content is preserved
- Main headings and paragraphs remain intact

**Expected output:**
- Article title and paragraphs
- Subsection headings
- Conclusion text
- NO ads, banners, or tracking elements

---

### 2. deep-nesting/
**Purpose:** Test DOM traversal with deeply nested elements

**Contains:**
- 20+ levels of nested div structures
- Content buried deep in the DOM tree
- Nested navigation menus
- Multiple wrapper/container layers

**What to verify:**
- Content is extracted regardless of nesting depth
- All wrapper divs are handled correctly
- Navigation and sidebars are removed
- Semantic content is preserved

**Expected output:**
- Main article heading and content
- Subsection with paragraphs
- List items
- NO navigation or sidebar content

---

### 3. large-content/
**Purpose:** Test performance with large file sizes

**Contains:**
- 50+ chapters of lorem ipsum text
- Thousands of words of content
- Multiple sections and headings
- Large HTML file size (100KB+)

**What to verify:**
- Tool handles large files without errors
- Processing completes in reasonable time
- All sections are preserved
- Memory usage remains acceptable

**Expected output:**
- All chapter headings
- All paragraph content
- Proper structure maintained

---

### 4. malformed-html/
**Purpose:** Test robustness with invalid HTML

**Contains:**
- Unclosed tags (p, div, span, script, style)
- Mismatched closing tags
- Orphaned closing tags
- Duplicate attributes
- Invalid nesting (block in inline)
- Incomplete attributes
- Multiple body/head tags
- Mixed content models

**What to verify:**
- Parser handles errors gracefully
- Tool doesn't crash on malformed input
- Valid content is still extracted
- Cleaning pipeline completes successfully

**Expected output:**
- Real content headings and paragraphs
- List items
- Article structure preserved despite errors

---

### 5. script-heavy/
**Purpose:** Test script and style removal

**Contains:**
- 10+ external script tags
- Multiple inline scripts
- Large inline CSS blocks
- External stylesheets
- Deferred and async scripts
- Analytics and tracking code
- Social media widgets

**What to verify:**
- All `<script>` tags removed (external and inline)
- All `<style>` tags removed
- External stylesheet links removed
- Actual content preserved

**Expected output:**
- Article heading and content
- Subsection headings
- Paragraphs and lists
- NO scripts or styles

---

### 6. navigation-heavy/
**Purpose:** Test navigation removal

**Contains:**
- Top navigation bar
- Breadcrumb navigation
- Secondary navigation
- Left sidebar navigation
- Right sidebar navigation
- Table of contents
- Related articles nav
- Pagination
- Footer navigation (4 sections)
- Mobile menu
- Social navigation

**What to verify:**
- All `<nav>` elements removed
- Sidebar navigation removed
- Footer navigation removed
- Article content preserved

**Expected output:**
- Article heading and content
- Section headings and paragraphs
- NO navigation elements

---

### 7. minimal-content/
**Purpose:** Test extraction when content is scarce

**Contains:**
- Only 1 paragraph of actual content
- 15+ ad elements
- Multiple sidebars with ads
- Cookie banners
- Popups
- Social widgets
- Tracking iframes
- Comments widget

**What to verify:**
- Single paragraph is correctly extracted
- All clutter is removed
- Output is minimal but correct
- Tool doesn't fail on mostly-empty output

**Expected output:**
- Article title (1 line)
- Single paragraph of content (1 line)
- Nothing else

---

### 8. modern-blog/
**Purpose:** Test realistic blog post scenario

**Contains:**
- Typical blog post structure
- Header with navigation
- Breadcrumbs
- Author card
- Newsletter widget
- Inline ads between paragraphs
- Social share buttons
- Comments section
- Related posts
- Footer
- Tracking pixels

**Simulates:** Real-world blog like Medium, Dev.to, or WordPress sites

**What to verify:**
- Article title and meta removed/preserved appropriately
- All 10 tips (headings + paragraphs) extracted
- Introduction and conclusion preserved
- All ads removed
- Navigation/footer/widgets removed

**Expected output:**
- Main heading
- 10 subsection headings (H2)
- All paragraph content
- Conclusion
- NO ads, widgets, navigation, or social buttons

---

### 9. nextjs-ssr/
**Purpose:** Test client-side JavaScript rendering (Next.js/React pattern)

**Contains:**
- Empty `<div id="__next"></div>` mount point
- All content rendered via JavaScript on DOMContentLoaded
- Article content stored in JavaScript data object
- Simulates client-side rendered React/Next.js app

**What to verify:**
- Tool returns minimal/empty content (no JS execution)
- Demonstrates limitation with SPAs
- No actual article content extracted
- Only skeleton HTML visible

**Expected output:**
- Empty or minimal HTML skeleton
- NO article content (requires JS execution)
- This test should FAIL to extract content (expected behavior)

---

### 10. vue-app/
**Purpose:** Test client-side JavaScript rendering (Vue.js pattern)

**Contains:**
- Empty `<div id="app"></div>` mount point
- All content rendered via JavaScript
- Vue-style component structure in JS
- Simulates client-side rendered Vue SPA

**What to verify:**
- Tool returns minimal/empty content
- Demonstrates SPA limitation
- No actual article content extracted

**Expected output:**
- Empty or minimal HTML skeleton
- NO article content (requires JS execution)
- This test should FAIL to extract content (expected behavior)

---

### 11. svelte-app/
**Purpose:** Test client-side JavaScript rendering (Svelte pattern)

**Contains:**
- Empty `<body>` tag
- All content rendered via JavaScript
- Svelte-compiled component pattern
- Simulates client-side rendered Svelte app

**What to verify:**
- Tool returns empty or minimal content
- Demonstrates SPA limitation
- No actual article content extracted

**Expected output:**
- Empty or minimal HTML
- NO article content (requires JS execution)
- This test should FAIL to extract content (expected behavior)

---

## Important Note on Framework Tests

**Tests 9-11 (nextjs-ssr, vue-app, svelte-app) are designed to FAIL.**

These tests demonstrate a critical limitation: webfetch-clean does not execute JavaScript and cannot extract content from client-side rendered Single Page Applications (SPAs). When testing these sites:

- Expected result: Empty or skeleton HTML only
- Actual article content: NOT extracted (requires browser JS execution)
- This is the correct behavior for a static HTML parser

**Real-world implications:**
- Many modern websites use SSR (Server-Side Rendering) which pre-renders HTML on the server
- These pre-rendered sites WILL work fine with webfetch-clean
- Pure client-side SPAs will return empty/skeleton content only

---

## Running Tests

### Starting the Test Server

The easiest way to test is using the included Go HTTP server:

```bash
# Build the server
cd test-sites
go build -o server server.go

# Run the server (default port 8080)
./server

# Or run with custom port
./server -port 3000

# Or run directly without building
go run server.go
```

The server will display all available test site URLs on startup:
```
Starting test site server...
Serving directory: /path/to/webfetch-clean/test-sites
Server running at: http://localhost:8080

Available test sites:
  - http://localhost:8080/heavy-ads/
  - http://localhost:8080/deep-nesting/
  - http://localhost:8080/large-content/
  - http://localhost:8080/malformed-html/
  - http://localhost:8080/script-heavy/
  - http://localhost:8080/navigation-heavy/
  - http://localhost:8080/minimal-content/
  - http://localhost:8080/modern-blog/
  - http://localhost:8080/nextjs-ssr/
  - http://localhost:8080/vue-app/
  - http://localhost:8080/svelte-app/
```

Then test with webfetch-clean:
```bash
webfetch-clean --cli --url http://localhost:8080/heavy-ads/ --format markdown
```

### CLI Mode

Test each site individually:

```bash
# Heavy ads test
webfetch-clean --cli --url file://$(pwd)/test-sites/heavy-ads/index.html --format markdown

# Deep nesting test
webfetch-clean --cli --url file://$(pwd)/test-sites/deep-nesting/index.html --format markdown

# Large content test
webfetch-clean --cli --url file://$(pwd)/test-sites/large-content/index.html --format markdown

# Malformed HTML test
webfetch-clean --cli --url file://$(pwd)/test-sites/malformed-html/index.html --format markdown

# Script-heavy test
webfetch-clean --cli --url file://$(pwd)/test-sites/script-heavy/index.html --format markdown

# Navigation-heavy test
webfetch-clean --cli --url file://$(pwd)/test-sites/navigation-heavy/index.html --format markdown

# Minimal content test
webfetch-clean --cli --url file://$(pwd)/test-sites/minimal-content/index.html --format markdown

# Modern blog test
webfetch-clean --cli --url file://$(pwd)/test-sites/modern-blog/index.html --format markdown
```

### Alternative: Python Server

If you prefer Python over the Go server:

```bash
# Start Python HTTP server
python3 -m http.server 8080

# Test with webfetch-clean
webfetch-clean --cli --url http://localhost:8080/heavy-ads/ --format markdown
```

### Batch Testing

Run all tests and save outputs (requires test server running):

```bash
#!/usr/bin/env bash

# Create output directory
mkdir -p test-results

# Test server URL (default)
SERVER="http://localhost:8080"

# Run all tests
for site in heavy-ads deep-nesting large-content malformed-html script-heavy navigation-heavy minimal-content modern-blog nextjs-ssr vue-app svelte-app; do
    echo "Testing $site..."
    webfetch-clean --cli \
        --url "$SERVER/$site/" \
        --format markdown \
        --output "test-results/$site-output.md"
    echo "✓ $site complete"
done

echo ""
echo "All tests complete! Results in test-results/"
```

Or test with file:// URLs (no server needed):

```bash
#!/usr/bin/env bash
mkdir -p test-results
for site in heavy-ads deep-nesting large-content malformed-html script-heavy navigation-heavy minimal-content modern-blog nextjs-ssr vue-app svelte-app; do
    echo "Testing $site..."
    webfetch-clean --cli \
        --url "file://$(pwd)/$site/index.html" \
        --format markdown \
        --output "test-results/$site-output.md"
done
```

## Success Criteria

A successful test should produce output that:

1. **Contains semantic content**
   - Headings (h1, h2, h3)
   - Paragraphs
   - Lists
   - Code blocks (if present)
   - Tables (if present)

2. **Removes clutter**
   - NO `<script>` tags
   - NO `<style>` tags
   - NO `<nav>` elements
   - NO ad-related divs
   - NO tracking iframes
   - NO social widgets
   - NO cookie banners
   - NO popups

3. **Maintains structure**
   - Heading hierarchy preserved
   - Paragraph order maintained
   - List structure intact

4. **Handles edge cases**
   - Doesn't crash on malformed HTML
   - Processes large files efficiently
   - Extracts content regardless of nesting depth

## Adding New Test Sites

To add a new test site:

1. Create a new directory: `mkdir test-sites/new-test/`
2. Create `index.html` with your test case
3. Document the test case in this README
4. Add to batch testing script
5. Verify expected output

## Notes

- These are stress tests designed to push the tool to its limits
- Real-world pages may be less extreme but will contain combinations of these patterns
- If any test fails, investigate the specific cleaning rule that needs adjustment
- File URLs may not work in all environments - use HTTP server for consistent results

---

**Last Updated:** 2026-01-14
