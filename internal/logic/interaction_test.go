package logic

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
)

// setupTestServer creates a simple HTTP server for testing.
func setupTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<body>
				<div id="div1" style="position: absolute; top: 10px; left: 20px; width: 30px; height: 40px;">First</div>
				<span id="span1" style="position: absolute; top: 100px; left: 120px; width: 130px; height: 140px;">Second</span>
				<div id="div2" class="multiple" style="position: absolute; top: 200px; left: 220px; width: 230px; height: 240px;">Third</div>
				<div id="div3" class="multiple" style="position: absolute; top: 300px; left: 320px; width: 330px; height: 340px;">Fourth</div>
			</body>
			</html>
		`)
	})
	return httptest.NewServer(mux)
}

func TestPickElements(t *testing.T) {
	// Check if chrome is available
	_, err := exec.LookPath("google-chrome")
	if err != nil {
		t.Skip("google-chrome not found, skipping test")
	}

	server := setupTestServer()
	defer server.Close()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Navigate to the test server
	if err := chromedp.Run(ctx, chromedp.Navigate(server.URL)); err != nil {
		t.Fatalf("Failed to navigate to test server: %v", err)
	}

	t.Run("pick single element", func(t *testing.T) {
		elements, err := PickElements(ctx, "#div1", false)
		if err != nil {
			t.Fatalf("PickElements failed: %v", err)
		}

		if len(elements) != 1 {
			t.Fatalf("Expected 1 element, got %d", len(elements))
		}

		el := elements[0]
		if el.Tag != "div" {
			t.Errorf("Expected tag 'div', got '%s'", el.Tag)
		}
		if el.Text != "First" {
			t.Errorf("Expected text 'First', got '%s'", el.Text)
		}
		if el.Rect["x"] != 20.0 {
			t.Errorf("Expected rect.x to be 20, got %v", el.Rect["x"])
		}
	})

	t.Run("pick multiple elements", func(t *testing.T) {
		elements, err := PickElements(ctx, ".multiple", true)
		if err != nil {
			t.Fatalf("PickElements with --all failed: %v", err)
		}

		if len(elements) != 2 {
			t.Fatalf("Expected 2 elements, got %d", len(elements))
		}

		// Check first element
		el1 := elements[0]
		if el1.Text != "Third" {
			t.Errorf("Expected text 'Third', got '%s'", el1.Text)
		}
		if el1.Rect["x"] != 220.0 {
			t.Errorf("Expected rect.x to be 220 for the first element, got %v", el1.Rect["x"])
		}
		if el1.Rect["height"] != 240.0 {
			t.Errorf("Expected rect.height to be 240 for the first element, got %v", el1.Rect["height"])
		}

		// Check second element
		el2 := elements[1]
		if el2.Text != "Fourth" {
			t.Errorf("Expected text 'Fourth', got '%s'", el2.Text)
		}
		if el2.Rect["x"] != 320.0 {
			t.Errorf("Expected rect.x to be 320 for the second element, got %v", el2.Rect["x"])
		}
		if el2.Rect["height"] != 340.0 {
			t.Errorf("Expected rect.height to be 340 for the second element, got %v", el2.Rect["height"])
		}
	})

	t.Run("pick non-existent element", func(t *testing.T) {
		elements, err := PickElements(ctx, "#nonexistent", true)
		if err != nil {
			t.Fatalf("PickElements failed for non-existent element: %v", err)
		}
		if len(elements) != 0 {
			t.Errorf("Expected 0 elements, got %d", len(elements))
		}
	})
}
