package cmd

import (
	"log"

	"browser-tools-go/internal/browser"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var port int
	var headless bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a persistent Chrome instance",
		Run: func(cmd *cobra.Command, args []string) {
			if err := browser.Start(port, headless); err != nil {
				log.Fatalf("✗ Failed to start browser: %v", err)
			}
		},
	}

	cmd.Flags().IntVar(&port, "port", 9222, "Port for debugging")
	cmd.Flags().BoolVar(&headless, "headless", false, "Run headless")
	return cmd
}

func newCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close the persistent Chrome instance",
		Run: func(cmd *cobra.Command, args []string) {
			if err := browser.Close(); err != nil {
				log.Fatalf("✗ Failed to close browser: %v", err)
			}
		},
	}
	return cmd
}
