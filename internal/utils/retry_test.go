package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestRetryableError_Error はRetryableErrorのErrorメソッドをテストします
func TestRetryableError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *RetryableError
		wantMsg string
	}{
		{
			name:    "with message",
			err:     &RetryableError{Err: errors.New("original"), Message: "retry failed"},
			wantMsg: "retry failed: original",
		},
		{
			name:    "without message",
			err:     &RetryableError{Err: errors.New("original")},
			wantMsg: "original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.wantMsg) {
				t.Errorf("Expected error message containing '%s', got '%s'", tt.wantMsg, got)
			}
		})
	}
}

// TestDefaultRetryConfig はDefaultRetryConfigの初期化をテストします
func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts=3, got %d", config.MaxAttempts)
	}

	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("Expected InitialBackoff=100ms, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 2*time.Second {
		t.Errorf("Expected MaxBackoff=2s, got %v", config.MaxBackoff)
	}

	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected BackoffMultiplier=2.0, got %v", config.BackoffMultiplier)
	}

	if config.IsRetryable == nil {
		t.Error("IsRetryable should not be nil")
	}

	if config.OnRetry == nil {
		t.Error("OnRetry should not be nil")
	}
}

// TestDefaultIsRetryable はDefaultIsRetryableの判定をテストします
func TestDefaultIsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		shouldRetry bool
	}{
		{
			name:        "timeout error - retryable",
			err:         errors.New("timeout exceeded"),
			shouldRetry: true,
		},
		{
			name:        "connection refused - retryable",
			err:         errors.New("connection refused"),
			shouldRetry: true,
		},
		{
			name:        "context canceled - not retryable",
			err:         errors.New("context canceled"),
			shouldRetry: false,
		},
		{
			name:        "not found - not retryable",
			err:         errors.New("not found"),
			shouldRetry: false,
		},
		{
			name:        "generic error - not retryable",
			err:         errors.New("some error"),
			shouldRetry: false,
		},
		{
			name:        "nil error - not retryable",
			err:         nil,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultIsRetryable(tt.err)
			if result != tt.shouldRetry {
				t.Errorf("Expected shouldRetry=%v for error '%v', got %v", tt.shouldRetry, tt.err, result)
			}
		})
	}
}

// TestRetry_Success はRetry関数が成功する場合をテストします
func TestRetry_Success(t *testing.T) {
	attempt := 0
	fn := func() error {
		attempt++
		if attempt < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	ctx := context.Background()
	err := Retry(ctx, fn, nil)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempt != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempt)
	}
}

// TestRetry_MaxAttempts は最大リトライ回数をテストします
func TestRetry_MaxAttempts(t *testing.T) {
	attempt := 0
	fn := func() error {
		attempt++
		return errors.New("persistent error")
	}

	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		IsRetryable:       func(err error) bool { return true },
	}

	ctx := context.Background()
	err := Retry(ctx, fn, config)

	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if attempt != config.MaxAttempts {
		t.Errorf("Expected %d attempts, got %d", config.MaxAttempts, attempt)
	}
}

// TestRetry_NonRetryableError はリトライ不可なエラーをテストします
func TestRetry_NonRetryableError(t *testing.T) {
	attempt := 0
	fn := func() error {
		attempt++
		return errors.New("fatal error")
	}

	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		IsRetryable:       func(err error) bool { return false },
	}

	ctx := context.Background()
	err := Retry(ctx, fn, config)

	if err == nil {
		t.Error("Expected error")
	}

	if attempt != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempt)
	}
}

// TestRetry_ContextCanceled はコンテキストキャンセルをテストします
func TestRetry_ContextCanceled(t *testing.T) {
	attempt := 0
	fn := func() error {
		attempt++
		return errors.New("retryable error")
	}

	config := &RetryConfig{
		MaxAttempts:       5,
		InitialBackoff:    100 * time.Millisecond,
		IsRetryable:       func(err error) bool { return true },
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	err := Retry(ctx, fn, config)

	if err == nil {
		t.Error("Expected error due to canceled context")
	}

	if attempt != 0 {
		t.Errorf("Expected 0 attempts due to canceled context, got %d", attempt)
	}
}

// TestRetry_CallbackInvoked はコールバックが呼び出されることをテストします
func TestRetry_CallbackInvoked(t *testing.T) {
	callbackInvoked := false
	attempt := 0

	fn := func() error {
		attempt++
		if attempt == 1 {
			return errors.New("temporary error")
		}
		return nil
	}

	config := &RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		IsRetryable:    func(err error) bool { return true },
		OnRetry: func(a int, e error) {
			callbackInvoked = true
			if a != 1 {
				t.Errorf("Expected attempt=1 in callback, got %d", a)
			}
		},
	}

	ctx := context.Background()
	err := Retry(ctx, fn, config)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if !callbackInvoked {
		t.Error("Expected OnRetry callback to be invoked")
	}
}

