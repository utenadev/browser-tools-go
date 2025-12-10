# Browser Tools Go

This is a Go reimplementation of [utena/browser-tools](https://github.com/utenadev/browser-tools), which was forked from [badlogic/browser-tools](https://github.com/badlogic/browser-tools).

Go-native Chrome DevTools Protocol tools for agent-assisted web automation. These tools connect to a running Chrome instance with remote debugging enabled.

## Installation

```bash
go install github.com/user/browser-tools-go@latest
```
*(Note: Replace `github.com/user/browser-tools-go` with the actual repository path)*

This will install the `browser-tools-go` command in your `$GOPATH/bin` directory.

## Help

```bash
browser-tools-go --help          # Show all commands
browser-tools-go start --help    # Show start command options
browser-tools-go <command> --help # Show specific command help
```

## How to Invoke These Tools

**CRITICAL FOR AGENTS**: Use the `browser-tools-go` command with subcommands.

✓ CORRECT:
```bash
browser-tools-go start
browser-tools-go navigate https://example.com
browser-tools-go pick "#submit-button"
```

✗ INCORRECT:
```bash
./browser-tools-go start  # Not necessary if in PATH
```

## Session Management

The connection to the running browser is managed automatically.

- `start` launches Chrome and saves the connection info.
- `close` terminates the Chrome instance and cleans up the connection info.
- All other commands automatically use the saved connection info.

### Start Chrome

```bash
browser-tools-go start              # Fresh profile
browser-tools-go start --headless   # Run in headless mode
```

Launch Chrome with remote debugging enabled.

### Close Chrome

```bash
browser-tools-go close
```
Closes the Chrome instance that was started by `start`.

## Commands

### Navigate

```bash
browser-tools-go navigate https://google.com
```

Navigate the current tab to a new URL.

### Screenshot

```bash
browser-tools-go screenshot my-shot.png
browser-tools-go screenshot my-shot.png --url https://example.com
browser-tools-go screenshot --url https://example.com --full-page
```

Capture a screenshot. If path is omitted, saves to a temporary file.
- `--url <url>`: Navigate to a URL before taking the screenshot.
- `--full-page`: Capture the entire page.

### Pick Elements

```bash
browser-tools-go pick "#submit-button"
browser-tools-go pick ".item-class" --all
```

Picks and extracts information about elements matching a CSS selector.
- **`<selector>`**: The CSS selector to match.
- **`--all`**: Extract information from all matching elements instead of just the first one.

### Evaluate JavaScript

```bash
browser-tools-go eval 'document.title'
browser-tools-go eval 'document.querySelectorAll("a").length'
```

Execute a JavaScript expression in the active tab. The code is run in an async context. The result is returned as JSON.

### Cookies

```bash
browser-tools-go cookies
```

Display all cookies for the current browser context.

### Search Google

```bash
browser-tools-go search "rust programming"
browser-tools-go search "climate change" --n 10
browser-tools-go search "machine learning" --n 3 --content
```

Search Google and return results.
- `--n <num>`: Number of results to return (default: 5).
- `--content`: Fetch and extract readable content (as plain text) from each result.

### Extract Page Content

```bash
browser-tools-go content
browser-tools-go content https://example.com
browser-tools-go content --format text
```

Extracts readable content from a URL or the current page.
- `--format <format>`: Output format (`markdown`, `text`, or `html`, default: `markdown`).

### Hacker News Scraper

```bash
browser-tools-go hn-scraper
browser-tools-go hn-scraper 10
```

Scrapes the top stories from the Hacker News front page. Provide an optional limit for the number of stories (default: 30).
