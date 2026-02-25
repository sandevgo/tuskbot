package conv

import (
	"testing"
)

func TestMarkdownToTelegramHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
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
			if got != tt.expected {
				t.Errorf("MarkdownToTelegramHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
