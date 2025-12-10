package browser

import (
	"context"
	"fmt"

	"browser-tools-go/internal/config"

	"github.com/chromedp/chromedp"
)

// NewPersistentContext creates a new browser context connected to a persistent, remote browser instance.
func NewPersistentContext() (context.Context, context.CancelFunc, error) {
	info, err := config.LoadWsInfo()
	if err != nil {
		return nil, nil, fmt.Errorf("could not load browser session, is it running? Error: %w", err)
	}

	allocCtx, cancel1 := chromedp.NewRemoteAllocator(context.Background(), info.Url)
	ctx, cancel2 := chromedp.NewContext(allocCtx)

	cancel := func() {
		cancel2()
		cancel1()
	}
	return ctx, cancel, nil
}

// NewTemporaryContext creates a new browser context with its own temporary browser instance.
func NewTemporaryContext(headless bool) (context.Context, context.CancelFunc, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
	)

	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel2 := chromedp.NewContext(allocCtx)

	cancel := func() {
		cancel2()
		cancel1()
	}
	return ctx, cancel, nil
}
