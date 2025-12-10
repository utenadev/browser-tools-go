package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	// Verify path ends with expected filename
	if filepath.Base(path) != "ws.json" {
		t.Errorf("Expected path to end with 'ws.json', got: %s", filepath.Base(path))
	}

	// Verify path contains .browser-tools-go directory
	dir := filepath.Dir(path)
	if filepath.Base(dir) != ".browser-tools-go" {
		t.Errorf("Expected path to be in '.browser-tools-go' directory, got: %s", dir)
	}
}

func TestWsInfoMarshalUnmarshal(t *testing.T) {
	// Test WsInfo struct serialization
	testURL := "ws://127.0.0.1:9222"
	testPID := 12345

	info := WsInfo{Url: testURL, Pid: testPID}

	// Marshal
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal info: %v", err)
	}

	// Unmarshal
	var loadedInfo WsInfo
	if err := json.Unmarshal(data, &loadedInfo); err != nil {
		t.Fatalf("Failed to unmarshal info: %v", err)
	}

	if loadedInfo.Url != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, loadedInfo.Url)
	}
	if loadedInfo.Pid != testPID {
		t.Errorf("Expected PID %d, got %d", testPID, loadedInfo.Pid)
	}
}

func TestSaveAndLoadWsInfo_Integration(t *testing.T) {
	// Skip if we can't get config path (e.g., no home directory)
	configPath, err := GetConfigPath()
	if err != nil {
		t.Skipf("Cannot get config path: %v", err)
	}

	// Ensure the directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Clean up any existing file first
	os.Remove(configPath)

	testURL := "ws://127.0.0.1:9999"
	testPID := 54321

	// Save
	if err := SaveWsInfo(testURL, testPID); err != nil {
		t.Fatalf("SaveWsInfo failed: %v", err)
	}

	// Load
	loadedInfo, err := LoadWsInfo()
	if err != nil {
		t.Fatalf("LoadWsInfo failed: %v", err)
	}

	if loadedInfo.Url != testURL {
		t.Errorf("Expected URL %s, got %s", testURL, loadedInfo.Url)
	}
	if loadedInfo.Pid != testPID {
		t.Errorf("Expected PID %d, got %d", testPID, loadedInfo.Pid)
	}

	// Clean up
	if err := RemoveWsInfo(); err != nil {
		t.Errorf("RemoveWsInfo failed: %v", err)
	}

	// Verify file is removed
	_, err = LoadWsInfo()
	if err == nil {
		t.Error("Expected error loading removed config, but got none")
	}
}

func TestRemoveWsInfo_NonExistentFile(t *testing.T) {
	// First ensure file doesn't exist
	configPath, err := GetConfigPath()
	if err != nil {
		t.Skipf("Cannot get config path: %v", err)
	}

	// Remove if exists
	os.Remove(configPath)

	// RemoveWsInfo should not fail even if file doesn't exist
	err = RemoveWsInfo()
	if err != nil {
		t.Errorf("RemoveWsInfo failed on non-existent file: %v", err)
	}
}
