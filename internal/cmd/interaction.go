package cmd

import (
	"fmt"
	"log"
	"strings"

	"browser-tools-go/internal/logic"

	"github.com/spf13/cobra"
)

func newPickCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:               "pick <selector>",
		Short:             "Pick and extract information about elements matching a CSS selector",
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

			log.Printf("🔍 Picking elements with selector: %s (all=%t)...", args[0], all)

			results, err := logic.PickElements(bc.ctx, args[0], all)
			if err != nil {
				return fmt.Errorf("✗ Failed to pick elements: %w", err)
			}
			if len(results) == 0 {
				log.Println("✅ No elements found.")
				return nil
			}

			if all {
				prettyPrintResults(results)
			} else {
				prettyPrintResults(results[0])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Extract info from all matching elements")
	return cmd
}

func newEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "eval <javascript>",
		Short:             "Execute a JavaScript expression",
		Args:              cobra.MinimumNArgs(1),
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

			js := strings.Join(args, " ")
			log.Printf("📝 Evaluating JavaScript: %s", js)

			result, err := logic.EvaluateJS(bc.ctx, js)
			if err != nil {
				return fmt.Errorf("✗ Failed to evaluate JavaScript: %w", err)
			}
			prettyPrintResults(result)
			return nil
		},
	}
	return cmd
}

func newCookiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "cookies",
		Short:             "Display all cookies for the current browser context",
		Args:              cobra.NoArgs,
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

			log.Println("🌐 Retrieving cookies...")

			cookies, err := logic.GetCookies(bc.ctx)
			if err != nil {
				return fmt.Errorf("✗ Failed to get cookies: %w", err)
			}
			prettyPrintResults(cookies)
			return nil
		},
	}
	return cmd
}
