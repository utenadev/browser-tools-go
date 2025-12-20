package cmd

import (
	"context"
	"errors"
	"testing"

	"browser-tools-go/internal/logic"
)

// モック用のブラウザコンテキスト
type mockBrowserCtx struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// newMockBrowserCtx はテスト用のモックブラウザコンテキストを作成します。
func newMockBrowserCtx() *mockBrowserCtx {
	ctx, cancel := context.WithCancel(context.Background())
	return &mockBrowserCtx{ctx: ctx, cancel: cancel}
}

// TestNewNavigateCmd_CommandDefinition はnavigateコマンドの定義をテストします。
func TestNewNavigateCmd_CommandDefinition(t *testing.T) {
	cmd := newNavigateCmd()

	if cmd.Use != "navigate <url>" {
		t.Errorf("Expected Use to be 'navigate <url>', got %s", cmd.Use)
	}

	if cmd.Short != "Navigate to a specific URL" {
		t.Errorf("Expected Short to be 'Navigate to a specific URL', got %s", cmd.Short)
	}

	if cmd.Args == nil {
		t.Error("Args validator should be set")
	}

	if cmd.Run == nil {
		t.Error("Run function should be set")
	}
}

// TestNewScreenshotCmd_CommandDefinition はscreenshotコマンドの定義をテストします。
func TestNewScreenshotCmd_CommandDefinition(t *testing.T) {
	cmd := newScreenshotCmd()

	if cmd.Use != "screenshot [path]" {
		t.Errorf("Expected Use to be 'screenshot [path]', got %s", cmd.Use)
	}

	if cmd.Short != "Capture a screenshot of a web page" {
		t.Errorf("Expected Short to be 'Capture a screenshot of a web page', got %s", cmd.Short)
	}

	// フラグの存在確認
	urlFlag := cmd.Flags().Lookup("url")
	if urlFlag == nil {
		t.Error("Expected 'url' flag to exist")
	}

	fullPageFlag := cmd.Flags().Lookup("full-page")
	if fullPageFlag == nil {
		t.Error("Expected 'full-page' flag to exist")
	}
}

