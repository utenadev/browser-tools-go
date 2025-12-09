package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

type ElementInfo struct {
	Tag      string                 `json:"tag"`
	Text     string                 `json:"text"`
	Attrs    map[string]string      `json:"attrs"`
	Rect     map[string]interface{} `json:"rect"`
	Children []ElementInfo          `json:"children"`
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

func logicPick(ctx context.Context, selector string, all bool) {
	log.Printf("🔍 Picking elements with selector: %s (all=%t)...", selector, all)

	var nodes []*cdp.Node
	var res []ElementInfo

	// Get nodes by selector
	err := chromedp.Run(ctx,
		chromedp.Nodes(selector, &nodes, chromedp.NodeVisible, chromedp.ByQuery),
	)
	if err != nil || len(nodes) == 0 {
		log.Printf("⚠️  No elements found for selector: %s (Error: %v)", selector, err)
		return
	}

	// If not 'all', only process the first element
	if !all {
		nodes = nodes[:1]
	}

	for _, node := range nodes {
		// Get attributes
		attrs := make(map[string]string)
		if node.Attributes != nil {
			for i := 0; i < len(node.Attributes); i += 2 {
				if i+1 < len(node.Attributes) {
					attrs[node.Attributes[i]] = node.Attributes[i+1]
				}
			}
		}

		// Get text content
		var text string
		err = chromedp.Run(ctx, chromedp.Text(selector, &text, chromedp.NodeVisible, chromedp.ByQuery))
		if err != nil {
			text = ""
			// Try to get text from the specific node using evaluate
			err = chromedp.Run(ctx,
				chromedp.EvaluateAsDevTools(
					fmt.Sprintf(`
						var element = document.querySelector('%s');
						element ? element.textContent || element.innerText : "";
					`, strings.ReplaceAll(selector, "'", "\\'")),
					&text,
				),
			)
			if err != nil {
				text = ""
			}
		}

		// Get bounding rect
		var rect map[string]interface{}
		err = chromedp.Run(ctx,
			chromedp.Evaluate(
				fmt.Sprintf(`
					var element = document.querySelector('%s');
					if (element) {
						var rect = element.getBoundingClientRect();
						JSON.stringify({
							x: rect.x,
							y: rect.y,
							width: rect.width,
							height: rect.height
						});
					} else {
						JSON.stringify({});
					}
				`, strings.ReplaceAll(selector, "'", "\\'")),
				&rect,
			),
		)
		if err != nil {
			rect = make(map[string]interface{})
		}

		// Recursively get children (simplified approach)
		children := []ElementInfo{}

		res = append(res, ElementInfo{
			Tag:      node.LocalName,
			Text:     strings.TrimSpace(text),
			Attrs:    attrs,
			Rect:     rect,
			Children: children,
		})
	}

	if all {
		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			log.Fatalf("✗ Failed to marshal result: %v", err)
		}
		fmt.Println(string(output))
	} else if len(res) > 0 {
		output, err := json.MarshalIndent(res[0], "", "  ")
		if err != nil {
			log.Fatalf("✗ Failed to marshal result: %v", err)
		}
		fmt.Println(string(output))
	}
}

