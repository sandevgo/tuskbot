package conv

import (
	"strings"
	"testing"
)

func TestMarkdownToTelegramHTML(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   string
		skipExact  bool
		assertFunc func(t *testing.T, got string)
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text",
			input:    "Hello world",
			expected: "Hello world\n",
		},
		{
			name:     "bold text",
			input:    "**bold**",
			expected: "<strong>bold</strong>\n",
		},
		{
			name:     "italic text",
			input:    "*italic*",
			expected: "<em>italic</em>\n",
		},
		{
			name:     "bold and italic",
			input:    "***bold italic***",
			expected: "<strong><em>bold italic</em></strong>\n",
		},
		{
			name:     "raw HTML underline preserved",
			input:    "<u>underline</u>",
			expected: "<u>underline</u>\n",
		},
		{
			name:     "double underscore is bold (standard markdown)",
			input:    "__bold__",
			expected: "<strong>bold</strong>\n",
		},
		{
			name:     "strikethrough",
			input:    "~~strikethrough~~",
			expected: "<del>strikethrough</del>\n",
		},
		{
			name:     "inline code",
			input:    "`code`",
			expected: "<code>code</code>\n",
		},
		{
			name:     "code block",
			input:    "```\ncode block\n```",
			expected: "<pre><code>code block\n</code></pre>\n",
		},
		{
			name:     "single blank line between paragraphs",
			input:    "First paragraph.\n\nSecond paragraph.",
			expected: "First paragraph.\n\nSecond paragraph.\n",
		},
		{
			name:     "multiple blank lines collapse to single paragraph break",
			input:    "First paragraph.\n\n\n\nSecond paragraph.",
			expected: "First paragraph.\n\nSecond paragraph.\n",
		},
		{
			name:      "markdown table keeps rows on separate lines",
			input:     "| Name | Value |\n|---|---|\n| A | 1 |\n| B | 2 |",
			skipExact: true,
			assertFunc: func(t *testing.T, got string) {
				t.Helper()
				if strings.Count(got, "\n\n") > 0 {
					t.Fatalf("expected no double newlines in table output, got %q", got)
				}
				for _, mustContain := range []string{"Name", "Value", "A", "1", "B", "2"} {
					if !strings.Contains(got, mustContain) {
						t.Fatalf("expected table output to contain %q, got %q", mustContain, got)
					}
				}
			},
		},
		{
			name:      "table with empty cells does not create blank lines",
			input:     "| Name | Value |\n|---|---|\n| A |  |\n|  | 2 |",
			skipExact: true,
			assertFunc: func(t *testing.T, got string) {
				t.Helper()
				if strings.Count(got, "\n\n") > 0 {
					t.Fatalf("expected no double newlines in table output, got %q", got)
				}
				for _, mustContain := range []string{"Name", "Value", "A", "2"} {
					if !strings.Contains(got, mustContain) {
						t.Fatalf("expected table output to contain %q, got %q", mustContain, got)
					}
				}
			},
		},
		{
			name:     "code block with language",
			input:    "```go\nfunc main() {}\n```",
			expected: "<pre><code class=\"language-go\">func main() {}\n</code></pre>\n",
		},
		{
			name:     "blockquote",
			input:    "> quote",
			expected: "<blockquote>\nquote\n</blockquote>\n",
		},
		{
			name:     "link",
			input:    "[link](https://example.com)",
			expected: "<a href=\"https://example.com\">link</a>\n",
		},
		{
			name:     "header tags stripped",
			input:    "# Info",
			expected: "Info\n",
		},
		{
			name:     "script tags sanitized",
			input:    "<script>alert('xss')</script>",
			expected: "\n",
		},
		{
			name:     "mixed formatting",
			input:    "**Bold** and *italic* with `code`",
			expected: "<strong>Bold</strong> and <em>italic</em> with <code>code</code>\n",
		},
		{
			name:     "link with target blank stripped",
			input:    "[link](https://example.com)",
			expected: "<a href=\"https://example.com\">link</a>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarkdownToTelegramHTML([]byte(tt.input))
			if !tt.skipExact && got != tt.expected {
				t.Errorf("MarkdownToTelegramHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
			if tt.assertFunc != nil {
				tt.assertFunc(t, got)
			}
		})
	}
}
