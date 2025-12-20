package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SelectorConfig はWebサイトのセレクタ設定を保持します
type SelectorConfig struct {
	GoogleSearch *GoogleSearchSelectors `json:"google_search"`
	HackerNews   *HackerNewsSelectors   `json:"hacker_news"`
}

// GoogleSearchSelectors はGoogle検索のセレクタ定義です
type GoogleSearchSelectors struct {
	SearchContainer []string `json:"search_container"`
	ResultItem      []string `json:"result_item"`
	Title           []string `json:"title"`
	URL             []string `json:"url"`
	Snippet         []string `json:"snippet"`
	FallbackWait    []string `json:"fallback_wait"`
}

// HackerNewsSelectors はHacker Newsのセレクタ定義です
type HackerNewsSelectors struct {
	MainTable       []string `json:"main_table"`
	TitleLink       []string `json:"title_link"`
	Score           []string `json:"score"`
	Author          []string `json:"author"`
	Time            []string `json:"time"`
	Comments        []string `json:"comments"`
	FallbackWait    []string `json:"fallback_wait"`
}

// DefaultSelectorConfig はデフォルトのセレクタ設定です
func DefaultSelectorConfig() *SelectorConfig {
	return &SelectorConfig{
		GoogleSearch: &GoogleSearchSelectors{
			SearchContainer: []string{"div#search", "div#rso", "div.g"},
			ResultItem:      []string{"div.g", "div.rc", "div.Gx5Zad"},
			Title:           []string{"h3", "h3.LC20lb", "div.v9i61e"},
			URL:             []string{"a", "a[href]", "a[ping]"},
			Snippet:         []string{"div.VwiC3b", "div.s", "div.BNeawe"},
			FallbackWait:    []string{"div#search", "div.g", "body"},
		},
		HackerNews: &HackerNewsSelectors{
			MainTable:    []string{"table.itemlist", "table#hnmain", "table"},
			TitleLink:    []string{"span.titleline > a", "a.storylink", "td.title > a"},
			Score:        []string{".score", ".subtext .score"},
			Author:       []string{".hnuser", ".subtext a.hnuser", "td.subtext a[href*=\"user?id=\"]"},
			Time:         []string{"span.age a", ".subtext span.age a", "td.subtext span.age"},
			Comments:     []string{"td.subtext > a:last-child", "a[href*=\"item?id=\"]"},
			FallbackWait: []string{"table.itemlist", "body"},
		},
	}
}

// LoadSelectorConfig はセレクタ設定ファイルを読み込みます
// ファイルが存在しない場合、デフォルト設定を返します
func LoadSelectorConfig(configPath string) (*SelectorConfig, error) {
	if configPath == "" {
		// デフォルトパスを設定
		home, err := os.UserHomeDir()
		if err != nil {
			return DefaultSelectorConfig(), nil
		}
		configPath = filepath.Join(home, ".browser-tools-go", "selectors.json")
	}

	// ファイルの存在確認
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// ファイルが存在しない場合はデフォルトを返す
		return DefaultSelectorConfig(), nil
	}

	// ファイルを読み込み
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SelectorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// デフォルト値で補完
	config.mergeWithDefaults()

	return &config, nil
}

// SaveSelectorConfig はセレクタ設定をファイルに保存します
func SaveSelectorConfig(config *SelectorConfig, configPath string) error {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configPath = filepath.Join(home, ".browser-tools-go", "selectors.json")
	}

	// ディレクトリの作成
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

// mergeWithDefaults は設定をデフォルト値で補完します
func (c *SelectorConfig) mergeWithDefaults() {
	defaults := DefaultSelectorConfig()

	if c.GoogleSearch == nil {
		c.GoogleSearch = defaults.GoogleSearch
	} else {
		c.GoogleSearch = mergeGoogleSearchSelectors(c.GoogleSearch, defaults.GoogleSearch)
	}

	if c.HackerNews == nil {
		c.HackerNews = defaults.HackerNews
	} else {
		c.HackerNews = mergeHackerNewsSelectors(c.HackerNews, defaults.HackerNews)
	}
}

