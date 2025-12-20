package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"browser-tools-go/internal/models"
	"browser-tools-go/internal/utils"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// FetchWithRetry はリトライ機能付きでWebページをフェッチします
func FetchWithRetry(ctx context.Context, targetURL string, maxRetries int) error {
	retryConfig := &utils.RetryConfig{
		MaxAttempts:       maxRetries,
		InitialBackoff:    500 * time.Millisecond,
		MaxBackoff:        2 * time.Second,
		BackoffMultiplier: 2.0,
		IsRetryable: func(err error) bool {
			// ネットワークエラーやタイムアウトはリトライ
			errMsg := strings.ToLower(err.Error())
			retryableErrors := []string{
				"timeout",
				"connection",
				"network",
				"temporary",
			}
			for _, keyword := range retryableErrors {
				if strings.Contains(errMsg, keyword) {
					return true
				}
			}
			return false
		},
		OnRetry: func(attempt int, err error) {
			log.Printf("Retry %d/3 for %s: %v", attempt, targetURL, err)
		},
	}

	fetchFn := func() error {
		return chromedp.Run(ctx, chromedp.Navigate(targetURL))
	}

	return utils.Retry(ctx, fetchFn, retryConfig)
}

// EnhancedSearch はセレクタフォールバック付きのGoogle検索です
func EnhancedSearch(ctx context.Context, query string, numResults int, fetchContent bool, config *utils.SelectorConfig) ([]models.SearchResult, error) {
	if config == nil {
		config = utils.DefaultSelectorConfig()
	}

	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query))

	// フェッチ（リトライ付き）
	if err := FetchWithRetry(ctx, searchURL, 3); err != nil {
		return nil, fmt.Errorf("failed to navigate to google: %w", err)
	}

	// ページ読み込み確認（複数のウエイトセレクタ）
	for _, selector := range config.GoogleSearch.FallbackWait {
		err := chromedp.Run(ctx, chromedp.WaitVisible(selector, chromedp.BySearch))
		if err == nil {
			break
		}
		log.Printf("Fallback selector failed: %s", selector)
	}

	// 検索結果抽出
	results, err := extractSearchResults(ctx, config.GoogleSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to extract search results: %w", err)
	}

	// 結果数制限
	minLen := len(results)
	if numResults > 0 && numResults < minLen {
		minLen = numResults
	}
	results = results[:minLen]

	// コンテンツ取得
	if fetchContent {
		resultsWithContent, err := fetchContentForResults(ctx, results, 3)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch content: %w", err)
		}
		results = resultsWithContent
	}

	return results, nil
}

// extractSearchResults は検索結果の抽出を試みます（複数セレクタ対応）
func extractSearchResults(ctx context.Context, selectors *utils.GoogleSearchSelectors) ([]models.SearchResult, error) {
	var lastErr error

	// 複数のセレクタ戦略を試す
	for _, itemSelector := range selectors.ResultItem {
		for _, titleSelector := range selectors.Title {
			for _, snippetSelector := range selectors.Snippet {
				results, err := tryExtractOneStrategy(ctx, itemSelector, titleSelector, snippetSelector)
				if err == nil && len(results) > 0 {
					log.Printf("Successfully extracted %d results with selectors: item=%s, title=%s, snippet=%s",
						len(results), itemSelector, titleSelector, snippetSelector)
					return results, nil
				}
				lastErr = err
				log.Printf("Selector strategy failed: item=%s, title=%s, snippet=%s: %v",
					itemSelector, titleSelector, snippetSelector, err)
			}
		}
	}

	return nil, fmt.Errorf("all selector strategies failed: %w", lastErr)
}

