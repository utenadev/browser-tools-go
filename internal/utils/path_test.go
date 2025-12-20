package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateFilePath_EmptyPath は空パスを拒否することをテストします。
func TestValidateFilePath_EmptyPath(t *testing.T) {
	_, err := ValidateFilePath("", false, ".")
	if err != ErrEmptyPath {
		t.Errorf("Expected ErrEmptyPath, got %v", err)
	}
}

// TestValidateFilePath_NullByte はNULLバイトを含むパスを拒否することをテストします。
func TestValidateFilePath_NullByte(t *testing.T) {
	_, err := ValidateFilePath("file\x00name.txt", false, ".")
	if err != ErrInvalidPath {
		t.Errorf("Expected ErrInvalidPath, got %v", err)
	}
}

// TestValidateFilePath_AbsolutePath は絶対パスを拒否することをテストします（allowAbsolute=falseの場合）。
func TestValidateFilePath_AbsolutePathNotAllowed(t *testing.T) {
	_, err := ValidateFilePath("/tmp/test.txt", false, ".")
	if err != ErrPathTraversal {
		t.Errorf("Expected ErrPathTraversal, got %v", err)
	}
}

// TestValidateFilePath_AbsolutePathAllowed は絶対パスを許可することをテストします（allowAbsolute=trueの場合）。
func TestValidateFilePath_AbsolutePathAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")

	result, err := ValidateFilePath(path, true, "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Clean(path)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestValidateFilePath_ParentDirectory は親ディレクトリ参照（../）を検出することをテストします。
func TestValidateFilePath_ParentDirectory(t *testing.T) {
	testCases := []string{
		"../test.txt",
		"subdir/../../test.txt",
		"../../../etc/passwd",
		"~/.ssh/id_rsa",
	}

	for _, path := range testCases {
		_, err := ValidateFilePath(path, false, ".")
		if err != ErrPathTraversal && err != ErrOutsideWorkingDir {
			t.Errorf("Expected path traversal error for %s, got %v", path, err)
		}
	}
}

// TestValidateFilePath_ValidRelativePath は安全な相対パスを受け入れることをテストします。
func TestValidateFilePath_ValidRelativePath(t *testing.T) {
	testCases := []struct {
		path    string
		baseDir string
	}{
		{"test.txt", "."},
		{"subdir/test.txt", "."},
		{"../test.txt", ".."},
		{"dir/subdir/file.txt", "."},
	}

	for _, tc := range testCases {
		result, err := ValidateFilePath(tc.path, false, tc.baseDir)
		if err != nil {
			t.Errorf("Expected no error for %s with baseDir %s, got %v", tc.path, tc.baseDir, err)
		}

		expected := filepath.Clean(tc.path)
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	}
}

// TestValidateFilePath_BaseDirSecurity はベースディレクトリ外へのアクセスを防ぐことをテストします。
func TestValidateFilePath_BaseDirSecurity(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "safe")

	// ベースディレクトリ外へのパス
	_, err := ValidateFilePath("../escape.txt", false, baseDir)
	if err != ErrOutsideWorkingDir {
		t.Errorf("Expected ErrOutsideWorkingDir, got %v", err)
	}

	// ベースディレクトリ内へのパス
	result, err := ValidateFilePath("safe.txt", false, baseDir)
	if err != nil {
		t.Errorf("Expected no error for safe path, got %v", err)
	}
	if result != "safe.txt" {
		t.Errorf("Expected 'safe.txt', got %s", result)
	}
}

// TestValidateScreenshotPath_EmptyPath は空パスの場合にデフォルトファイル名を返すことをテストします。
func TestValidateScreenshotPath_EmptyPath(t *testing.T) {
	result, err := ValidateScreenshotPath("", ".")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "screenshot.png" {
		t.Errorf("Expected 'screenshot.png', got %s", result)
	}
}

// TestValidateScreenshotPath_ExtensionAdded は拡張子がない場合にPNGを追加することをテストします。
func TestValidateScreenshotPath_ExtensionAdded(t *testing.T) {
	result, err := ValidateScreenshotPath("myfile", ".")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "myfile.png" {
		t.Errorf("Expected 'myfile.png', got %s", result)
	}
}

// TestValidateScreenshotPath_ExtensionChanged はPNG以外の拡張子をPNGに変更することをテストします。
func TestValidateScreenshotPath_ExtensionChanged(t *testing.T) {
	result, err := ValidateScreenshotPath("myfile.jpg", ".")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "myfile.png" {
		t.Errorf("Expected 'myfile.png', got %s", result)
	}
}

