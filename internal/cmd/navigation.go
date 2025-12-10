package cmd

import (
	"log"

	"browser-tools-go/internal/logic"

	"github.com/spf13/cobra"
)

func newNavigateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "navigate <url>",
		Short:             "Navigate to a specific URL",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: persistentPreRunE,
		Run: func(cmd *cobra.Command, args []string) {
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}
			defer bc.cancel()

			log.Printf("🚀 Navigating to %s...", args[0])
			if err := logic.Navigate(bc.ctx, args[0]); err != nil {
				log.Fatalf("✗ Failed to navigate: %v", err)
			}
			log.Println("✅ Navigation successful.")
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
		PersistentPreRunE: persistentPreRunE,
		Run: func(cmd *cobra.Command, args []string) {
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
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
				log.Fatalf("✗ Failed to take screenshot: %v", err)
			}
			log.Printf("✅ Screenshot saved to: %s", savedPath)
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "URL to navigate to first")
	cmd.Flags().BoolVar(&fullPage, "full-page", false, "Take a full page screenshot")
	return cmd
}
