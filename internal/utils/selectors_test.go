package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultSelectorConfig はデフォルト設定生成をテストします
func TestDefaultSelectorConfig(t *testing.T) {
	config := DefaultSelectorConfig()

	if config.GoogleSearch == nil {
		t.Error("GoogleSearch selectors should not be nil")
	}

	if config.HackerNews == nil {
		t.Error("HackerNews selectors should not be nil")
	}

	// Google Searchセレクタが正しいことを確認
	if len(config.GoogleSearch.SearchContainer) == 0 {
		t.Error("SearchContainer should have selectors")
	}

	if len(config.GoogleSearch.ResultItem) == 0 {
		t.Error("ResultItem should have selectors")
	}

	if len(config.GoogleSearch.FallbackWait) == 0 {
		t.Error("FallbackWait should have selectors")
	}

	// Hacker Newsセレクタが正しいことを確認
	if len(config.HackerNews.MainTable) == 0 {
		t.Error("MainTable should have selectors")
	}

	if len(config.HackerNews.TitleLink) == 0 {
		t.Error("TitleLink should have selectors")
	}
}

// TestSaveAndLoadSelectorConfig はセレクタ設定の保存と読み込みをテストします
func TestSaveAndLoadSelectorConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "selectors.json")

	// デフォルト設定を作成
	originalConfig := DefaultSelectorConfig()

	// 保存
	err := SaveSelectorConfig(originalConfig, configPath)
	if err != nil {
		t.Fatalf("SaveSelectorConfig failed: %v", err)
	}

	// ファイルが存在することを確認
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// 読み込み
	loadedConfig, err := LoadSelectorConfig(configPath)
	if err != nil {
		t.Fatalf("LoadSelectorConfig failed: %v", err)
	}

	// 設定が一致することを確認
	if len(loadedConfig.GoogleSearch.SearchContainer) != len(originalConfig.GoogleSearch.SearchContainer) {
		t.Error("Loaded config does not match original config")
	}

	if len(loadedConfig.HackerNews.TitleLink) != len(originalConfig.HackerNews.TitleLink) {
		t.Error("Loaded HackerNews config does not match original config")
	}
}

// TestLoadSelectorConfig_NotExist は存在しないファイル読み込みをテストします
func TestLoadSelectorConfig_NotExist(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.json")
	config, err := LoadSelectorConfig(nonExistentPath)
	if err != nil {
		t.Fatalf("LoadSelectorConfig should not fail for non-existent file: %v", err)
	}

	if config == nil {
		t.Error("LoadSelectorConfig should return default config for non-existent file")
	}

	// 返された設定がデフォルトであることを確認
	if config.GoogleSearch == nil || config.HackerNews == nil {
		t.Error("Returned config should be default configuration")
	}
}

// TestMergeWithDefaults はデフォルト値のマージをテストします
func TestMergeWithDefaults(t *testing.T) {
	config := &SelectorConfig{
		GoogleSearch: &GoogleSearchSelectors{
			SearchContainer: []string{"custom.main"},
		},
	}

	config.mergeWithDefaults()

	// カスタム値が保持されていることを確認
	if config.GoogleSearch.SearchContainer[0] != "custom.main" {
		t.Error("Custom selector should be preserved")
	}

	// デフォルト値が追加されていることを確認
	if len(config.GoogleSearch.ResultItem) == 0 {
		t.Error("ResultItem should be merged from defaults")
	}

	if len(config.HackerNews.TitleLink) == 0 {
		t.Error("HackerNews should be populated from defaults")
	}
}