// TestCalculateBackoff はバックオフ計算をテストします
func TestCalculateBackoff(t *testing.T) {
	config := &RetryConfig{
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first retry",
			attempt:  0,
			expected: 100 * time.Millisecond,
		},
		{
			name:     "second retry",
			attempt:  1,
			expected: 200 * time.Millisecond,
		},
		{
			name:     "third retry",
			attempt:  2,
			expected: 400 * time.Millisecond,
		},
		{
			name:     "max backoff",
			attempt:  10,
			expected: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt, config)
			if backoff != tt.expected {
				t.Errorf("Expected backoff %v for attempt %d, got %v", tt.expected, tt.attempt, backoff)
			}
		})
	}
}

// TestExponentialBackoff はExponentialBackoff関数をテストします
func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		initial    time.Duration
		max        time.Duration
		multiplier float64
		expected   time.Duration
	}{
		{
			name:       "attempt 0",
			attempt:    0,
			initial:    100 * time.Millisecond,
			max:        1 * time.Second,
			multiplier: 2.0,
			expected:   0,
		},
		{
			name:       "attempt 1",
			attempt:    1,
			initial:    100 * time.Millisecond,
			max:        1 * time.Second,
			multiplier: 2.0,
			expected:   100 * time.Millisecond,
		},
		{
			name:       "attempt 2",
			attempt:    2,
			initial:    100 * time.Millisecond,
			max:        1 * time.Second,
			multiplier: 2.0,
			expected:   200 * time.Millisecond,
		},
		{
			name:       "attempt 3 exceeds max",
			attempt:    3,
			initial:    400 * time.Millisecond,
			max:        500 * time.Millisecond,
			multiplier: 2.0,
			expected:   500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExponentialBackoff(tt.attempt, tt.initial, tt.max, tt.multiplier)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsSelectorNotFoundError はIsSelectorNotFoundError関数をテストします
func TestIsSelectorNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "selector not found",
			err:      errors.New("could not get nodes for selector"),
			expected: true,
		},
		{
			name:     "element not found",
			err:      errors.New("no elements found"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSelectorNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v for error '%v', got %v", tt.expected, tt.err, result)
			}
		})
	}
}

// TestNewTemporaryError はNewTemporaryError関数をテストします
func TestNewTemporaryError(t *testing.T) {
	originalErr := errors.New("original error")
	tempErr := NewTemporaryError(originalErr, "operation failed")

	if tempErr.Err != originalErr {
		t.Error("Original error not preserved")
	}

	if !tempErr.ShouldRetry {
		t.Error("ShouldRetry should be true")
	}

	if tempErr.Message != "operation failed" {
		t.Errorf("Expected message 'operation failed', got '%s'", tempErr.Message)
	}
}

// TestMaxRetriesExceededError はMaxRetriesExceededErrorをテストします
func TestMaxRetriesExceededError(t *testing.T) {
	lastErr := errors.New("last error")
	maxErr := &MaxRetriesExceededError{
		Attempts: 5,
		LastErr:  lastErr,
	}

	errMsg := maxErr.Error()
	if !strings.Contains(errMsg, "max retries exceeded") {
		t.Error("Error message should contain 'max retries exceeded'")
	}
	if !strings.Contains(errMsg, "5") {
		t.Error("Error message should contain attempt count")
	}
}

// TestIsMaxRetriesExceeded はIsMaxRetriesExceeded関数をテストします
func TestIsMaxRetriesExceeded(t *testing.T) {
	maxErr := &MaxRetriesExceededError{Attempts: 3}
	normalErr := errors.New("normal error")

	if !IsMaxRetriesExceeded(maxErr) {
		t.Error("IsMaxRetriesExceeded should return true for MaxRetriesExceededError")
	}

	if IsMaxRetriesExceeded(normalErr) {
		t.Error("IsMaxRetriesExceeded should return false for normal errors")
	}
}

// BenchmarkRetry はRetry関数のベンチマークです
func BenchmarkRetry(b *testing.B) {
	fn := func() error {
		return nil
	}
	config := DefaultRetryConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Retry(context.Background(), fn, config)
	}
}

// BenchmarkCalculateBackoff はcalculateBackoffのベンチマークです
func BenchmarkCalculateBackoff(b *testing.B) {
	config := DefaultRetryConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateBackoff(i%10, config)
	}
}

// BenchmarkExponentialBackoff はExponentialBackoffのベンチマークです
func BenchmarkExponentialBackoff(b *testing.B) {
	initial := 100 * time.Millisecond
	max := 2 * time.Second
	multiplier := 2.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExponentialBackoff(i%10, initial, max, multiplier)
	}
}

// ExampleDefaultIsRetryable はDefaultIsRetryableの使用例です
func ExampleDefaultIsRetryable() {
	err := errors.New("connection timeout")
	if DefaultIsRetryable(err) {
		// リトライ可能
	}
	// 出力:
}

// ExampleRetry はRetryの使用例です
func ExampleRetry() {
	fn := func() error {
		// 何らかの操作
		return nil
	}

	err := Retry(context.Background(), fn, nil)
	if err != nil {
		// エラー処理
	}
	// 出力:
}