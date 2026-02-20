// Package main provides HTML cleaning functionality for the webfetch-clean tool.
package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CleanHTML cleans the HTML by removing ads, scripts, styles, navigation, and other clutter.
// Returns the cleaned HTML as a string.
func CleanHTML(html string, preserveMainOnly bool, removeImages bool, stripLinks bool) (string, error) {
	if html == "" {
		return "", fmt.Errorf("HTML content cannot be empty")
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// If preserveMainOnly is true, extract only main/article content
	if preserveMainOnly {
		main := doc.Find("main, article").First()
		if main.Length() > 0 {
			// Get the HTML of main/article before modifying the document
			mainHTML, err := main.Html()
			if err != nil {
				return "", fmt.Errorf("failed to extract main content: %w", err)
			}
			// Replace entire body with just the main/article content
			doc.Find("body").SetHtml(mainHTML)
		}
	}

	// Pass 1: Remove noise elements
	doc.Find("head, script, style, nav").Remove()

	// Pass 2: Remove ad-related elements (by class/id containing "ad", "advertisement", "banner")
	doc.Find("[class*='ad' i], [id*='ad' i]").Each(func(_ int, s *goquery.Selection) {
		// Check if it's actually ad-related (not "read", "header", "thread", etc.)
		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		combined := strings.ToLower(class + " " + id)

		// Match patterns that are likely ads
		adPatterns := []string{
			"advertisement",
			"adsbygoogle",
			"ad-",
			"-ad-",
			"-ad",
			"_ad_",
			"_ad",
			"ad_",
			" ad ",
		}

		for _, pattern := range adPatterns {
			if strings.Contains(combined, pattern) {
				s.Remove()
				return
			}
		}
	})

	// More specific ad removals
	doc.Find("[class*='advertisement' i], [id*='advertisement' i]").Remove()
	doc.Find("[class*='banner' i], [id*='banner' i]").Each(func(_ int, s *goquery.Selection) {
		// Keep elements with "banner" that might be headers, but remove actual ad banners
		class, _ := s.Attr("class")
		if !strings.Contains(strings.ToLower(class), "header") {
			s.Remove()
		}
	})

	// Pass 3: Remove tracking and iframes
	doc.Find("iframe").Remove()

	// Pass 4: Remove clutter elements
	doc.Find("footer, aside").Remove()

	// Sidebar removal: skip elements where "sidebar" only appears negated (e.g. "no-sidebar")
	doc.Find("[class*='sidebar' i], [id*='sidebar' i]").Each(func(_ int, s *goquery.Selection) {
		nodeName := goquery.NodeName(s)
		if nodeName == "body" || nodeName == "html" {
			return
		}
		class, _ := s.Attr("class")
		id, _ := s.Attr("id")
		combined := strings.ToLower(class + " " + id)
		// Strip negated references, then check if "sidebar" still appears
		negations := []string{"no-sidebar", "no_sidebar", "without-sidebar", "without_sidebar"}
		stripped := combined
		for _, neg := range negations {
			stripped = strings.ReplaceAll(stripped, neg, "")
		}
		if strings.Contains(stripped, "sidebar") {
			s.Remove()
		}
	})

	// Guard all clutter selectors against accidentally removing body/html
	doc.Find("[class*='menu' i]:not(main *, article *)").Not("body").Not("html").Remove()
	doc.Find("[class*='popup' i], [id*='popup' i]").Not("body").Not("html").Remove()
	doc.Find("[class*='modal' i], [id*='modal' i]").Not("body").Not("html").Remove()
	doc.Find("[class*='cookie' i], [id*='cookie' i]").Not("body").Not("html").Remove()
	doc.Find("[class*='social' i], [id*='social' i]").Not("body").Not("html").Remove()
	doc.Find("[class*='share' i], [id*='share' i]").Not("body").Not("html").Remove()
	doc.Find("[class*='comment' i], [id*='comment' i]").Not("body").Not("html").Remove()

	// Pass 5: Remove images if requested
	if removeImages {
		doc.Find("img").Remove()
	}

	// Pass 6: Strip links if requested (replace <a> with text content)
	if stripLinks {
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			s.ReplaceWithHtml(s.Text())
		})
	}

	// Pass 7: Strip inline attributes (keep only semantic ones)
	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		// Get all attributes
		nodeAttrs := s.Get(0).Attr
		attrs := make([]string, 0, len(nodeAttrs))
		for _, attr := range nodeAttrs {
			attrs = append(attrs, attr.Key)
		}

		// Remove all attributes except these semantic ones
		semanticAttrs := map[string]bool{
			"href":  true,
			"src":   true,
			"alt":   true,
			"title": true,
		}

		for _, attr := range attrs {
			if !semanticAttrs[attr] {
				s.RemoveAttr(attr)
			}
		}
	})

	// Get the cleaned HTML
	cleanedHTML, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to generate cleaned HTML: %w", err)
	}

	return cleanedHTML, nil
}
