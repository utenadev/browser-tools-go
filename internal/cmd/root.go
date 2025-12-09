package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"browser-tools-go/internal/browser"
	"github.com/spf13/cobra"
)

// Exit codes
const (
	ExitSuccess = 0
	ExitError   = 1
)

// NewRootCmd creates a new root command for the application.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "browser-tools-go",
		Short: "A Go implementation of browser-tools",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Configure logging to stderr
			log.SetOutput(os.Stderr)
		},
	}

	// Add browser lifecycle commands
	rootCmd.AddCommand(newStartCmd(), newCloseCmd(), newRunCmd())

	// Add action commands
	rootCmd.AddCommand(newNavigateCmd(), newScreenshotCmd(), newPickCmd(), newEvalCmd(), newCookiesCmd(), newSearchCmd(), newContentCmd(), newHnScraperCmd())

	return rootCmd
}


// Execute creates the root command and executes it.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(ExitError)
	}
}

// browserCtx manages the browser context for a command.
type browserCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
	err    error
}

// browserCtxKey is the key for storing the browserCtx in a context.Context.
type browserCtxKeyType string

const browserCtxKey browserCtxKeyType = "browserCtx"

// persistentPreRun ensures a browser context is available.
func persistentPreRun(cmd *cobra.Command, args []string) {
	if cmd.Context().Value(browserCtxKey) != nil {
		return
	}

	ctx, cancel, err := browser.NewPersistentContext()
	if err != nil {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations["error"] = fmt.Sprintf("Failed to connect to browser: %v. Is it running?", err)
		return
	}

	browserCtxVal := &browserCtx{ctx: ctx, cancel: cancel}
	ctxWithBrowser := context.WithValue(cmd.Context(), browserCtxKey, browserCtxVal)
	cmd.SetContext(ctxWithBrowser)
}

// persistentPostRun cancels the browser context.
func persistentPostRun(cmd *cobra.Command, args []string) {
	if cmd.Context().Value(browserCtxKey) != nil {
		val := cmd.Context().Value(browserCtxKey)
		if bc, ok := val.(*browserCtx); ok && bc.cancel != nil {
			// Don't cancel if the parent is the 'run' command, as it manages the lifecycle.
			if cmd.Parent() != nil && cmd.Parent().Use == "run <subcommand> [args...]" {
				return
			}
			bc.cancel()
		}
	}
}


// handleCmdErr checks for an error from the pre-run steps and handles it.
func handleCmdErr(cmd *cobra.Command) bool {
	if cmd.Annotations != nil {
		if errMsg, ok := cmd.Annotations["error"]; ok {
			log.Println(errMsg)
			return true
		}
	}
	return false
}

// getBrowserCtx retrieves the browser context safely.
func getBrowserCtx(cmd *cobra.Command) (*browserCtx, error) {
	val := cmd.Context().Value(browserCtxKey)
	if val == nil {
		return nil, fmt.Errorf("browser context not available")
	}
	bc, ok := val.(*browserCtx)
	if !ok {
		return nil, fmt.Errorf("invalid browser context type")
	}
	return bc, nil
}


// prettyPrintResults marshals the data to JSON and prints it to stdout.
func prettyPrintResults(data interface{}) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
}
