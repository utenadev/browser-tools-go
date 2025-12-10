package browser

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// WaitForWS polls a WebSocket URL until it becomes available or the timeout is reached.
func WaitForWS(ctx context.Context, url string, maxWait time.Duration) error {
	addr := strings.TrimPrefix(url, "ws://")

	dialer := net.Dialer{
		Timeout: time.Second, // Timeout for each individual dial attempt
	}

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			log.Println("✅ Browser WebSocket is ready.")
			return nil
		}
		time.Sleep(100 * time.Millisecond) // Wait before retrying
	}

	return fmt.Errorf("browser websocket not ready after %v", maxWait)
}
