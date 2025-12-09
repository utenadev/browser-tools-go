package logic

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Navigate navigates the browser to a specific URL.
func Navigate(ctx context.Context, url string) error {
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	return nil
}

// Screenshot captures a screenshot of the current page.
func Screenshot(ctx context.Context, targetURL, filePath string, fullPage bool) (string, error) {
	tasks := make(chromedp.Tasks, 0)
	if targetURL != "" {
		tasks = append(tasks, chromedp.Navigate(targetURL))
	}

	var buf []byte
	if fullPage {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, err = page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatPng).WithCaptureBeyondViewport(true).Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to capture full page screenshot: %w", err)
			}
			return nil
		}))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		return "", fmt.Errorf("failed to take screenshot: %w", err)
	}

	if filePath == "" {
		tmpFile, err := os.CreateTemp("", "screenshot-*.png")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %w", err)
		}
		filePath = tmpFile.Name()
		tmpFile.Close()
	}

	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	return filePath, nil
}
