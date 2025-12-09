package cmd

import (
	"log"
	"strings"

	"browser-tools-go/internal/logic"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var n int
	var content bool

	cmd := &cobra.Command{
		Use:               "search <query>",
		Short:             "Search Google and return results",
		Args:              cobra.MinimumNArgs(1),
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,
		Run: func(cmd *cobra.Command, args []string) {
			if handleCmdErr(cmd) {
				return
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}

			query := strings.Join(args, " ")
			log.Printf("🔍 Searching Google for: %s (results: %d, content: %t)", query, n, content)

			results, err := logic.Search(bc.ctx, query, n, content)
			if err != nil {
				log.Fatalf("✗ Failed to perform search: %v", err)
			}
			prettyPrintResults(results)
		},
	}

	cmd.Flags().IntVar(&n, "n", 5, "Number of results to return")
	cmd.Flags().BoolVar(&content, "content", false, "Fetch content from each result")
	return cmd
}

func newContentCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:               "content [url]",
		Short:             "Extracts readable content from a URL or the current page",
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,
		Run: func(cmd *cobra.Command, args []string) {
			if handleCmdErr(cmd) {
				return
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}

			var url string
			if len(args) > 0 {
				url = args[0]
			}
			log.Printf("📄 Extracting content (format: %s)", format)

			result, err := logic.GetContent(bc.ctx, url, format)
			if err != nil {
				log.Fatalf("✗ Failed to extract content: %v", err)
			}
			prettyPrintResults(result)
		},
	}

	cmd.Flags().StringVar(&format, "format", "markdown", "Output format (markdown, text, or html)")
	return cmd
}

func newHnScraperCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:               "hn-scraper",
		Short:             "Scrapes the top stories from the Hacker News front page",
		Args:              cobra.NoArgs,
		PersistentPreRun:  persistentPreRun,
		PersistentPostRun: persistentPostRun,
		Run: func(cmd *cobra.Command, args []string) {
			if handleCmdErr(cmd) {
				return
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}

			log.Printf("📰 Scraping Hacker News (limit: %d)...", limit)

			submissions, err := logic.HnScraper(bc.ctx, limit)
			if err != nil {
				log.Fatalf("✗ Failed to scrape Hacker News: %v", err)
			}
			prettyPrintResults(submissions)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Number of stories to fetch")
	return cmd
}
