package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"browser-tools-go/internal/models"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// Search performs a Google search and returns the results.
func Search(ctx context.Context, query string, numResults int, fetchContent bool) ([]models.SearchResult, error) {
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))

	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible("div#search"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to google and wait for results: %w", err)
	}

	var searchResultsJSON string
	script := `
		(() => {
			const results = [];
			const items = document.querySelectorAll('div#search div.g');
			for (let i = 0; i < items.length; i++) {
				const item = items[i];
				const titleEl = item.querySelector('h3');
				const linkEl = item.querySelector('a');
				const snippetEl = item.querySelector('div.VwiC3b');
				if (titleEl && linkEl && snippetEl) {
					results.push({
						title: titleEl.innerText,
						link: linkEl.href,
						snippet: snippetEl.innerText
					});
				}
			}
			return JSON.stringify(results);
		})();
	`
	err = chromedp.Run(ctx, chromedp.Evaluate(script, &searchResultsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to extract search results with script: %w", err)
	}

	var rawResults []map[string]string
	if err := json.Unmarshal([]byte(searchResultsJSON), &rawResults); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	minLen := len(rawResults)
	if numResults > 0 && numResults < minLen {
		minLen = numResults
	}

	results := make([]models.SearchResult, minLen)
	for i := 0; i < minLen; i++ {
		results[i] = models.SearchResult{
			Title:   rawResults[i]["title"],
			Link:    rawResults[i]["link"],
			Snippet: rawResults[i]["snippet"],
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
				log.Printf("Warning: could not fetch content for %s: %v\n", results[i].Link, err)
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
		processedContent = strings.TrimSpace(doc.Find("body").Text())
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
	hnURL := "https://news.ycombinator.com"
	err := chromedp.Run(ctx,
		chromedp.Navigate(hnURL),
		chromedp.WaitVisible("table.itemlist"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to hacker news: %w", err)
	}

	var titles, urls, scoreTexts, authorTexts, timeTexts, commentTexts []string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('span.titleline > a')).map(a => a.textContent)`, &titles),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('span.titleline > a')).map(a => a.href)`, &urls),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.score')).map(el => el.textContent)`, &scoreTexts),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.hnuser')).map(el => el.textContent)`, &authorTexts),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('span.age a')).map(el => el.title || el.textContent)`, &timeTexts),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('td.subtext > a')).filter(a => a.textContent.includes('comment')).map(a => a.textContent)`, &commentTexts),
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
