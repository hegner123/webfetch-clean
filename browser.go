package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// BrowserStableWait is the duration to wait for the DOM to stabilize
// after the initial page load, allowing SPA hydration to complete.
const BrowserStableWait = 2 * time.Second

// FetchBrowser fetches a URL using a headless Chromium browser via Rod.
// This handles JavaScript-rendered pages (SPAs) by executing the page's
// JavaScript before capturing the DOM. Returns the rendered HTML, the
// final URL after any client-side navigation, and any error.
func FetchBrowser(ctx context.Context, rawURL string, timeout int) (content string, finalURL string, err error) {
	if rawURL == "" {
		return "", "", fmt.Errorf("URL cannot be empty")
	}

	if err := ValidateURLScheme(rawURL); err != nil {
		return "", "", err
	}

	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	l := launcher.New().Headless(true)
	controlURL, err := l.Launch()
	if err != nil {
		return "", "", fmt.Errorf("failed to launch browser: %w", err)
	}
	defer l.Cleanup()

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return "", "", fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer func() {
		closeErr := browser.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close browser: %w", closeErr)
		}
	}()

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return "", "", fmt.Errorf("failed to create page: %w", err)
	}
	defer func() {
		closeErr := page.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close page: %w", closeErr)
		}
	}()

	page = page.Context(ctx)

	if err := page.Navigate(rawURL); err != nil {
		return "", "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return "", "", fmt.Errorf("failed waiting for page load: %w", err)
	}

	// Wait for DOM to stabilize after SPA hydration.
	// Non-fatal: some pages never fully stabilize, so we proceed regardless.
	_ = page.WaitDOMStable(BrowserStableWait, 0) // DOM may never fully stabilize—proceed with best-effort content

	info, err := page.Info()
	if err != nil {
		return "", "", fmt.Errorf("failed to get page info: %w", err)
	}
	finalURL = info.URL

	html, err := page.HTML()
	if err != nil {
		return "", "", fmt.Errorf("failed to get page HTML: %w", err)
	}

	if html == "" {
		return "", finalURL, fmt.Errorf("no content received from URL")
	}

	return html, finalURL, nil
}