func mergeGoogleSearchSelectors(current, defaults *GoogleSearchSelectors) *GoogleSearchSelectors {
	if len(current.SearchContainer) == 0 {
		current.SearchContainer = defaults.SearchContainer
	}
	if len(current.ResultItem) == 0 {
		current.ResultItem = defaults.ResultItem
	}
	if len(current.Title) == 0 {
		current.Title = defaults.Title
	}
	if len(current.URL) == 0 {
		current.URL = defaults.URL
	}
	if len(current.Snippet) == 0 {
		current.Snippet = defaults.Snippet
	}
	if len(current.FallbackWait) == 0 {
		current.FallbackWait = defaults.FallbackWait
	}
	return current
}

func mergeHackerNewsSelectors(current, defaults *HackerNewsSelectors) *HackerNewsSelectors {
	if len(current.MainTable) == 0 {
		current.MainTable = defaults.MainTable
	}
	if len(current.TitleLink) == 0 {
		current.TitleLink = defaults.TitleLink
	}
	if len(current.Score) == 0 {
		current.Score = defaults.Score
	}
	if len(current.Author) == 0 {
		current.Author = defaults.Author
	}
	if len(current.Time) == 0 {
		current.Time = defaults.Time
	}
	if len(current.Comments) == 0 {
		current.Comments = defaults.Comments
	}
	if len(current.FallbackWait) == 0 {
		current.FallbackWait = defaults.FallbackWait
	}
	return current
}

// FirstMatchingSelector は複数のセレクタ候補から最初にマッチしたものを返します
// すべてのセレクタが失敗した場合は空文字列を返します
func FirstMatchingSelector(candidates []string) string {
	for _, selector := range candidates {
		if selector != "" {
			return selector
		}
	}
	return ""
}

// JoinSelectors は複数のセレクタをカンマ区切りで結合します
func JoinSelectors(selectors []string) string {
	// 空のセレクタを除外
	validSelectors := make([]string, 0, len(selectors))
	for _, s := range selectors {
		s = strings.TrimSpace(s)
		if s != "" {
			validSelectors = append(validSelectors, s)
		}
	}
	return strings.Join(validSelectors, ", ")
}

// GenerateAlternativeSelectors は与えられたセレクタに基づいて代替セレクタを生成します
func GenerateAlternativeSelectors(baseSelector string) []string {
	alternatives := []string{baseSelector}

	// セレクタを解析して代替案を生成
	parts := strings.Split(baseSelector, " ")
	if len(parts) > 1 {
		// 最後の要素のみ
		alternatives = append(alternatives, parts[len(parts)-1])
		// クラス指定なし
		lastPart := parts[len(parts)-1]
		if strings.Contains(lastPart, ".") || strings.Contains(lastPart, "#") {
			tag := strings.TrimRight(lastPart, ".#0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-")
			if tag != "" {
				alternatives = append(alternatives, tag)
			}
		}
	}

	return alternatives
}

// ValidateSelectorSyntax はセレクタの文法を簡単に検証します
func ValidateSelectorSyntax(selector string) error {
	if selector == "" {
		return fmt.Errorf("empty selector")
	}

	// 基本的な文法チェック
	if strings.Contains(selector, "..") {
		return fmt.Errorf("invalid selector: contains '..'")
	}

	// XPathのようなパス表記を拒否
	if strings.HasPrefix(selector, "/") || strings.Contains(selector, "//") {
		return fmt.Errorf("invalid selector: XPath syntax not supported")
	}

	return nil
}

// FormatSelectorForJS はセレクタをJavaScriptで安全に扱える形式にエスケープします
func FormatSelectorForJS(selector string) string {
	// シングルクォートをエスケープ
	escaped := strings.ReplaceAll(selector, `'`, `\'`)
	// バックスラッシュをエスケープ
	escaped = strings.ReplaceAll(escaped, `\`, `\\`)
	return escaped
}