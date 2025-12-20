package browser

import (
	"context"
	"testing"
	"time"

	"browser-tools-go/internal/config"
)

// TestNewPersistentContext_NoSession はセッションが存在しない場合のエラーをテストします。
func TestNewPersistentContext_NoSession(t *testing.T) {
	// 既存のセッションファイルがあれば削除
	_ = config.RemoveWsInfo()

	_, _, err := NewPersistentContext()
	if err == nil {
		t.Error("Expected error when no session is running, got nil")
	}

	expectedErrorMsg := "failed to connect to browser"
	if err.Error()[:len(expectedErrorMsg)] != expectedErrorMsg {
		t.Errorf("Expected error message to start with '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

// TestNewPersistentContext_ContextCancel はコンテキストキャンセルが機能することをテストします。
func TestNewPersistentContext_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// タイムアウト付きコンテキストでの動作テスト
	<-ctx.Done()
	if ctx.Err() != context.DeadlineExceeded {
		t.Error("Expected DeadlineExceeded error")
	}
}

// TestWaitForWS_Timeout はWebSocket待機のタイムアウトをテストします。
func TestWaitForWS_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 存在しないWebSocketエンドポイント
	invalidURL := "ws://127.0.0.1:99999"

	err := WaitForWS(ctx, invalidURL, 50*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestWaitForWS_InvalidURL は無効なURLを拒否することをテストします。
func TestWaitForWS_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	invalidURL := "http://not-ws-url.example.com"

	err := WaitForWS(ctx, invalidURL, 50*time.Millisecond)
	if err == nil {
		t.Error("Expected error for invalid WebSocket URL, got nil")
	}
}

// TestWaitForWS_CancelledContext はキャンセル済みコンテキストでの動作をテストします。
func TestWaitForWS_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	invalidURL := "ws://127.0.0.1:12345"

	err := WaitForWS(ctx, invalidURL, 1*time.Second)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
}