// TestValidateScreenshotPath_ValidPath は有効なPNGパスを受け入れることをテストします。
func TestValidateScreenshotPath_ValidPath(t *testing.T) {
	result, err := ValidateScreenshotPath("screenshots/capture.png", ".")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Clean("screenshots/capture.png")
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestValidateScreenshotPath_PathTraversal はスクリーンショットパスでのパストラバーサルを防ぐことをテストします。
func TestValidateScreenshotPath_PathTraversal(t *testing.T) {
	_, err := ValidateScreenshotPath("../secrets.txt", ".")
	if err != ErrPathTraversal {
		t.Errorf("Expected ErrPathTraversal, got %v", err)
	}
}

// TestSecureWriteFile_Success はSecureWriteFileが正常に動作することをテストします。
func TestSecureWriteFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := "testfile.txt"
	data := []byte("test data")

	err := SecureWriteFile(filePath, data, 0644, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// ファイルが実際に書き込まれたか確認
	fullPath := filepath.Join(tmpDir, filePath)
	readData, err := os.ReadFile(fullPath)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}

	if string(readData) != string(data) {
		t.Errorf("Expected %s, got %s", string(data), string(readData))
	}
}

// TestSecureWriteFile_PathTraversal はSecureWriteFileがパストラバーサルを防ぐことをテストします。
func TestSecureWriteFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := "../escape.txt"
	data := []byte("malicious data")

	err := SecureWriteFile(filePath, data, 0644, tmpDir)
	if err != ErrOutsideWorkingDir {
		t.Errorf("Expected ErrOutsideWorkingDir, got %v", err)
	}
}

// TestValidateFilePathStrict はValidateFilePathStrictの動作をテストします。
func TestValidateFilePathStrict(t *testing.T) {
	// 相対パスは許可
	result, err := ValidateFilePathStrict("test.txt")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "test.txt" {
		t.Errorf("Expected 'test.txt', got %s", result)
	}

	// 絶対パスは拒否
	_, err = ValidateFilePathStrict("/tmp/test.txt")
	if err != ErrPathTraversal {
		t.Errorf("Expected ErrPathTraversal, got %v", err)
	}
}

// TestValidateFilePathLenient はValidateFilePathLenientの動作をテストします。
func TestValidateFilePathLenient(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")

	// 絶対パスは許可
	result, err := ValidateFilePathLenient(path)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Clean(path)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// パストラバーサルは拒否
	_, err = ValidateFilePathLenient("../etc/passwd")
	if err != ErrPathTraversal {
		t.Errorf("Expected ErrPathTraversal, got %v", err)
	}
}

// TestGetSafeAbsolutePath_Success はGetSafeAbsolutePathが正常に動作することをテストします。
func TestGetSafeAbsolutePath_Success(t *testing.T) {
	tmpDir := t.TempDir()
	relPath := "test.txt"

	absPath, err := GetSafeAbsolutePath(relPath, tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := filepath.Join(tmpDir, relPath)
	if absPath != expected {
		t.Errorf("Expected %s, got %s", expected, absPath)
	}
}

// TestGetSafeAbsolutePath_PathTraversal はGetSafeAbsolutePathがパストラバーサルを防ぐことをテストします。
func TestGetSafeAbsolutePath_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	relPath := "../escape.txt"

	_, err := GetSafeAbsolutePath(relPath, tmpDir)
	if err != ErrOutsideWorkingDir {
		t.Errorf("Expected ErrOutsideWorkingDir, got %v", err)
	}
}

// BenchmarkValidateFilePath はValidateFilePathのベンチマークテストです。
func BenchmarkValidateFilePath(b *testing.B) {
	baseDir := "."
	path := "test/safe/path.txt"

	for i := 0; i < b.N; i++ {
		_, _ = ValidateFilePath(path, false, baseDir)
	}
}

// ExampleValidateFilePath はパストラバーサル検出の例です。
func ExampleValidateFilePath_detectTraversal() {
	// 親ディレクトリ参照を検出
	_, err := ValidateFilePath("../etc/passwd", false, ".")
	if err != nil {
		// パストラバーサル攻撃を検出
	}
	// 出力:
}

// ExampleValidateScreenshotPath_safeScreenshot は安全なスクリーンショットパスの例です。
func ExampleValidateScreenshotPath_safeScreenshot() {
	// 安全なスクリーンショット保存
	path, err := ValidateScreenshotPath("my_screenshot", ".")
	if err != nil {
		// エラー処理
	}
	_ = path
	// 出力:
}