func logicEval(ctx context.Context, jsExpression string) {
	log.Printf("📝 Evaluating JavaScript: %s", jsExpression)

	var result interface{}
	err := chromedp.Run(ctx, chromedp.Evaluate(jsExpression, &result))
	if err != nil {
		log.Fatalf("✗ Failed to evaluate JavaScript: %v", err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("✗ Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
}

func logicCookies(ctx context.Context) {
	log.Println("🌐 Retrieving cookies...")

	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		log.Fatalf("✗ Failed to get cookies: %v", err)
	}

	// Create a slice to store cookies in a serializable format
	cookieList := make([]map[string]interface{}, len(cookies))
	for i, cookie := range cookies {
		cookieList[i] = map[string]interface{}{
			"name":     cookie.Name,
			"value":    cookie.Value,
			"domain":   cookie.Domain,
			"path":     cookie.Path,
			"expires":  cookie.Expires,
			"size":     cookie.Size,
			"httpOnly": cookie.HTTPOnly,
			"secure":   cookie.Secure,
			"session":  cookie.Session,
		}
	}

	output, err := json.MarshalIndent(cookieList, "", "  ")
	if err != nil {
		log.Fatalf("✗ Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
}

func logicSearch(ctx context.Context, query string, numResults int, fetchContent bool) {
	log.Printf("🔍 Searching Google for: %s (results: %d, fetchContent: %t)", query, numResults, fetchContent)

	// First, navigate to Google search page
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))
	log.Printf("🚀 Navigating to %s...", searchURL)

	if err := chromedp.Run(ctx, chromedp.Navigate(searchURL)); err != nil {
		log.Fatalf("✗ Failed to navigate to Google: %v", err)
	}

	// Wait for results to load
	time.Sleep(2 * time.Second)

	// Extract search results
	var titles []string
	var links []string
	var snippets []string

	err := chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('h3')).map(el => el.innerText)`, &titles),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('div#search a')).slice(0,20).filter(a => a.href.includes('http') && !a.href.includes('google.com')).map(a => a.href)`, &links),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('div.VwiC3b')).map(el => el.innerText)`, &snippets),
	)
	if err != nil {
		log.Printf("⚠️  Error extracting results: %v", err)
		// Try alternative selectors
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a h3')).map(el => el.innerText)`, &titles),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).slice(0,20).filter(a => a.href.includes('http') && !a.href.includes('google.com')).map(a => a.href)`, &links),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('span')).map(el => el.innerText)`, &snippets),
		)
		if err != nil {
			log.Fatalf("✗ Failed to extract search results: %v", err)
		}
	}

	// Limit results to requested number
	if numResults > 0 && numResults < len(titles) {
		titles = titles[:numResults]
		links = links[:numResults]
		if len(snippets) > numResults {
			snippets = snippets[:numResults]
		}
	}

	// Make sure all slices have the same length
	// Sometimes the selectors return different number of elements
	minLen := len(titles)
	if len(links) < minLen {
		minLen = len(links)
	}
	if len(snippets) < minLen {
		minLen = len(snippets)
	}
	titles = titles[:minLen]
	links = links[:minLen]
	snippets = snippets[:minLen]

	// Create result objects
	results := make([]SearchResult, minLen)
	for i := 0; i < minLen; i++ {
		results[i] = SearchResult{
			Title:   strings.TrimSpace(titles[i]),
			Link:    links[i],
			Snippet: strings.TrimSpace(snippets[i]),
			Content: "", // Will be filled if fetchContent is true
		}
	}

	// If requested, fetch content from each result
	if fetchContent {
		for i := range results {
			log.Printf("📄 Fetching content from: %s", results[i].Link)

			// Navigate to the page
			if err := chromedp.Run(ctx, chromedp.Navigate(results[i].Link)); err != nil {
				log.Printf("⚠️  Failed to navigate to %s: %v", results[i].Link, err)
				continue
			}

			// Wait for page to load
			time.Sleep(2 * time.Second)

			// Extract readable content
			var content string
			err := chromedp.Run(ctx,
				chromedp.Evaluate(
					"document.querySelector('body').innerText || document.body.textContent",
					&content,
				),
			)
			if err != nil {
				log.Printf("⚠️  Failed to extract content from %s: %v", results[i].Link, err)
				continue
			}

			// Limit content length for readability
			if len(content) > 2000 {
				content = content[:2000] + "..."
			}

			results[i].Content = content
		}

		// Navigate back to search results after fetching content
		if err := chromedp.Run(ctx, chromedp.Navigate(searchURL)); err != nil {
			log.Printf("⚠️  Failed to navigate back to search results: %v", err)
		}
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("✗ Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
}

func logicContent(ctx context.Context, targetURL string, format string) {
	log.Printf("📄 Extracting content (format: %s)", format)

	// Navigate if URL is provided
	if targetURL != "" {
		log.Printf("🚀 Navigating to %s...", targetURL)
		if err := chromedp.Run(ctx, chromedp.Navigate(targetURL)); err != nil {
			log.Fatalf("✗ Failed to navigate: %v", err)
		}
		time.Sleep(2 * time.Second) // Wait for page to load
	}

	// Extract the page content
	var content string
	err := chromedp.Run(ctx,
		chromedp.Evaluate("document.querySelector('body').innerHTML", &content),
	)
	if err != nil {
		log.Fatalf("✗ Failed to extract page content: %v", err)
	}

	// Get the page title
	var title string
	err = chromedp.Run(ctx,
		chromedp.Evaluate("document.title", &title),
	)
	if err != nil {
		log.Printf("⚠️  Could not get page title: %v", err)
		title = "Untitled"
	}

	// Process content according to format
	switch format {
	case "text":
		// Convert HTML to plain text
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			log.Fatalf("✗ Failed to parse HTML: %v", err)
		}
		content = strings.TrimSpace(doc.Text())
	case "markdown":
		// Convert HTML to markdown
		converter := md.NewConverter("", true, nil)
		content, err = converter.ConvertString(content)
		if err != nil {
			log.Fatalf("✗ Failed to convert HTML to markdown: %v", err)
		}
	case "html":
		// Keep raw HTML (do nothing)
	default:
		log.Fatalf("✗ Unsupported format: %s. Use 'text', 'markdown', or 'html'", format)
	}

	// Output the result
	result := map[string]interface{}{
		"title":   title,
		"content": content,
		"format":  format,
		"url":     targetURL,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("✗ Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
}

func logicHnScraper(ctx context.Context, limit int) {
	log.Printf("📰 Scraping Hacker News (limit: %d)...", limit)

	hnURL := "https://news.ycombinator.com"
	log.Printf("🚀 Navigating to %s...", hnURL)

	if err := chromedp.Run(ctx, chromedp.Navigate(hnURL)); err != nil {
		log.Fatalf("✗ Failed to navigate to Hacker News: %v", err)
	}

	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Extract the story information
	var titles []string
	var urlTexts []string  // This will contain the URLs or empty strings if not available
	var scoreTexts []string
	var authorTexts []string
	var timeTexts []string
	var commentTexts []string
	var moreLinks []string // For additional links (if needed)

	// Extract the required elements
	err := chromedp.Run(ctx,
		// Get story titles
		chromedp.Evaluate(`Array.from(document.querySelectorAll('span.titleline > a')).map(a => a.textContent)`, &titles),
		// Get links (for stories that have them)
		chromedp.Evaluate(`Array.from(document.querySelectorAll('span.titleline > a')).map(a => a.href)`, &urlTexts),
		// Get scores
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.score')).map(el => el.textContent)`, &scoreTexts),
		// Get authors
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.hnuser')).map(el => el.textContent)`, &authorTexts),
		// Get time info
		chromedp.Evaluate(`Array.from(document.querySelectorAll('span.age a')).map(el => el.title || el.textContent)`, &timeTexts),
		// Get comment counts
		chromedp.Evaluate(`Array.from(document.querySelectorAll('td.subtext > a')).filter(a => a.textContent.includes('comment')).map(a => a.textContent)`, &commentTexts),
		// Get "More" links if needed
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).filter(a => a.textContent.trim() === 'more').map(a => a.href)`, &moreLinks),
	)
	if err != nil {
		// If the first attempt fails, try alternative selectors
		log.Printf("⚠️  First attempt failed, trying alternative selectors: %v", err)
		err = chromedp.Run(ctx,
			// Get story titles
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a.storylink')).map(a => a.textContent)`, &titles),
			// Get links
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a.storylink')).map(a => a.href)`, &urlTexts),
			// Get scores
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.score')).map(el => el.textContent)`, &scoreTexts),
			// Get authors
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.hnuser')).map(el => el.textContent)`, &authorTexts),
			// Get time info
			chromedp.Evaluate(`Array.from(document.querySelectorAll('span.age a')).map(el => el.title || el.textContent)`, &timeTexts),
			// Get comment counts
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).filter(a => a.textContent.includes('comment')).map(a => a.textContent)`, &commentTexts),
		)
		if err != nil {
			log.Fatalf("✗ Failed to extract data from Hacker News: %v", err)
		}
	}

	// Calculate the minimum length of all slices to avoid index out of bounds
	minLen := len(titles)
	if len(urlTexts) < minLen {
		minLen = len(urlTexts)
	}
	if len(scoreTexts) < minLen {
		minLen = len(scoreTexts)
	}
	if len(authorTexts) < minLen {
		minLen = len(authorTexts)
	}
	if len(timeTexts) < minLen {
		minLen = len(timeTexts)
	}
	if len(commentTexts) < minLen {
		minLen = len(commentTexts)
	}

	// Limit to the specified number if applicable
	if limit > 0 && limit < minLen {
		minLen = limit
	}

	// Create submission objects
	submissions := make([]HnSubmission, minLen)
	for i := 0; i < minLen; i++ {
		// Extract score (remove " points" from string)
		score := 0
		scoreStr := regexp.MustCompile(`\d+`).FindString(scoreTexts[i])
		if scoreStr != "" {
			score, _ = strconv.Atoi(scoreStr)
		}

		// Extract comments count (remove " comments" or " comment" from string)
		comments := 0
		commentsStr := regexp.MustCompile(`\d+`).FindString(commentTexts[i])
		if commentsStr != "" {
			comments, _ = strconv.Atoi(commentsStr)
		}

		// Create the submission object
		submissions[i] = HnSubmission{
			ID:       fmt.Sprintf("%d", i+1), // Assign a simple ID based on index
			Title:    strings.TrimSpace(titles[i]),
			URL:      urlTexts[i],
			Points:   score,
			Author:   "",
			Time:     timeTexts[i],
			Comments: comments,
			HnURL:    fmt.Sprintf("%s/item?id=%d", hnURL, i+1), // Placeholder HN URL
		}

		// If we have author info, use it
		if i < len(authorTexts) {
			submissions[i].Author = strings.TrimSpace(authorTexts[i])
		}
	}

	// Output the results
	output, err := json.MarshalIndent(submissions, "", "  ")
	if err != nil {
		log.Fatalf("✗ Failed to marshal result: %v", err)
	}
	fmt.Println(string(output))
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
		var chromePath string
		var ok bool
		// Try common chrome paths
		possiblePaths := []string{
			"chrome",
			"google-chrome",
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
		}

		// Expand %USERNAME% to actual user name
		homeDir, _ := os.UserHomeDir()
		userChromePath := filepath.Join(homeDir, "AppData", "Local", "Google", "Chrome", "Application", "chrome.exe")
		possiblePaths = append(possiblePaths, userChromePath)

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				chromePath = path
				ok = true
				break
			}
		}

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

var pickCmd = &cobra.Command{
	Use:               "pick <selector>",
	Short:             "Pick and extract information about elements matching a CSS selector",
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicPick(ctx, args[0], all)
	},
}

var evalCmd = &cobra.Command{
	Use:               "eval <javascript>",
	Short:             "Execute a JavaScript expression in the active tab",
	Args:              cobra.ExactArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicEval(ctx, args[0])
	},
}

var cookiesCmd = &cobra.Command{
	Use:               "cookies",
	Short:             "Display all cookies for the current browser context",
	Args:              cobra.NoArgs,
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicCookies(ctx)
	},
}

var searchCmd = &cobra.Command{
	Use:               "search <query>",
	Short:             "Search Google and return results",
	Args:              cobra.MinimumNArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		n, _ := cmd.Flags().GetInt("n")
		content, _ := cmd.Flags().GetBool("content")
		query := strings.Join(args, " ")
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicSearch(ctx, query, n, content)
	},
}

var contentCmd = &cobra.Command{
	Use:               "content [url]",
	Short:             "Extracts readable content from a URL or the current page",
	Args:              cobra.MaximumNArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		var url string
		if len(args) > 0 {
			url = args[0]
		}
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicContent(ctx, url, format)
	},
}

var hnScraperCmd = &cobra.Command{
	Use:               "hn-scraper [limit]",
	Short:             "Scrapes the top stories from the Hacker News front page",
	Args:              cobra.MaximumNArgs(1),
	PersistentPreRunE: needsBrowser,
	Run: func(cmd *cobra.Command, args []string) {
		var limit int
		if len(args) > 0 {
			var err error
			limit, err = strconv.Atoi(args[0])
			if err != nil {
				log.Fatalf("✗ Invalid limit: %v", err)
			}
		}
		ctx, cancel := newPersistentContext()
		defer cancel()
		logicHnScraper(ctx, limit)
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
        _, _, err := rootCmd.Find(args)
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
        case "pick":
            all, _ := cmd.Flags().GetBool("all")
            if len(subcommandArgs) != 1 { log.Fatal("pick requires a selector") }
            logicPick(ctx, subcommandArgs[0], all)
        case "hn-scraper":
            var limit int
            if len(subcommandArgs) > 0 {
                var err error
                limit, err = strconv.Atoi(subcommandArgs[0])
                if err != nil {
                    log.Fatalf("✗ Invalid limit: %v", err)
                }
            }
            logicHnScraper(ctx, limit)
        default:
            log.Fatalf("Subcommand '%s' is not supported by 'run' in this simplified implementation.", subcommandName)
        }
    },
}


// ... (rest of the command definitions would be here, but keeping it brief)
// We need to make sure the flags for subcommands are also available on the `run` command.

func init() {
	rootCmd.AddCommand(startCmd, closeCmd, navigateCmd, screenshotCmd, pickCmd, evalCmd, cookiesCmd, searchCmd, contentCmd, hnScraperCmd, runCmd) // Add other cmds later

	startCmd.Flags().Int("port", 9222, "Port for debugging")
	startCmd.Flags().Bool("headless", false, "Run headless")

	screenshotCmd.Flags().String("url", "", "URL to navigate to first")
	screenshotCmd.Flags().Bool("full-page", false, "Take a full page screenshot")

	pickCmd.Flags().Bool("all", false, "Extract information from all matching elements instead of just the first one")

	searchCmd.Flags().Int("n", 5, "Number of results to return")
	searchCmd.Flags().Bool("content", false, "Fetch and extract readable content from each result")

	contentCmd.Flags().String("format", "markdown", "Output format (markdown, text, or html)")

	// Flags for `run` must include flags for all subcommands it might run.
	runCmd.Flags().Bool("headless", false, "Run the temporary browser in headless mode")
	runCmd.Flags().String("url", "", "URL for subcommands like screenshot")
	runCmd.Flags().Bool("full-page", false, "Full page for screenshot subcommand")

	// Add pick flags to run command as well
	runCmd.Flags().Bool("all", false, "All flag for pick subcommand")

	// Add search flags to run command as well
	runCmd.Flags().Int("n", 5, "Number of results for search subcommand")
	runCmd.Flags().Bool("content", false, "Content flag for search subcommand")

	// Add content flags to run command as well
	runCmd.Flags().String("format", "markdown", "Format flag for content subcommand")
}