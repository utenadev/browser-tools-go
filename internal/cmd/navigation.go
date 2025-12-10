package cmd

import (
	"fmt"
	"log"

	"browser-tools-go/internal/logic"

	"github.com/spf13/cobra"
)

func newNavigateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "navigate <url>",
		Short:             "Navigate to a specific URL",
		Args:              cobra.ExactArgs(1),
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			if handleCmdErr(cmd) {
				return fmt.Errorf("browser context error")
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				return fmt.Errorf("✗ %w", err)
			}
			defer bc.cancel()

			log.Printf("🚀 Navigating to %s...", args[0])
			if err := logic.Navigate(bc.ctx, args[0]); err != nil {
				return fmt.Errorf("✗ Failed to navigate: %w", err)
			}
			log.Println("✅ Navigation successful.")
			return nil
		},
	}
	return cmd
}

func newScreenshotCmd() *cobra.Command {
	var url string
	var fullPage bool

	cmd := &cobra.Command{
		Use:               "screenshot [path]",
		Short:             "Capture a screenshot of a web page",
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			if handleCmdErr(cmd) {
				return fmt.Errorf("browser context error")
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				return fmt.Errorf("✗ %w", err)
			}
			defer bc.cancel()

			filePath := ""
			if len(args) > 0 {
				filePath = args[0]
			}

			if url != "" {
				log.Printf("🚀 Navigating to %s...", url)
			}
			log.Println("📸 Taking screenshot...")

			savedPath, err := logic.Screenshot(bc.ctx, url, filePath, fullPage)
			if err != nil {
				return fmt.Errorf("✗ Failed to take screenshot: %w", err)
			}
			log.Printf("✅ Screenshot saved to: %s", savedPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "URL to navigate to first")
	cmd.Flags().BoolVar(&fullPage, "full-page", false, "Take a full page screenshot")
	return cmd
}
