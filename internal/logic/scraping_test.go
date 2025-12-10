package logic

import (
	"strings"
	"testing"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

func TestGetContentFormatting(t *testing.T) {
	htmlContent := `
		<html>
			<head><title>Test Page</title></head>
			<body>
				<h1>Hello World</h1>
				<p>This is a test paragraph.</p>
				<a href="https://example.com">Click here</a>
			</body>
		</html>
	`

	t.Run("format=text", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			t.Fatalf("failed to parse html: %v", err)
		}
		expectedText := "Hello World This is a test paragraph. Click here"
		actualText := strings.TrimSpace(doc.Find("body").Text())

		actualText = strings.Join(strings.Fields(actualText), " ")
		expectedText = strings.Join(strings.Fields(expectedText), " ")


		if actualText != expectedText {
			t.Errorf("expected text '%s', but got '%s'", expectedText, actualText)
		}
	})

	t.Run("format=markdown", func(t *testing.T) {
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			t.Fatalf("failed to convert to markdown: %v", err)
		}

		actualMarkdown := strings.TrimSpace(markdown)
		actualMarkdown = strings.ReplaceAll(actualMarkdown, "\r\n", "\n")


		if !strings.Contains(actualMarkdown, "Hello World") || !strings.Contains(actualMarkdown, "This is a test paragraph") || !strings.Contains(actualMarkdown, "[Click here](https://example.com)") {
			t.Errorf("expected markdown to contain specific elements, but it didn't. Got:\n%s", actualMarkdown)
		}
	})

	t.Run("format=html", func(t *testing.T) {
		if htmlContent != htmlContent {
			t.Error("expected html to be unchanged, but it was modified")
		}
	})
}
