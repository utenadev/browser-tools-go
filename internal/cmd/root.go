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

var rootCmd = &cobra.Command{
	Use:   "browser-tools-go",
	Short: "A Go implementation of browser-tools",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configure logging to stderr
		log.SetOutput(os.Stderr)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitError)
	}
}

// browserCtx manages the browser context for a command.
type browserCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
	err    error
}

// persistentPreRun creates a persistent browser context for commands that need it.
func persistentPreRun(cmd *cobra.Command, args []string) {
	ctx, cancel, err := browser.NewPersistentContext()
	if err != nil {
		log.Printf("Error creating persistent context: %v", err)
		// Store the error in the command's annotations to be handled by the Run function.
		cmd.Annotations = map[string]string{"error": err.Error()}
		return
	}

	cmd.SetContext(context.WithValue(cmd.Context(), "browserCtx", &browserCtx{ctx: ctx, cancel: cancel}))
}

// persistentPostRun cancels the browser context.
func persistentPostRun(cmd *cobra.Command, args []string) {
	if browserCtxVal := cmd.Context().Value("browserCtx"); browserCtxVal != nil {
		if bc, ok := browserCtxVal.(*browserCtx); ok && bc.cancel != nil {
			bc.cancel()
		}
	}
}

// handleCmdErr checks for an error from the pre-run steps and handles it.
func handleCmdErr(cmd *cobra.Command) bool {
	if cmd.Annotations != nil {
		if errMsg, ok := cmd.Annotations["error"]; ok {
			log.Printf("Error: %s", errMsg)
			return true
		}
	}
	return false
}

// prettyPrintResults marshals the data to JSON and prints it to stdout.
func prettyPrintResults(data interface{}) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
}

func init() {
	// Add browser lifecycle commands
	rootCmd.AddCommand(newStartCmd(), newCloseCmd(), newRunCmd())

	// Add action commands
	rootCmd.AddCommand(newNavigateCmd(), newScreenshotCmd(), newPickCmd(), newEvalCmd(), newCookiesCmd(), newSearchCmd(), newContentCmd(), newHnScraperCmd())
}
