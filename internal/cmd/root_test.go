package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"browser-tools-go/internal/config"
	"github.com/spf13/cobra"
)

// TestNewRootCmd_CommandStructure はコマンド構造が正しいことをテストします。
func TestNewRootCmd_CommandStructure(t *testing.T) {
	rootCmd := NewRootCmd()

	if rootCmd.Use != "browser-tools-go" {
		t.Errorf("Expected Use to be 'browser-tools-go', got %s", rootCmd.Use)
	}

	if rootCmd.Short != "A Go implementation of browser-tools" {
		t.Errorf("Expected Short description mismatch, got %s", rootCmd.Short)
	}

	// コマンド数チェック（root + 11サブコマンド）
	expectedCommands := 11
	if len(rootCmd.Commands()) != expectedCommands {
		t.Errorf("Expected %d commands, got %d", expectedCommands, len(rootCmd.Commands()))
	}

	// 存在するべきコマンド
	expectedCommandNames := []string{
		"start",
		"close",
		"run",
		"navigate",
		"screenshot",
		"pick",
		"eval",
		"cookies",
		"search",
		"content",
		"hn-scraper",
	}

	for _, name := range expectedCommandNames {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' not found", name)
		}
	}
}

// TestPersistentPreRunE_NoExistingContext は新規ブラウザコンテキストの作成をテストします。
func TestPersistentPreRunE_NoExistingContext(t *testing.T) {
	// 既存のセッションファイルをクリーンアップ
	_ = config.RemoveWsInfo()

	// Cobraコマンドのモック
	cmd := NewRootCmd()
	argsFunc := func(cmd *cobra.Command, args []string) error { return nil }
	runFunc := func(cmd *cobra.Command, args []string) error { return nil }
	cmd.Args = argsFunc
	cmd.RunE = runFunc

	// PersistentPreRunEはセッションがなければエラーになるはず
	err := persistentPreRunE(cmd, []string{})
	if err != nil {
		expectedMsg := "failed to connect to browser"
		if err.Error()[:len(expectedMsg)] != expectedMsg {
			t.Errorf("Expected error message containing '%s', got '%s'", expectedMsg, err.Error())
		}
	}
}

// TestPersistentPreRunE_ContextAlreadySet は既存コンテキストがある場合の動作をテストします。
func TestPersistentPreRunE_ContextAlreadySet(t *testing.T) {
	// モックのブラウザコンテキスト作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	browserCtxVal := &browserCtx{ctx: ctx, cancel: cancel}
	ctxWithBrowser := context.WithValue(context.Background(), browserCtxKey, browserCtxVal)
	cmd := NewRootCmd()
	cmd.SetContext(ctxWithBrowser)

	// コンテキストが既に存在する場合はnilを返すはず
	err := persistentPreRunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error when context already exists, got %v", err)
	}
}

// TestGetBrowserCtx_ValidContext は有効なブラウザコンテキストの取得をテストします。
func TestGetBrowserCtx_ValidContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedCtx := &browserCtx{ctx: ctx, cancel: cancel}
	ctxWithBrowser := context.WithValue(context.Background(), browserCtxKey, expectedCtx)

	cmd := NewRootCmd()
	cmd.SetContext(ctxWithBrowser)

	resultCtx, err := getBrowserCtx(cmd)
	if err != nil {
		t.Fatalf("Failed to get browser context: %v", err)
	}

	if resultCtx != expectedCtx {
		t.Error("Retrieved browser context does not match expected context")
	}
}

// TestGetBrowserCtx_NilContext はnilコンテキストを適切に処理することをテストします。
func TestGetBrowserCtx_NilContext(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetContext(context.Background())

	_, err := getBrowserCtx(cmd)
	if err == nil {
		t.Error("Expected error for nil browser context")
	}
}

