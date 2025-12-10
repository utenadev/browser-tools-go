package cmd

import (
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
		Run: func(cmd *cobra.Command, args []string) {
			if handleCmdErr(cmd) {
				return
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}

			log.Printf("🔍 Picking elements with selector: %s (all=%t)...", args[0], all)

			results, err := logic.PickElements(bc.ctx, args[0], all)
			if err != nil {
				log.Fatalf("✗ Failed to pick elements: %v", err)
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
		Run: func(cmd *cobra.Command, args []string) {
			if handleCmdErr(cmd) {
				return
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}

			js := strings.Join(args, " ")
			log.Printf("📝 Evaluating JavaScript: %s", js)

			result, err := logic.EvaluateJS(bc.ctx, js)
			if err != nil {
				log.Fatalf("✗ Failed to evaluate JavaScript: %v", err)
			}
			prettyPrintResults(result)
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
		Run: func(cmd *cobra.Command, args []string) {
			if handleCmdErr(cmd) {
				return
			}
			bc, err := getBrowserCtx(cmd)
			if err != nil {
				log.Fatalf("✗ %v", err)
			}

			log.Println("🌐 Retrieving cookies...")

			cookies, err := logic.GetCookies(bc.ctx)
			if err != nil {
				log.Fatalf("✗ Failed to get cookies: %v", err)
			}
			prettyPrintResults(cookies)
		},
	}
	return cmd
}
