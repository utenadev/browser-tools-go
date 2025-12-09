package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"browser-tools-go/internal/config"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-shiori/go-readability"
	"github.com/spf13/cobra"
)

// --- Structs ---
type SearchResult struct {
	Title   string
	Link    string
	Snippet string
	Content string
}

type HnSubmission struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Points   int    `json:"points"`
	Author   string `json:"author"`
	Time     string `json:"time"`
	Comments int    `json:"comments"`
	HnURL    string `json:"hnUrl"`
}

// --- Main execution ---
func main() {
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// --- Browser Context Helpers ---
func newPersistentContext() (context.Context, context.CancelFunc) {
	info, err := config.LoadWsInfo()
	if err != nil {
		log.Fatalf("✗ Could not load browser session. Is it running? Error: %v", err)
	}
	allocCtx, cancel1 := chromedp.NewRemoteAllocator(context.Background(), info.Url)
	ctx, cancel2 := chromedp.NewContext(allocCtx)
	return ctx, func() {
		cancel2()
		cancel1()
	}
}

func newTemporaryContext(headless bool) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
	)
	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel2 := chromedp.NewContext(allocCtx)
	return ctx, func() {
		cancel2()
		cancel1()
	}
}

// --- Command Logic Functions ---
// These functions contain the core logic and are called by the cobra commands.

func logicNavigate(ctx context.Context, url string) {
	log.Printf("🚀 Navigating to %s...", url)
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		log.Fatalf("✗ Failed to navigate: %v", err)
	}
	log.Println("✅ Navigation successful.")
}

func logicScreenshot(ctx context.Context, targetURL, filePath string, fullPage bool) {
	tasks := make(chromedp.Tasks, 0)
	if targetURL != "" {
		log.Printf("🚀 Navigating to %s...", targetURL)
		tasks = append(tasks, chromedp.Navigate(targetURL))
	}

	log.Println("📸 Taking screenshot...")
	var buf []byte
	if fullPage {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, err = page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatPng).WithCaptureBeyondViewport(true).Do(ctx)
			return err
		}))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		log.Fatalf("✗ Failed to take screenshot: %v", err)
	}
	
	if filePath == "" {
		tmpFile, err := os.CreateTemp("", "screenshot-*.png")
		if err != nil {
			log.Fatalf("✗ Failed to create temporary file: %v", err)
		}
		filePath = tmpFile.Name()
		tmpFile.Close()
	}

	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		log.Fatalf("✗ Failed to save screenshot: %v", err)
	}
	log.Printf("✅ Screenshot saved to: %s", filePath)
}

// ... other logic functions ...

// --- Cobra Command Definitions ---
var rootCmd = &cobra.Command{
	Use:   "browser-tools-go",
	Short: "A Go implementation of browser-tools",
}

func needsBrowser(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(mustGetConfigPath()); os.IsNotExist(err) {
		return fmt.Errorf("✗ Chrome is not running. Please start it first with 'browser-tools-go start' or use the 'run' command")
	}
	return nil
}

func mustGetConfigPath() string {
	path, err := config.GetConfigPath()
	if err != nil {
		log.Fatalf("Could not determine config path: %v", err)
	}
	return path
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a persistent Chrome instance",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := config.LoadWsInfo(); err == nil {
			log.Fatal("✗ Browser is already running. Use 'close' to stop it first.")
		}
		port, _ := cmd.Flags().GetInt("port")
		headless, _ := cmd.Flags().GetBool("headless")
		chromePath, ok := chromedp.LookPath()
		if !ok {
			log.Fatal("✗ Could not find Chrome installation.")
		}
		userDataDir := strings.Replace(mustGetConfigPath(), "ws.json", "user-data", 1)
		chromeArgs := []string{
			fmt.Sprintf("--remote-debugging-port=%d", port),
			fmt.Sprintf("--user-data-dir=%s", userDataDir),
		}
		if headless {
			chromeArgs = append(chromeArgs, "--headless=new")
		}
		proc := exec.Command(chromePath, chromeArgs...)
		if runtime.GOOS == "windows" {
			proc.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
		}
		err := proc.Start()
		if err != nil {
			log.Fatalf("✗ Failed to start Chrome: %v", err)
		}
		wsURL := fmt.Sprintf("ws://127.0.0.1:%d", port)
		log.Printf("⏳ Waiting for browser to be ready at %s...", wsURL)
		time.Sleep(2 * time.Second)
		err = config.SaveWsInfo(wsURL, proc.Process.Pid)
		if err != nil {
			log.Fatalf("✗ Failed to save session info: %v", err)
		}
		log.Printf("✅ Browser started successfully with PID %d.", proc.Process.Pid)
	},
}