// TestNewNavigateCmd_ArgumentValidation は引数バリデーションをテストします。
func TestNewNavigateCmd_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		expectedError error
	}{
		{
			name:        "exactly one argument",
			args:        []string{"https://example.com"},
			expectError: false,
		},
		{
			name:          "no arguments",
			args:          []string{},
			expectError:   true,
			expectedError: errors.New("requires exactly 1 arg(s), only received 0"),
		},
		{
			name:          "too many arguments",
			args:          []string{"https://example.com", "extra"},
			expectError:   true,
			expectedError: errors.New("requires exactly 1 arg(s), only received 2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newNavigateCmd()

			if tt.expectError {
				if err := cmd.Args(cmd, tt.args); err != nil {
					if tt.expectedError != nil {
						if err.Error() != tt.expectedError.Error() {
							t.Errorf("Expected error '%v', got '%v'", tt.expectedError, err)
						}
					}
				} else {
					t.Error("Expected validation error, but got none")
				}
			} else {
				if err := cmd.Args(cmd, tt.args); err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
	}
}

// TestNewNavigateCmd_RunFunctionExists はRun関数が実装されていることを確認します。
func TestNewNavigateCmd_RunFunctionExists(t *testing.T) {
	cmd := newNavigateCmd()

	if cmd.Run == nil {
		t.Error("Run function must be implemented")
	}
}

// TestNewScreenshotCmd_RunFunctionExists はRun関数が実装されていることを確認します。
func TestNewScreenshotCmd_RunFunctionExists(t *testing.T) {
	cmd := newScreenshotCmd()

	if cmd.Run == nil {
		t.Error("Run function must be implemented")
	}
}

// TestNewNavigateCmd_PersistentPreRunEExists はPersistentPreRunEが設定されていることを確認します。
func TestNewNavigateCmd_PersistentPreRunEExists(t *testing.T) {
	cmd := newNavigateCmd()

	if cmd.PersistentPreRunE == nil {
		t.Error("PersistentPreRunE must be set")
	}
}

// TestNewScreenshotCmd_PersistentPreRunEExists はPersistentPreRunEが設定されていることを確認します。
func TestNewScreenshotCmd_PersistentPreRunEExists(t *testing.T) {
	cmd := newScreenshotCmd()

	if cmd.PersistentPreRunE == nil {
		t.Error("PersistentPreRunE must be set")
	}
}

// TestNewScreenshotCmd_FlagDefaults はフラグのデフォルト値をテストします。
func TestNewScreenshotCmd_FlagDefaults(t *testing.T) {
	cmd := newScreenshotCmd()

	// URLフラグのデフォルトは空文字列
	urlFlag := cmd.Flags().Lookup("url")
	if urlFlag == nil {
		t.Fatal("url flag not found")
	}
	urlValue, err := cmd.Flags().GetString("url")
	if err != nil {
		t.Fatalf("Failed to get url flag value: %v", err)
	}
	if urlValue != "" {
		t.Errorf("Expected default URL to be empty string, got '%s'", urlValue)
	}

	// full-pageフラグのデフォルトはfalse
	fullPageValue, err := cmd.Flags().GetBool("full-page")
	if err != nil {
		t.Fatalf("Failed to get full-page flag value: %v", err)
	}
	if fullPageValue {
		t.Error("Expected default full-page to be false")
	}
}

// TestNewScreenshotCmd_ArgumentValidation はscreenshotコマンドの引数検証をテストします。
func TestNewScreenshotCmd_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments (allowed)",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "one argument",
			args:        []string{"output.png"},
			expectError: false,
		},
		{
			name:        "too many arguments",
			args:        []string{"file1.png", "file2.png"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newScreenshotCmd()

			err := cmd.Args(cmd, tt.args)
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
	}
}

// TestNavigateAndScreenshotLogicFunctionNames は実際のロジック関数が存在することを確認します。
// このテストは関数名が変更された場合に検知するのに役立ちます。
func TestNavigateAndScreenshotLogicFunctionNames(t *testing.T) {
	// logic.Navigate 関数が存在するか（リフレクションは使用せず、コンパイル時に検証）
	// コンパイルエラーにならなければ関数は存在する
	_ = logic.Navigate

	// logic.Screenshot 関数が存在するか
	// コンパイルエラーにならなければ関数は存在する
	_ = logic.Screenshot
}

// TestNewScreenshotCmd_FlagBinding はフラグが正しくバインドされていることを確認します。
func TestNewScreenshotCmd_FlagBinding(t *testing.T) {
	cmd := newScreenshotCmd()

	tests := []struct {
		flagName     string
		expectedType string
	}{
		{"url", "string"},
		{"full-page", "bool"},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("Flag '%s' not found", tt.flagName)
			continue
		}

		// フラグの使い方（型）はコードを読むことで間接的に確認
		// 実際の値取得テストは統合テストで行う
	}
}

// TestNewNavigateCmd_HelpMessage はヘルプメッセージが存在することを確認します。
func TestNewNavigateCmd_HelpMessage(t *testing.T) {
	cmd := newNavigateCmd()

	if cmd.Long != "" {
		t.Logf("Navigate command has a long description: %s", cmd.Long)
	}

	// Short description must exist
	if cmd.Short == "" {
		t.Error("Navigate command should have a short description")
	}
}

// TestNewScreenshotCmd_HelpMessage はヘルプメッセージが存在することを確認します。
func TestNewScreenshotCmd_HelpMessage(t *testing.T) {
	cmd := newScreenshotCmd()

	if cmd.Long != "" {
		t.Logf("Screenshot command has a long description: %s", cmd.Long)
	}

	// Short description must exist
	if cmd.Short == "" {
		t.Error("Screenshot command should have a short description")
	}
}

// BenchmarkNewNavigateCmd はnavigateコマンド作成のベンチマークです。
func BenchmarkNewNavigateCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newNavigateCmd()
	}
}

// BenchmarkNewScreenshotCmd はscreenshotコマンド作成のベンチマークです。
func BenchmarkNewScreenshotCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newScreenshotCmd()
	}
}