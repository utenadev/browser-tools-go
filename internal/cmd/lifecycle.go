package cmd

import (
	"fmt"

	"browser-tools-go/internal/browser"

	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var port int
	var headless bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a persistent Chrome instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := browser.Start(port, headless); err != nil {
				return fmt.Errorf("✗ Failed to start browser: %w", err)
			}
			return nil
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := browser.Close(); err != nil {
				return fmt.Errorf("✗ Failed to close browser: %w", err)
			}
			return nil
		},
	}
	return cmd
}
