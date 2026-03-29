package telegram

import (
	"io"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestSplitHTML(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLen    int
		minChunks int
		assert    func(t *testing.T, chunks []string)
	}{
		{
			name:      "no split needed",
			input:     "<strong>hello</strong>",
			maxLen:    64,
			minChunks: 1,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				if len(chunks) != 1 {
					t.Fatalf("expected 1 chunk, got %d", len(chunks))
				}
				if chunks[0] != "<strong>hello</strong>" {
					t.Fatalf("unexpected chunk: %q", chunks[0])
				}
			},
		},
		{
			name:      "split through code content",
			input:     "<code>" + strings.Repeat("0123456789", 20) + "</code>",
			maxLen:    80,
			minChunks: 2,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				assertWrappedChunks(t, chunks, "<code>", "</code>")
			},
		},
		{
			name:      "nested pre code across boundaries",
			input:     "<pre><code class=\"language-go\">" + strings.Repeat("fmt.Println(1)\n", 20) + "</code></pre>",
			maxLen:    120,
			minChunks: 2,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				assertWrappedChunks(t, chunks, "<pre><code class=\"language-go\">", "</code></pre>")
			},
		},
		{
			name:      "nested strong em across boundaries",
			input:     "<strong><em>" + strings.Repeat("nested text ", 25) + "</em></strong>",
			maxLen:    90,
			minChunks: 2,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				assertWrappedChunks(t, chunks, "<strong><em>", "</em></strong>")
			},
		},
		{
			name:      "entity safety",
			input:     "<code>" + strings.Repeat("&lt;tag&gt; &amp; value ", 20) + "</code>",
			maxLen:    85,
			minChunks: 2,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				assertWrappedChunks(t, chunks, "<code>", "</code>")

				foundLT := false
				foundAmp := false
				for _, chunk := range chunks {
					if strings.Contains(chunk, "&lt;tag&gt;") {
						foundLT = true
					}
					if strings.Contains(chunk, "&amp;") {
						foundAmp = true
					}
				}
				if !foundLT || !foundAmp {
					t.Fatalf("expected entities to be preserved across chunks, chunks=%q", chunks)
				}
			},
		},
		{
			name:      "long unbroken formatted text",
			input:     "<strong>" + strings.Repeat("x", 350) + "</strong>",
			maxLen:    70,
			minChunks: 4,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				assertWrappedChunks(t, chunks, "<strong>", "</strong>")
			},
		},
		{
			name:      "newline preferred splitting",
			input:     "<strong>line1\nline2 line3 line4 line5 line6</strong>",
			maxLen:    36,
			minChunks: 2,
			assert: func(t *testing.T, chunks []string) {
				t.Helper()
				if !strings.Contains(chunks[0], "line1\n") {
					t.Fatalf("first chunk does not include newline breakpoint: %q", chunks[0])
				}
				if strings.Contains(chunks[0], "line2") {
					t.Fatalf("expected newline-priority split before line2, got: %q", chunks[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := splitHTML(tt.input, tt.maxLen)

			if len(chunks) < tt.minChunks {
				t.Fatalf("expected at least %d chunks, got %d", tt.minChunks, len(chunks))
			}

			for i, chunk := range chunks {
				if chunk == "" {
					t.Fatalf("chunk %d is empty", i)
				}
				if len(chunk) > tt.maxLen {
					t.Fatalf("chunk %d length %d exceeds maxLen %d: %q", i, len(chunk), tt.maxLen, chunk)
				}
				assertChunkWellFormed(t, i, chunk)
			}

			if tt.assert != nil {
				tt.assert(t, chunks)
			}
		})
	}
}

func assertWrappedChunks(t *testing.T, chunks []string, prefix, suffix string) {
	t.Helper()

	for i, chunk := range chunks {
		if !strings.HasPrefix(chunk, prefix) {
			t.Fatalf("chunk %d does not start with %q: %q", i, prefix, chunk)
		}
		if !strings.HasSuffix(chunk, suffix) {
			t.Fatalf("chunk %d does not end with %q: %q", i, suffix, chunk)
		}
	}
}

func assertChunkWellFormed(t *testing.T, chunkIndex int, chunk string) {
	t.Helper()

	z := html.NewTokenizer(strings.NewReader(chunk))
	var stack []string

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				t.Fatalf("chunk %d tokenization error: %v", chunkIndex, z.Err())
			}
			if len(stack) > 0 {
				t.Fatalf("chunk %d has unclosed tags: %v", chunkIndex, stack)
			}
			return
		case html.StartTagToken:
			name, _ := z.TagName()
			stack = append(stack, string(name))
		case html.SelfClosingTagToken:
			// self-closing tags do not need to be closed
		case html.EndTagToken:
			name, _ := z.TagName()
			if len(stack) == 0 {
				t.Fatalf("chunk %d has unexpected closing tag </%s>", chunkIndex, name)
			}
			top := stack[len(stack)-1]
			if top != string(name) {
				t.Fatalf("chunk %d mismatched tags: expected </%s>, got </%s>", chunkIndex, top, name)
			}
			stack = stack[:len(stack)-1]
		}
	}
}
