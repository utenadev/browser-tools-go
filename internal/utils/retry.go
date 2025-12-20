package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// RetryableError はリトライ可能なエラーを示します
type RetryableError struct {
	Err          error
	Message      string
	ShouldRetry  bool
}

func (e *RetryableError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

// Unwrap implements the errors.Unwrap interface
func (e *RetryableError) Unwrap() error {
	return e.Err
}

// RetryConfig はリトライ設定を保持します
type RetryConfig struct {
	MaxAttempts       int           // 最大リトライ回数（初回を含む）
	InitialBackoff    time.Duration // 初回バックオフ時間
	MaxBackoff        time.Duration // 最大バックオフ時間
	BackoffMultiplier float64       // バックオフ倍率（指数バックオフ）
	IsRetryable       func(error) bool // リトライ可能か判定する関数
	OnRetry           func(attempt int, err error) // リトライ時のコールバック
}

// DefaultRetryConfig はデフォルトのリトライ設定です
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        2 * time.Second,
		BackoffMultiplier: 2.0,
		IsRetryable:       DefaultIsRetryable,
		OnRetry:           DefaultOnRetry,
	}
}

// DefaultIsRetryable はデフォルトのリトライ判定関数です
func DefaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// リトライ不可なエラー
	nonRetryable := []string{
		"context canceled",
		"context deadline exceeded",
		"invalid argument",
		"not found",
		"forbidden",
		"unauthorized",
	}

	for _, keyword := range nonRetryable {
		if strings.Contains(strings.ToLower(errMsg), keyword) {
			return false
		}
	}

	// リトライ可能なエラー
	retryable := []string{
		"timeout",
		"connection refused",
		"no such host",
		"network",
		"temporary",
		"busy",
		"overloaded",
	}

	for _, keyword := range retryable {
		if strings.Contains(strings.ToLower(errMsg), keyword) {
			return true
		}
	}

	// デフォルトはリトライしない
	return false
}

// DefaultOnRetry はデフォルトのリトライコールバックです
func DefaultOnRetry(attempt int, err error) {
	log.Printf("Retry attempt %d after error: %v", attempt, err)
}

// Retry は指定された関数をリトライ設定に従って実行します
func Retry(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error

	for attempt := 0; ; attempt++ {
		// 関数実行
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// リトライ可否判定
		retryable := config.IsRetryable(err)
		if !retryable {
			return err
		}

		// リトライ回数超過
		if attempt >= config.MaxAttempts-1 {
			break
		}

		// コンテキストの確認
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry canceled: %w", ctx.Err())
		default:
		}

		// バックオフ計算
		backoff := calculateBackoff(attempt, config)

		// リトライ通知
		if config.OnRetry != nil {
			config.OnRetry(attempt+1, err)
		}

		// バックオフ待機
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry canceled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// RetryWithSelector はセレクタエラーありのリトライをサポートします
func RetryWithSelector(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
		config.MaxAttempts = 2 // セレクタ問題は最小限のリトライ
	}

	return Retry(ctx, fn, config)
}

// calculateBackoff はバックオフ時間を計算します
func calculateBackoff(attempt int, config *RetryConfig) time.Duration {
	backoff := config.InitialBackoff

	// 指数バックオフ
	for i := 0; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
			break
		}
	}

	return backoff
}

// WaitForElement は要素が見つかるまで待機します（複数セレクタ対応）
func WaitForElement(ctx context.Context, selectors []string) error {
	if len(selectors) == 0 {
		return fmt.Errorf("no selectors provided")
	}

	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// セレクタを試す
		for _, selector := range selectors {
			// ここで実際の要素検索ロジックを実装
			// この実装はプロジェクトのChromeDP統合で置き換える
			_ = selector // プレースホルダー
			return nil
		}

		attempt++
		if attempt >= 5 { // 最大5回試す
			return fmt.Errorf("element not found with any selector")
		}

		// 待機
		time.Sleep(500 * time.Millisecond)
	}
}

// NewTemporaryError は一時的なエラーを作成します
func NewTemporaryError(err error, message string) *RetryableError {
	return &RetryableError{
		Err:         err,
		Message:     message,
		ShouldRetry: true,
	}
}

// IsSelectorNotFoundError はセレクタが見つからないエラーか判定します
func IsSelectorNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	keywords := []string{
		"could not get nodes",
		"selector not found",
		"no elements found",
		"element not found",
	}

	for _, keyword := range keywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// ExponentialBackoff は指数バックオフを実装します
func ExponentialBackoff(attempt int, initial, max time.Duration, multiplier float64) time.Duration {
	if attempt == 0 {
		return 0
	}

	backoff := initial
	for i := 1; i < attempt; i++ {
		backoff = time.Duration(float64(backoff) * multiplier)
		if backoff > max {
			return max
		}
	}

	return backoff
}

// MaxRetriesExceededError は最大リトライ回数超過エラーです
type MaxRetriesExceededError struct {
	Attempts int
	LastErr  error
}

func (e *MaxRetriesExceededError) Error() string {
	return fmt.Sprintf("max retries exceeded after %d attempts: %v", e.Attempts, e.LastErr)
}

// IsMaxRetriesExceeded は最大リトライ超過エラーか判定します
func IsMaxRetriesExceeded(err error) bool {
	_, ok := err.(*MaxRetriesExceededError)
	return ok
}