// tryExtractOneStrategy は1つのセレクタ戦略で抽出を試みます
func tryExtractOneStrategy(ctx context.Context, itemSel, titleSel, snippetSel string) ([]models.SearchResult, error) {
	var items []*cdp.Node
	err := chromedp.Run(ctx, chromedp.Nodes(itemSel, &items, chromedp.NodeVisible, chromedp.BySearch))
	if err != nil || len(items) == 0 {
		return nil, fmt.Errorf("failed to find result items: %w", err)
	}

	results := make([]models.SearchResult, 0, len(items))
	// セレクタエスケープ
	escapedTitleSel := utils.FormatSelectorForJS(titleSel)
	escapedSnippetSel := utils.FormatSelectorForJS(snippetSel)

	for i, item := range items {
		// JavaScriptによる要素抽出
		var titleText, snippetText, linkHref string

		extractScript := fmt.Sprintf(`
			(() => {
				const item = this;
				const titleEl = item.querySelector('%s');
				const linkEls = item.querySelectorAll('a');
				const snippetEl = item.querySelector('%s');

				const title = titleEl ? titleEl.innerText : '';
				const snippet = snippetEl ? snippetEl.innerText : '';
				const link = linkEls[0] ? linkEls[0].href : '';

				return {title, snippet, link};
			}).call(this);
		`, escapedTitleSel, escapedSnippetSel)

		var extractResult map[string]string
		err := chromedp.Run(ctx, chromedp.Evaluate(extractScript, &extractResult, func(p *cdproto.RuntimeEvaluateParams) *cdproto.RuntimeEvaluateParams {
			return p.WithObjectID(item.ObjectID)
		}))
		if err != nil {
			log.Printf("Failed to extract from item %d: %v", i, err)
			continue
		}

		if extractResult["title"] != "" && extractResult["link"] != "" {
			results = append(results, models.SearchResult{
				Title:   extractResult["title"],
				Link:    extractResult["link"],
				Snippet: extractResult["snippet"],
			})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no valid results extracted")
	}

	return results, nil
}

// fetchContentForResults は検索結果のコンテンツを取得します
func fetchContentForResults(ctx context.Context, results []models.SearchResult, maxConcurrent int) ([]models.SearchResult, error) {
	// 並列度を制限したコンテンツ取得
	semaphore := make(chan struct{}, maxConcurrent)

	for i := range results {
		semaphore <- struct{}{}
		go func(idx int) {
			defer func() { <-semaphore }()

			var content string
			err := chromedp.Run(ctx,
				chromedp.Navigate(results[idx].Link),
				chromedp.WaitVisible("body", chromedp.BySearch),
				chromedp.Evaluate("document.body.innerText", &content, chromedp.EvalIgnoreExceptions),
			)
			if err != nil {
				log.Printf("Warning: could not fetch content for %s: %v", results[idx].Link, err)
				return
			}

			// コンテンツの切り詰め
			if len(content) > 2000 {
				content = content[:2000] + "..."
			}
			results[idx].Content = content
		}(i)
	}

	// すべてのゴルーチンが完了するまで待機
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	return results, nil
}

// EnhancedHnScraper は強化版Hacker Newsスクレイパーです
func EnhancedHnScraper(ctx context.Context, limit int, config *utils.SelectorConfig) ([]models.HnSubmission, error) {
	if config == nil {
		config = utils.DefaultSelectorConfig()
	}

	hnURL := "https://news.ycombinator.com"

	// フェッチ（リトライ付き）
	if err := FetchWithRetry(ctx, hnURL, 3); err != nil {
		return nil, fmt.Errorf("failed to navigate to hacker news: %w", err)
	}

	// ページ読み込み確認
	var waitErr error
	for _, selector := range config.HackerNews.FallbackWait {
		err := chromedp.Run(ctx, chromedp.WaitVisible(selector, chromedp.BySearch))
		if err == nil {
			break
		}
		waitErr = err
	}

	if waitErr != nil {
		return nil, fmt.Errorf("failed to wait for hacker news page: %w", waitErr)
	}

	// データ抽出
	return extractHnData(ctx, limit, config.HackerNews)
}

// extractHnData はHacker Newsのデータを抽出します
func extractHnData(ctx context.Context, limit int, selectors *utils.HackerNewsSelectors) ([]models.HnSubmission, error) {
	var titles, urls, scoreTexts, authorTexts, timeTexts, commentTexts []string

	// 改良版抽出ロジック
	extractScript := fmt.Sprintf(`
		(() => {
			const titles = [];
			const urls = [];
			const scores = [];
			const authors = [];
			const times = [];
			const comments = [];

			// タイトルとURLの抽出
			const titleLinks = document.querySelectorAll('%s');
			titleLinks.forEach(el => {
				titles.push(el.textContent.trim());
				urls.push(el.href);
			});

			// スコアの抽出
			const scoreEls = document.querySelectorAll('%s');
			scoreEls.forEach(el => scores.push(el.textContent));

			// 著者の抽出
			const authorEls = document.querySelectorAll('%s');
			authorEls.forEach(el => authors.push(el.textContent));

			// 時間の抽出
			const timeEls = document.querySelectorAll('%s');
			timeEls.forEach(el => times.push(el.textContent || el.title));

			// コメント数の抽出
			const commentEls = document.querySelectorAll('td.subtext > a');
			commentEls.forEach(el => {
				if (el.textContent.includes('comment') || el.textContent.match(/\\d+\\s*comments?/i)) {
					comments.push(el.textContent);
				}
			});

			return {
				titles, urls, scores, authors, times, comments
			};
		})();
	`,
		utils.JoinSelectors(selectors.TitleLink),
		utils.JoinSelectors(selectors.Score),
		utils.JoinSelectors(selectors.Author),
		utils.JoinSelectors(selectors.Time),
	)

	err := chromedp.Run(ctx, chromedp.Evaluate(extractScript, &map[string]interface{}{
		"titles":   &titles,
		"urls":     &urls,
		"scores":   &scoreTexts,
		"authors":  &authorTexts,
		"times":    &timeTexts,
		"comments": &commentTexts,
	}, chromedp.EvalIgnoreExceptions))
	if err != nil {
		return nil, fmt.Errorf("failed to extract hacker news data: %w", err)
	}

	// 結果の構築
	minLen := utils.Min(len(titles), limit)
	if limit <= 0 || limit > len(titles) {
		minLen = len(titles)
	}

	submissions := make([]models.HnSubmission, 0, minLen)
	rePoints := regexp.MustCompile(`\d+`)

	for i := 0; i < minLen; i++ {
		points := 0
		if i < len(scoreTexts) {
			points, _ = strconv.Atoi(rePoints.FindString(scoreTexts[i]))
		}

		comments := 0
		if i < len(commentTexts) {
			comments, _ = strconv.Atoi(rePoints.FindString(commentTexts[i]))
		}

		author := ""
		if i < len(authorTexts) {
			author = authorTexts[i]
		}

		timeText := ""
		if i < len(timeTexts) {
			timeText = timeTexts[i]
		}

		submissions = append(submissions, models.HnSubmission{
			ID:       fmt.Sprintf("%d", i+1),
			Title:    titles[i],
			URL:      urls[i],
			Points:   points,
			Author:   author,
			Time:     timeText,
			Comments: comments,
			HnURL:    "", // 必要に応じて設定
		})
	}

	return submissions, nil
}

// Min is a utility function to find minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}