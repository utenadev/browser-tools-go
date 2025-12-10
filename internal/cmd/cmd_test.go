package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	// Check that help output contains expected commands
	expectedCommands := []string{"start", "close", "navigate", "screenshot", "pick", "eval", "cookies", "search", "content", "hn-scraper"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("Expected help output to contain '%s', but it didn't.\nOutput: %s", cmd, output)
		}
	}
}

func TestStartCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"start", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	// Check that start help contains expected flags
	if !strings.Contains(output, "--port") {
		t.Error("Expected start help to contain '--port' flag")
	}
	if !strings.Contains(output, "--headless") {
		t.Error("Expected start help to contain '--headless' flag")
	}
}

func TestSearchCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"search", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--n") {
		t.Error("Expected search help to contain '--n' flag")
	}
	if !strings.Contains(output, "--content") {
		t.Error("Expected search help to contain '--content' flag")
	}
}

func TestScreenshotCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"screenshot", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--url") {
		t.Error("Expected screenshot help to contain '--url' flag")
	}
	if !strings.Contains(output, "--full-page") {
		t.Error("Expected screenshot help to contain '--full-page' flag")
	}
}

func TestContentCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"content", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--format") {
		t.Error("Expected content help to contain '--format' flag")
	}
}

func TestPickCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"pick", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--all") {
		t.Error("Expected pick help to contain '--all' flag")
	}
}