var closeCmd = &cobra.Command{
	Use:   "close",
	Short: "Close the persistent Chrome instance",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := config.LoadWsInfo()
		if err != nil {
			log.Fatal("✗ Browser is not running.")
		}
		log.Printf("🛑 Closing browser with PID %d...", info.Pid)
		proc, err := os.FindProcess(info.Pid)
		if err == nil {
			if runtime.GOOS == "windows" {
				err = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(info.Pid)).Run()
			} else {
				err = proc.Signal(syscall.SIGTERM)
			}
		}
		if err := config.RemoveWsInfo(); err != nil {
			log.Fatalf("✗ Failed to remove session file: %v", err)
		}
		log.Println("✅ Browser session closed and cleaned up.")
	},
}

var navigateCmd = &cobra.Command{
	Use:               "navigate <url>",
	Short:             "Navigate to a specific URL",
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicNavigate(ctx, args[0])
	},
}

var screenshotCmd = &cobra.Command{
	Use:               "screenshot [path]",
	Short:             "Capture a screenshot of a web page",
	Args:              cobra.MaximumNArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		fullPage, _ := cmd.Flags().GetBool("full-page")
		filePath := ""
		if len(args) > 0 {
			filePath = args[0]
		}
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicScreenshot(ctx, url, filePath, fullPage)
	},
}

var runCmd = &cobra.Command{
    Use:   "run <subcommand> [args...]",
    Short: "Run a single command in a temporary browser instance",
    Long: `Run a subcommand with its own temporary browser that starts and stops automatically.
Example: browser-tools-go run screenshot my.png --url https://example.com`,
    Args: cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        headless, _ := cmd.Flags().GetBool("headless")
        
        ctx, cancel := newTemporaryContext(headless)
        defer cancel()

        subcommandName := args[0]
        subcommandArgs := args[1:]

        // Find the command to run
        foundCmd, _, err := rootCmd.Find(args)
        if err != nil {
            log.Fatalf("Error finding subcommand for 'run': %v", err)
        }
        
        // We need to re-create the command to avoid messing with the main command definitions
        // This is a simplified approach: we manually call the logic function.
        // A more robust solution would parse flags for the subcommand.
        log.Printf("🚀 Running '%s' in temporary browser...", subcommandName)

        switch subcommandName {
        case "navigate":
            if len(subcommandArgs) != 1 { log.Fatal("navigate requires a URL") }
            logicNavigate(ctx, subcommandArgs[0])
        case "screenshot":
             // This part is tricky because flags are parsed by the parent 'run' command.
             // For a robust solution, we'd need to parse flags specifically for the subcommand.
             // As a simplification, we'll assume flags are passed to `run` itself.
            url, _ := cmd.Flags().GetString("url")
            fullPage, _ := cmd.Flags().GetBool("full-page")
            filePath := ""
            // Positional args are harder to handle here. We look for non-flag args.
            nonFlagArgs := []string{}
            for _, arg := range subcommandArgs {
                if !strings.HasPrefix(arg, "-") {
                    nonFlagArgs = append(nonFlagArgs, arg)
                }
            }
            if len(nonFlagArgs) > 0 {
                filePath = nonFlagArgs[0]
            }
            logicScreenshot(ctx, url, filePath, fullPage)
        default:
            log.Fatalf("Subcommand '%s' is not supported by 'run' in this simplified implementation.", subcommandName)
        }
    },
}


// ... (rest of the command definitions would be here, but keeping it brief)
// We need to make sure the flags for subcommands are also available on the `run` command.

func init() {
	rootCmd.AddCommand(startCmd, closeCmd, navigateCmd, screenshotCmd, runCmd) // Add other cmds later

	startCmd.Flags().Int("port", 9222, "Port for debugging")
	startCmd.Flags().Bool("headless", false, "Run headless")

	screenshotCmd.Flags().String("url", "", "URL to navigate to first")
	screenshotCmd.Flags().Bool("full-page", false, "Take a full page screenshot")
	
	// Flags for `run` must include flags for all subcommands it might run.
	runCmd.Flags().Bool("headless", false, "Run the temporary browser in headless mode")
	runCmd.Flags().String("url", "", "URL for subcommands like screenshot")
	runCmd.Flags().Bool("full-page", false, "Full page for screenshot subcommand")
}