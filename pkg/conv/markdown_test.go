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
			name:     "2-column table renders as card code block",
			input:    "| Name | Value |\n|---|---|\n| A | 1 |\n| B | 2 |",
			expected: "<pre><code>1)\nName: A\nValue: 1\n\n2)\nName: B\nValue: 2\n</code></pre>\n",
			assertFunc: func(t *testing.T, got string) {
				t.Helper()
				if strings.Contains(got, "\n\n\n") {
					t.Fatalf("expected no double blank lines in table output, got %q", got)
				}
			},
		},
		{
			name:     "3-column table renders as card code block",
			input:    "| Name | Value | Note |\n|---|---|---|\n| A | 1 | ok |\n| B | 2 | done |",
			expected: "<pre><code>1)\nName: A\nValue: 1\nNote: ok\n\n2)\nName: B\nValue: 2\nNote: done\n</code></pre>\n",
			assertFunc: func(t *testing.T, got string) {
				t.Helper()
				if strings.Contains(got, "\n\n\n") {
					t.Fatalf("expected no double blank lines in table output, got %q", got)
				}
			},
		},
		{
			name:     "table rows with missing and extra cells map robustly",
			input:    "| Name | Value | Note |\n|---|---|---|\n| A | 1 |\n| B | 2 | done | extra |",
			expected: "<pre><code>1)\nName: A\nValue: 1\nNote:\n\n2)\nName: B\nValue: 2\nNote: done | extra\n</code></pre>\n",
		},
		{
			name:     "unicode dash delimiter table renders as card code block",
			input:    "| Сейчас | После |\n|——–|——-|\n| Bash | bash |\n| Read | read |",
			expected: "<pre><code>1)\nСейчас: Bash\nПосле: bash\n\n2)\nСейчас: Read\nПосле: read\n</code></pre>\n",
		},
		{
			name:     "table without outer pipes renders as card code block",
			input:    "Name | Value\n--- | ---\nA | 1",
			expected: "<pre><code>1)\nName: A\nValue: 1\n</code></pre>\n",
			assertFunc: func(t *testing.T, got string) {
				t.Helper()
				if strings.Count(got, "\n\n") > 0 {
					t.Fatalf("expected no double newlines in single-row table output, got %q", got)
				}
			},
		},
		{
			name:     "table-looking text inside fenced block stays regular code block",
			input:    "```\n| A | B |\n|---|---|\n| x | y |\n```",
			expected: "<pre><code>| A | B |\n|---|---|\n| x | y |\n</code></pre>\n",
		},
		{
			name:     "pipe text without delimiter is not table",
			input:    "Use a|b notation, not a table.",
			expected: "Use a|b notation, not a table.\n",
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
