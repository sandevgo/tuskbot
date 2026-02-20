package conv

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var (
	extensions = parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
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

	return string(sanitized)
}