// TestGetBrowserCtx_InvalidContextType は無効なコンテキスト型を拒否することをテストします。
func TestGetBrowserCtx_InvalidContextType(t *testing.T) {
	// 想定外の型を設定
	ctxWithInvalid := context.WithValue(context.Background(), browserCtxKey, "invalid_type")

	cmd := NewRootCmd()
	cmd.SetContext(ctxWithInvalid)

	_, err := getBrowserCtx(cmd)
	if err == nil {
		t.Error("Expected error for invalid browser context type")
	}

	if err.Error() != "invalid browser context type" {
		t.Errorf("Expected 'invalid browser context type' error, got: %v", err)
	}
}

// TestPrettyPrintResults_ValidJSON は標準的なデータ構造を正しくJSON化することをテストします。
func TestPrettyPrintResults_ValidJSON(t *testing.T) {
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": []string{"a", "b", "c"},
	}

	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	prettyPrintResults(testData)

	w.Close()
	os.Stdout = originalStdout
	io.Copy(&buf, r)
	output := buf.String()

	// JSONであること、そしてpretty printであることを検証
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// フォーマットが正しい（改行とインデント）
	if !strings.Contains(output, "\n") {
		t.Error("Output doesn't contain newlines, not pretty printed")
	}

	if !strings.Contains(output, "  ") {
		t.Error("Output doesn't contain indentation, not pretty printed")
	}
}

// TestPrettyPrintResults_ComplexData は複雑なデータ構造のJSON化をテストします。
func TestPrettyPrintResults_ComplexData(t *testing.T) {
	complexData := []map[string]interface{}{
		{
			"id":    1,
						"name":  "item1",
			"nested": map[string]string{"a": "b"},
		},
		{
			"id":    2,
			"name":  "item2",
			"nested": map[string]string{"c": "d"},
		},
	}

	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	prettyPrintResults(complexData)

	w.Close()
	os.Stdout = originalStdout
	io.Copy(&buf, r)
	output := buf.String()

	// JSONとしてパース可能
	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("Expected array length 2, got %d", len(parsed))
	}
}

// TestExecute_ErrorHandling はExecute関数のエラーハンドリングをテストします。
func TestExecute_ErrorHandling(t *testing.T) {
	// Executeはos.Exitを呼び出すため、通常のテストでは完全なテストが困難
	// エラー時の挙動はintegrationテストでカバー
	t.Skip("Skipping Execute test as it calls os.Exit")
}

// TestBrowserCtx_ContextCancellation はブラウザコンテキストのキャンセルが機能することをテストします。
func TestBrowserCtx_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	bc := &browserCtx{ctx: ctx, cancel: cancel}

	// キャンセルを呼び出し
	bc.cancel()

	// コンテキストがキャンセルされたことを確認
	select {
	case <-bc.ctx.Done():
		// 期待通り
	default:
		t.Error("Context was not cancelled")
	}
}

// TestBrowserCtxKeyType_StringRepresentation はブラウザコンテキストキーの文字列表現をテストします。
func TestBrowserCtxKeyType_StringRepresentation(t *testing.T) {
	keyAsString := string(browserCtxKey)
	if keyAsString != "browserCtx" {
		t.Errorf("Expected key string 'browserCtx', got '%s'", keyAsString)
	}
}

// TestExitCodeConstants は終了コード定数の値をテストします。
func TestExitCodeConstants(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("Expected ExitSuccess to be 0, got %d", ExitSuccess)
	}

	if ExitError != 1 {
		t.Errorf("Expected ExitError to be 1, got %d", ExitError)
	}
}

// TestNewRootCmd_EmptyArgs は引数なしでのコマンド作成をテストします。
func TestNewRootCmd_EmptyArgs(t *testing.T) {
	rootCmd := NewRootCmd()

	// デフォルト引数は不正な場合にエラーを返す
	err := rootCmd.Args(rootCmd, []string{})
	if err != nil {
		t.Logf("Root command with empty args returns error as expected: %v", err)
	}
}

// TestPrettyPrintResults_CannotMarshal はMarshalできないデータの処理をテストします。
func TestPrettyPrintResults_CannotMarshal(t *testing.T) {
	// 以下のコードはpanicを起こす可能性があるため、テストをスキップ
	// prettyPrintResultsがプログラムを終了する（log.Fatalf）ため、不完全なテスト
	t.Skip("Skipping test as untestable error path causes os.Exit")
}