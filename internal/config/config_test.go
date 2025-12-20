package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetConfigPath_Success は設定ファイルパスの取得をテストします。
func TestGetConfigPath_Success(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	expected := filepath.Join(home, ".browser-tools-go", "ws.json")
	if path != expected {
		t.Errorf("Expected config path %s, got %s", expected, path)
	}
}

// TestSaveAndLoadWsInfo はWsInfoの保存と読み込みをテストします。
func TestSaveAndLoadWsInfo(t *testing.T) {
	// テスト用の一時ディレクトリ作成
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	// HOME環境変数を一時ディレクトリに設定
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// テスト用のWsInfo
	testURL := "ws://127.0.0.1:9222"
	testPID := 12345

	// 保存
	err := SaveWsInfo(testURL, testPID)
	if err != nil {
		t.Fatalf("Failed to save WsInfo: %v", err)
	}

	// 読み込み
	info, err := LoadWsInfo()
	if err != nil {
		t.Fatalf("Failed to load WsInfo: %v", err)
	}

	// 検証
	if info.Url != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, info.Url)
	}

	if info.Pid != testPID {
		t.Errorf("Expected PID %d, got %d", testPID, info.Pid)
	}

	// クリーンアップ
	_ = RemoveWsInfo()
}

// TestLoadWsInfo_NotExist は存在しないファイルの読み込みをテストします。
func TestLoadWsInfo_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	_, err := LoadWsInfo()
	if err == nil {
		t.Error("Expected error for non-existent config file, got nil")
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist error, got %v", err)
	}
}

// TestSaveWsInfo_CreateDirectory はディレクトリ作成をテストします。
func TestSaveWsInfo_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// ディレクトリは存在しないはず
	testURL := "ws://127.0.0.1:9222"
	testPID := 12345

	err := SaveWsInfo(testURL, testPID)
	if err != nil {
		t.Fatalf("Failed to save WsInfo with directory creation: %v", err)
	}

	// ディレクトリが作成されたか確認
	configPath, _ := GetConfigPath()
	configDir := filepath.Dir(configPath)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %s", configDir)
	}

	// クリーンアップ
	_ = RemoveWsInfo()
}

// TestRemoveWsInfo_Success はWsInfoの削除をテストします。
func TestRemoveWsInfo_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// まずファイルを作成
	testURL := "ws://127.0.0.1:9222"
	testPID := 12345

	err := SaveWsInfo(testURL, testPID)
	if err != nil {
		t.Fatalf("Failed to save WsInfo: %v", err)
	}

	// 削除
	err = RemoveWsInfo()
	if err != nil {
		t.Fatalf("Failed to remove WsInfo: %v", err)
	}

	// ファイルが存在しないことを確認
	_, err = LoadWsInfo()
	if err == nil {
		t.Error("Expected error after removing config file, got nil")
	}
}

// TestRemoveWsInfo_NotExist は存在しないファイルの削除をテストします。
func TestRemoveWsInfo_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// 存在しないファイルの削除はエラーを返さない
	err := RemoveWsInfo()
	if err != nil {
		t.Errorf("Expected no error for removing non-existent file, got %v", err)
	}
}

// TestWsInfo_JSONSerialization はJSONシリアライゼーションをテストします。
func TestWsInfo_JSONSerialization(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// 特殊文字を含むURL
	testURL := "ws://127.0.0.1:9222/devtools/browser/123e4567-e89b-12d3-a456-426614174000"
	testPID := 12345

	err := SaveWsInfo(testURL, testPID)
	if err != nil {
		t.Fatalf("Failed to save WsInfo with complex URL: %v", err)
	}

	info, err := LoadWsInfo()
	if err != nil {
		t.Fatalf("Failed to load WsInfo: %v", err)
	}

	if info.Url != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, info.Url)
	}

	if info.Pid != testPID {
		t.Errorf("Expected PID %d, got %d", testPID, info.Pid)
	}

	// クリーンアップ
	_ = RemoveWsInfo()
}

// TestSaveWsInfo_FilePermissions はファイルパーミッションをテストします。
func TestSaveWsInfo_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testURL := "ws://127.0.0.1:9222"
	testPID := 12345

	err := SaveWsInfo(testURL, testPID)
	if err != nil {
		t.Fatalf("Failed to save WsInfo: %v", err)
	}

	// ファイルパーミッションの確認
	configPath, _ := GetConfigPath()
	stat, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// 0600 (owner read/write only) であることを確認
	expectedPerm := os.FileMode(0600)
	if stat.Mode().Perm() != expectedPerm {
		t.Errorf("Expected file permissions %#o, got %#o", expectedPerm, stat.Mode().Perm())
	}

	// クリーンアップ
	_ = RemoveWsInfo()
}

// TestGetConfigPath_MultipleCalls は複数回のGetConfigPath呼び出しで一貫性があることをテストします。
func TestGetConfigPath_MultipleCalls(t *testing.T) {
	path1, err1 := GetConfigPath()
	if err1 != nil {
		t.Fatalf("First call failed: %v", err1)
	}

	path2, err2 := GetConfigPath()
	if err2 != nil {
		t.Fatalf("Second call failed: %v", err2)
	}

	if path1 != path2 {
		t.Errorf("Config paths should be consistent: %s vs %s", path1, path2)
	}
}

// BenchmarkSaveWsInfo はSaveWsInfoのベンチマークテストです。
func BenchmarkSaveWsInfo(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testURL := "ws://127.0.0.1:9222"
	testPID := 12345

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SaveWsInfo(testURL, testPID)
	}
}

// BenchmarkLoadWsInfo はLoadWsInfoのベンチマークテストです。
func BenchmarkLoadWsInfo(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// セットアップ
	testURL := "ws://127.0.0.1:9222"
	testPID := 12345
	_ = SaveWsInfo(testURL, testPID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LoadWsInfo()
	}
}