// TestFirstMatchingSelector はFirstMatchingSelector関数をテストします
func TestFirstMatchingSelector(t *testing.T) {
	tests := []struct {
		name      string
		candidates []string
		expected  string
	}{
		{
			name:      "first valid selector",
			candidates: []string{"div.test", "div.other", "span.link"},
			expected:  "div.test",
		},
		{
			name:      "skip empty strings",
			candidates: []string{"", "div.valid", "span.link"},
			expected:  "div.valid",
		},
		{
			name:      "all empty",
			candidates: []string{"", "", ""},
			expected:  "",
		},
		{
			name:      "single selector",
			candidates: []string{"h1.title"},
			expected:  "h1.title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstMatchingSelector(tt.candidates)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestJoinSelectors はJoinSelectors関数をテストします
func TestJoinSelectors(t *testing.T) {
	tests := []struct {
		name      string
		selectors []string
		expected  string
	}{
		{
			name:      "multiple selectors",
			selectors: []string{"div.test", "span.link", "a.title"},
			expected:  "div.test, span.link, a.title",
		},
		{
			name:      "skip empty selectors",
			selectors: []string{"div.test", "", "", "span.link"},
			expected:  "div.test, span.link",
		},
		{
			name:      "trim whitespace",
			selectors: []string{" div.test ", "  span.link  "},
			expected:  "div.test, span.link",
		},
		{
			name:      "single selector",
			selectors: []string{"div.test"},
			expected:  "div.test",
		},
		{
			name:      "empty list",
			selectors: []string{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinSelectors(tt.selectors)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestGenerateAlternativeSelectors はGenerateAlternativeSelectors関数をテストします
func TestGenerateAlternativeSelectors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		minLen int
	}{
		{
			name:   "simple selector",
			input:  "div.test",
			minLen: 1,
		},
		{
			name:   "nested selector",
			input:  "div.container div.item",
			minLen: 2,
		},
		{
			name:   "id selector",
			input:  "div#main.test",
			minLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateAlternativeSelectors(tt.input)
			if len(result) < tt.minLen {
				t.Errorf("Expected at least %d alternatives, got %d", tt.minLen, len(result))
			}

			// 元のセレクタが含まれていることを確認
			foundOriginal := false
			for _, s := range result {
				if s == tt.input {
					foundOriginal = true
					break
				}
			}
			if !foundOriginal {
				t.Error("Original selector should be included in alternatives")
			}
		})
	}
}

// TestValidateSelectorSyntax はValidateSelectorSyntax関数をテストします
func TestValidateSelectorSyntax(t *testing.T) {
	tests := []struct {
		name        string
		selector    string
		expectError bool
	}{
		{
			name:        "valid selector",
			selector:    "div.test",
			expectError: false,
		},
		{
			name:        "empty selector",
			selector:    "",
			expectError: true,
		},
		{
			name:        "invalid double dot",
			selector:    "div..test",
			expectError: true,
		},
		{
			name:        "XPath syntax",
			selector:    "/html/body",
			expectError: true,
		},
		{
			name:        "XPath double slash",
			selector:    "div//test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSelectorSyntax(tt.selector)
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestFormatSelectorForJS はFormatSelectorForJS関数をテストします
func TestFormatSelectorForJS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no escaping needed",
			input:    "div.test",
			expected: "div.test",
		},
		{
			name:     "escape single quote",
			input:    "div[id='test']",
			expected: "div[id=\\'test\\']",
		},
		{
			name:     "escape backslash",
			input:    "div\\test",
			expected: "div\\\\test",
		},
		{
			name:     "complex selector",
			input:    `div[class='test' data-value='\test']`,
			expected: `div[class=\\'test\\' data-value=\\'\\\\test\\']`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSelectorForJS(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestSelectorConfig_JSONSerialization はJSONシリアライゼーションをテストします
func TestSelectorConfig_JSONSerialization(t *testing.T) {
	config := DefaultSelectorConfig()

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var decoded SelectorConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// デコードされた設定が元の設定と同じ構造を持つことを確認
	if decoded.GoogleSearch == nil {
		t.Error("Decoded GoogleSearch should not be nil")
	}

	if decoded.HackerNews == nil {
		t.Error("Decoded HackerNews should not be nil")
	}
}

// BenchmarkFirstMatchingSelector はFirstMatchingSelectorのベンチマークです
func BenchmarkFirstMatchingSelector(b *testing.B) {
	candidates := []string{"", "div.first", "span.second", "a.third"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FirstMatchingSelector(candidates)
	}
}

// BenchmarkJoinSelectors はJoinSelectorsのベンチマークです
func BenchmarkJoinSelectors(b *testing.B) {
	selectors := []string{"div.test", "span.link", "a.title", "p.description"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JoinSelectors(selectors)
	}
}

// BenchmarkValidateSelectorSyntax はValidateSelectorSyntaxのベンチマークです
func BenchmarkValidateSelectorSyntax(b *testing.B) {
	selector := "div.container div.item a.link"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSelectorSyntax(selector)
	}
}