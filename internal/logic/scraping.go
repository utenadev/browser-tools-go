package logic

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"browser-tools-go/internal/config"
	"browser-tools-go/internal/models"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// Search performs a Google search and returns the results.
func Search(ctx context.Context, query string, numResults int, fetchContent bool) ([]models.SearchResult, error) {
	sel := config.GetSelectors().Google
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))

	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(sel.ResultContainer),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to google and wait for results: %w", err)
	}

	var titles, links, snippets []string
	titleJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(el => el.innerText)`, sel.Title)
	linkJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).filter(a => a.href.startsWith('http') && !a.href.includes('google.com')).map(a => a.href)`, sel.Link)
	snippetJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(el => el.innerText)`, sel.Snippet)

	err = chromedp.Run(ctx,
		chromedp.Evaluate(titleJS, &titles),
		chromedp.Evaluate(linkJS, &links),
		chromedp.Evaluate(snippetJS, &snippets),
	)
	if err != nil {
		// Fallback selectors
		fallbackTitleJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(el => el.innerText)`, sel.TitleFallback)
		err = chromedp.Run(ctx,
			chromedp.Evaluate(fallbackTitleJS, &titles),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to extract search result titles: %w", err)
		}
	}

	minLen := len(titles)
	if len(links) < minLen {
		minLen = len(links)
	}
	if len(snippets) < minLen {
		minLen = len(snippets)
	}
	if numResults > 0 && numResults < minLen {
		minLen = numResults
	}

	results := make([]models.SearchResult, minLen)
	for i := 0; i < minLen; i++ {
		results[i] = models.SearchResult{
			Title:   strings.TrimSpace(titles[i]),
			Link:    links[i],
			Snippet: strings.TrimSpace(snippets[i]),
		}
	}

	if fetchContent {
		for i := range results {
			var content string
			err := chromedp.Run(ctx,
				chromedp.Navigate(results[i].Link),
				chromedp.WaitVisible("body"),
				chromedp.Evaluate("document.body.innerText", &content),
			)
			if err != nil {
				// Log warning instead of failing the whole operation
				fmt.Printf("Warning: could not fetch content for %s: %v\n", results[i].Link, err)
				continue
			}
			if len(content) > 2000 {
				content = content[:2000] + "..."
			}
			results[i].Content = content
		}
	}

	return results, nil
}

// GetContent extracts content from a URL or the current page.
func GetContent(ctx context.Context, targetURL, format string) (map[string]interface{}, error) {
	if targetURL != "" {
		err := chromedp.Run(ctx,
			chromedp.Navigate(targetURL),
			chromedp.WaitVisible("body"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to '%s': %w", targetURL, err)
		}
	}

	var content, title, currentURL string
	err := chromedp.Run(ctx,
		chromedp.InnerHTML("body", &content),
		chromedp.Title(&title),
		chromedp.Location(&currentURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract page content: %w", err)
	}

	if targetURL == "" {
		targetURL = currentURL
	}

	var processedContent string
	switch format {
	case "text":
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse html: %w", err)
		}
		processedContent = strings.TrimSpace(doc.Text())
	case "markdown":
		converter := md.NewConverter("", true, nil)
		processedContent, err = converter.ConvertString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to markdown: %w", err)
		}
	case "html":
		processedContent = content
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	result := map[string]interface{}{
		"title":   title,
		"content": processedContent,
		"format":  format,
		"url":     targetURL,
	}
	return result, nil
}

// HnScraper scrapes top stories from Hacker News.
func HnScraper(ctx context.Context, limit int) ([]models.HnSubmission, error) {
	sel := config.GetSelectors().HN
	hnURL := "https://news.ycombinator.com"
	err := chromedp.Run(ctx,
		chromedp.Navigate(hnURL),
		chromedp.WaitVisible(sel.Container),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to hacker news: %w", err)
	}

	var titles, urls, scoreTexts, authorTexts, timeTexts, commentTexts []string

	titleJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(a => a.textContent)`, sel.Title)
	urlJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(a => a.href)`, sel.URL)
	scoreJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(el => el.textContent)`, sel.Score)
	authorJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(el => el.textContent)`, sel.Author)
	timeJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).map(el => el.title || el.textContent)`, sel.Time)
	commentsJS := fmt.Sprintf(`Array.from(document.querySelectorAll('%s')).filter(a => a.textContent.includes('comment')).map(a => a.textContent)`, sel.Comments)

	err = chromedp.Run(ctx,
		chromedp.Evaluate(titleJS, &titles),
		chromedp.Evaluate(urlJS, &urls),
		chromedp.Evaluate(scoreJS, &scoreTexts),
		chromedp.Evaluate(authorJS, &authorTexts),
		chromedp.Evaluate(timeJS, &timeTexts),
		chromedp.Evaluate(commentsJS, &commentTexts),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract data from hacker news: %w", err)
	}

	minLen := len(titles)
	if limit > 0 && limit < minLen {
		minLen = limit
	}

	submissions := make([]models.HnSubmission, 0, minLen)
	rePoints := regexp.MustCompile(`\d+`)
	for i := 0; i < minLen; i++ {
		points := 0
		if i < len(scoreTexts) {
			p, _ := strconv.Atoi(rePoints.FindString(scoreTexts[i]))
			points = p
		}

		comments := 0
		if i < len(commentTexts) {
			c, _ := strconv.Atoi(rePoints.FindString(commentTexts[i]))
			comments = c
		}

		author := ""
		if i < len(authorTexts) {
			author = authorTexts[i]
		}

		submissions = append(submissions, models.HnSubmission{
			ID:       fmt.Sprintf("%d", i+1),
			Title:    titles[i],
			URL:      urls[i],
			Points:   points,
			Author:   author,
			Time:     timeTexts[i],
			Comments: comments,
			HnURL:    "", // HN URL is harder to get reliably, leave for now
		})
	}

	return submissions, nil
}
