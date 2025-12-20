package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ファイルパストラバーサル対策用エラー
var (
	ErrInvalidPath        = errors.New("invalid file path")
	ErrPathTraversal      = errors.New("path traversal detected")
	ErrOutsideWorkingDir  = errors.New("path outside working directory")
	ErrEmptyPath          = errors.New("empty path not allowed")
)

// ValidateFilePath はファイルパスの安全性を検証します。
// 以下のチェックを行います：
// 1. 空パスの拒否
// 2. 絶対パスの拒否（セキュリティポリシーに応じて）
// 3. 親ディレクトリ参照（../）の検出
// 4. 作業ディレクトリ外へのアクセス防止
// 5. NULLバイトの検出（古いシステム対策）
// 6. 危険な文字列パターンの検出
//
// 引数：
//   path: 検証するファイルパス
//   allowAbsolute: 絶対パスを許可するかどうか
//   baseDir: ベースディレクトリ（指定された場合、このディレクトリ外へのアクセスを禁止）
//
// 戻り値：
//   絶対パスとエラー（検証に失敗した場合）
func ValidateFilePath(path string, allowAbsolute bool, baseDir string) (string, error) {
	// 1. 空パスの拒否
	if len(path) == 0 {
		return "", ErrEmptyPath
	}

	// 2. NULLバイト検出
	for i := 0; i < len(path); i++ {
		if path[i] == 0 {
			return "", ErrInvalidPath
		}
	}

	// 3. 絶対パスの検出
	if filepath.IsAbs(path) {
		if !allowAbsolute {
			return "", ErrPathTraversal
		}
	}

	// 4. クリーンなパスに変換
	cleanPath := filepath.Clean(path)

	// 5. 親ディレクトリ参照の検出
	if hasPathTraversal(cleanPath) {
		return "", ErrPathTraversal
	}

	// 6. 作業ディレクトリ外へのアクセス防止
	if baseDir != "" {
		absBase, err := filepath.Abs(baseDir)
		if err != nil {
			return "", ErrInvalidPath
		}

		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", ErrInvalidPath
		}

		relPath, err := filepath.Rel(absBase, absPath)
		if err != nil {
			return "", ErrOutsideWorkingDir
		}

		if hasPathTraversal(relPath) {
			return "", ErrOutsideWorkingDir
		}
	}

	// 7. 危険な文字列パターンの検出
	dangerousPatterns := []string{"~", ".."}
	for _, pattern := range dangerousPatterns {
		if containsPathSegment(cleanPath, pattern) {
			return "", ErrPathTraversal
		}
	}

	return cleanPath, nil
}

// hasPathTraversal はパスに親ディレクトリ参照が含まれているかチェックします。
func hasPathTraversal(path string) bool {
	return filepath.IsAbs(path) || strings.HasPrefix(path, "..")
}

// containsPathSegment はパスに危険なパスセグメントが含まれているかチェックします。
func containsPathSegment(path, segment string) bool {
	separator := string(os.PathSeparator)
	return strings.HasPrefix(path, segment+separator) || strings.HasSuffix(path, separator+segment) || strings.HasPrefix(path, segment)
}

// SecureWriteFile はファイルパスを検証してからファイルに書き込みます。
func SecureWriteFile(filename string, data []byte, perm os.FileMode, baseDir string) error {
	// ファイルパスを検証
	validatedPath, err := ValidateFilePath(filename, false, baseDir)
	if err != nil {
		return err
	}

	// ディレクトリの作成
	if err := os.MkdirAll(filepath.Dir(validatedPath), 0755); err != nil {
		return err
	}

	// ファイルの書き込み
	if err := os.WriteFile(validatedPath, data, perm); err != nil {
		return err
	}

	return nil
}

// ValidateScreenshotPath はスクリーンショット保存用のファイルパスを検証します。
// デフォルトではカレントディレクトリ（または指定されたベースディレクトリ）に保存することを保証します。
func ValidateScreenshotPath(path string, baseDir string) (string, error) {
	// 空文字列の場合はデフォルトファイル名を返す
	if path == "" {
		return "screenshot.png", nil
	}

	// 拡張子チェック（PNG形式を期待）
	ext := filepath.Ext(path)
	if ext == "" {
		path += ".png"
	} else if ext != ".png" {
		// PNG以外の場合は、ログを出力してPNGに変更
		// 必要に応じてエラーとして扱うことも可能
		path = path[:len(path)-len(ext)] + ".png"
	}

	// ファイルパスの検証
	return ValidateFilePath(path, false, baseDir)
}

// ValidateFilePathStrict はより厳格な検証を行います。
// 絶対パスを許可せず、常にカレントディレクトリ内でのみ操作を許可します。
func ValidateFilePathStrict(path string) (string, error) {
	return ValidateFilePath(path, false, ".")
}

// ValidateFilePathLenient はより寛容な検証を行います。
// 絶対パスを許可しますが、パストラバーサル攻撃は防ぎます。
func ValidateFilePathLenient(path string) (string, error) {
	return ValidateFilePath(path, true, "")
}

// GetSafeAbsolutePath はパスを安全な絶対パスに変換します。
func GetSafeAbsolutePath(path string, baseDir string) (string, error) {
	validatedPath, err := ValidateFilePath(path, true, baseDir)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(validatedPath)
	if err != nil {
		return "", ErrInvalidPath
	}

	return absPath, nil
}