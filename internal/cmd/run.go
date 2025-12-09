package cmd

import (
	"context"
	"log"

	"browser-tools-go/internal/browser"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var headless bool

	cmd := &cobra.Command{
		Use:   "run <subcommand> [args...]",
		Short: "Run a single command in a temporary browser instance",
		Long: `Run a subcommand with its own temporary browser that starts and stops automatically.
Example: browser-tools-go run screenshot --url https://example.com my.png`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			log.Println("🚀 Starting temporary browser...")
			ctx, cancel, err := browser.NewTemporaryContext(headless)
			if err != nil {
				log.Printf("✗ Failed to create temporary browser: %v", err)
				return err
			}

			browserCtxVal := &browserCtx{ctx: ctx, cancel: cancel}

			rootCmd := cmd.Root()
			ctxWithBrowser := context.WithValue(rootCmd.Context(), browserCtxKey, browserCtxVal)
			rootCmd.SetContext(ctxWithBrowser)

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			log.Println("✅ Temporary browser closed.")
			rootCmd := cmd.Root()
			if browserCtxVal := rootCmd.Context().Value(browserCtxKey); browserCtxVal != nil {
				if bc, ok := browserCtxVal.(*browserCtx); ok && bc.cancel != nil {
					bc.cancel()
				}
			}
		},
		// This is the key: allow cobra to parse flags for the subcommand.
		TraverseChildren: true,
		// Tell cobra that 'run' itself doesn't have a run function.
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}

	cmd.Flags().BoolVar(&headless, "headless", true, "Run the temporary browser in headless mode")
	// This allows subcommands to have their own flags without 'run' complaining.
	cmd.FParseErrWhitelist.UnknownFlags = true

	return cmd
}
