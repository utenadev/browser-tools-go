package cmd

import (
	"log"
	"strconv"

	"browser-tools-go/internal/browser"
	"browser-tools-go/internal/logic"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var headless bool
	// Flags for subcommands need to be re-declared here.
	var url string
	var fullPage bool
	var all bool
	var n int
	var content bool
	var format string
	var limit int

	cmd := &cobra.Command{
		Use:   "run <subcommand> [args...]",
		Short: "Run a single command in a temporary browser instance",
		Long: `Run a subcommand with its own temporary browser that starts and stops automatically.
Example: browser-tools-go run screenshot my.png --url https://example.com`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("🚀 Starting temporary browser...")
			ctx, cancel, err := browser.NewTemporaryContext(headless)
			if err != nil {
				log.Fatalf("✗ Failed to create temporary browser: %v", err)
			}
			defer func() {
				cancel()
				log.Println("✅ Temporary browser closed.")
			}()

			subcommandName := args[0]
			subcommandArgs := args[1:]

			log.Printf("🚀 Running '%s' in temporary browser...", subcommandName)

			// This approach is not ideal, but it's a direct port of the original logic.
			// A better long-term solution would involve more sophisticated command composition.
			switch subcommandName {
			case "navigate":
				if len(subcommandArgs) != 1 {
					log.Fatal("navigate requires a URL")
				}
				if err := logic.Navigate(ctx, subcommandArgs[0]); err != nil {
					log.Fatalf("✗ Navigate command failed: %v", err)
				}
			case "screenshot":
				filePath := ""
				if len(subcommandArgs) > 0 {
					filePath = subcommandArgs[0]
				}
				savedPath, err := logic.Screenshot(ctx, url, filePath, fullPage)
				if err != nil {
					log.Fatalf("✗ Screenshot command failed: %v", err)
				}
				log.Printf("✅ Screenshot saved to: %s", savedPath)
			case "pick":
				if len(subcommandArgs) != 1 {
					log.Fatal("pick requires a selector")
				}
				results, err := logic.PickElements(ctx, subcommandArgs[0], all)
				if err != nil {
					log.Fatalf("✗ Pick command failed: %v", err)
				}
				if len(results) == 0 {
					log.Println("✅ No elements found.")
					return
				}
				if all {
					prettyPrintResults(results)
				} else {
					prettyPrintResults(results[0])
				}
			case "hn-scraper":
				if len(subcommandArgs) > 0 {
					limit, err = strconv.Atoi(subcommandArgs[0])
					if err != nil {
						log.Fatalf("✗ Invalid limit for hn-scraper: %v", err)
					}
				}
				submissions, err := logic.HnScraper(ctx, limit)
				if err != nil {
					log.Fatalf("✗ hn-scraper command failed: %v", err)
				}
				prettyPrintResults(submissions)
			// Add other cases here as needed
			default:
				log.Fatalf("Subcommand '%s' is not supported by 'run'.", subcommandName)
			}
		},
	}

	// Add flags for `run` and all the subcommands it can execute.
	cmd.Flags().BoolVar(&headless, "headless", false, "Run the temporary browser in headless mode")
	cmd.Flags().StringVar(&url, "url", "", "URL for subcommands like screenshot")
	cmd.Flags().BoolVar(&fullPage, "full-page", false, "Full page for screenshot subcommand")
	cmd.Flags().BoolVar(&all, "all", false, "All flag for pick subcommand")
	cmd.Flags().IntVar(&n, "n", 5, "Number of results for search subcommand")
	cmd.Flags().BoolVar(&content, "content", false, "Content flag for search subcommand")
	cmd.Flags().StringVar(&format, "format", "markdown", "Format flag for content subcommand")
	cmd.Flags().IntVar(&limit, "limit", 10, "Limit for hn-scraper subcommand")

	return cmd
}
