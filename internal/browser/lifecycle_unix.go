//go:build !windows

package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"browser-tools-go/internal/config"
)

func mustGetConfigPath() string {
	path, err := config.GetConfigPath()
	if err != nil {
		log.Fatalf("Could not determine config path: %v", err)
	}
	return path
}

// Start launches a new persistent Chrome instance.
func Start(port int, headless bool) error {
	if _, err := config.LoadWsInfo(); err == nil {
		return fmt.Errorf("browser is already running. Use 'close' to stop it first")
	}

	var chromePath string
	for _, executable := range []string{"google-chrome", "chrome", "chromium"} {
		path, err := exec.LookPath(executable)
		if err == nil {
			chromePath = path
			break
		}
	}

	if chromePath == "" {
		return fmt.Errorf("could not find Chrome installation")
	}

	userDataDir := strings.Replace(mustGetConfigPath(), "ws.json", "user-data", 1)
	chromeArgs := []string{
		fmt.Sprintf("--remote-debugging-port=%d", port),
		fmt.Sprintf("--user-data-dir=%s", userDataDir),
	}
	if headless {
		chromeArgs = append(chromeArgs, "--headless=new")
	}

	proc := exec.Command(chromePath, chromeArgs...)

	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start Chrome: %w", err)
	}

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d", port)
	log.Printf("⏳ Waiting for browser to be ready at %s...", wsURL)
	if err := WaitForWS(context.Background(), wsURL, 5*time.Second); err != nil {
		_ = proc.Process.Kill()
		return fmt.Errorf("error waiting for browser: %w", err)
	}

	if err := config.SaveWsInfo(wsURL, proc.Process.Pid); err != nil {
		_ = proc.Process.Kill()
		return fmt.Errorf("failed to save session info: %w", err)
	}

	log.Printf("✅ Browser started successfully with PID %d.", proc.Process.Pid)
	return nil
}

// Close terminates the persistent Chrome instance.
func Close() error {
	info, err := config.LoadWsInfo()
	if err != nil {
		return fmt.Errorf("browser is not running")
	}

	log.Printf("🛑 Closing browser with PID %d...", info.Pid)
	proc, err := os.FindProcess(info.Pid)
	if err != nil {
		log.Printf("⚠️ Could not find process with PID %d: %v. The process may have already exited.", info.Pid, err)
	} else {
		err = proc.Signal(syscall.SIGTERM)
		if err != nil {
			log.Printf("⚠️ Failed to terminate process: %v. Attempting cleanup anyway.", err)
		}
	}

	if err := config.RemoveWsInfo(); err != nil {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	log.Println("✅ Browser session closed and cleaned up.")
	return nil
}
