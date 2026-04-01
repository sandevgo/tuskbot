package conv

import (
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var (
	extensions = (parser.CommonExtensions &^ parser.Tables) | parser.NoEmptyLineBeforeBlock
	htmlFlags  = html.CommonFlags | html.HrefTargetBlank
	tgPolicy   = bluemonday.NewPolicy()
)

func init() {
	// Allowed tags https://core.telegram.org/bots/api#html-style
	tgPolicy.AllowElements("b", "strong", "i", "em", "u", "ins", "s", "strike", "del", "code", "pre", "blockquote")
	tgPolicy.AllowAttrs("href").OnElements("a")
	tgPolicy.AllowAttrs("class").OnElements("code")
}

func MarkdownToTelegramHTML(md []byte) string {
	// Render HTML
	p := parser.NewWithExtensions(extensions)
	renderer := html.NewRenderer(html.RendererOptions{Flags: htmlFlags})
	unsafeHTML := markdown.Render(p.Parse(md), renderer)

	// Sanitize tags
	sanitized := tgPolicy.SanitizeBytes(unsafeHTML)

	return normalizeTelegramHTML(string(sanitized))
}

func normalizeTelegramHTML(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	if strings.TrimSpace(s) == "" {
		if strings.Contains(s, "\n") {
			return "\n"
		}
		return ""
	}

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	var b strings.Builder
	b.Grow(len(s))

	prevBlank := false
	for i, line := range lines {
		blank := line == ""
		if blank && prevBlank {
			continue
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
		prevBlank = blank
	}

	return b.String()
}
