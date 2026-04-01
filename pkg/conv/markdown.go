package conv

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

var (
	extensions           = (parser.CommonExtensions &^ parser.Tables) | parser.NoEmptyLineBeforeBlock
	htmlFlags            = html.CommonFlags | html.HrefTargetBlank
	tgPolicy             = bluemonday.NewPolicy()
	tableDelimiterCellRe = regexp.MustCompile(`^:?(?:-|\p{Pd}){3,}:?$`)
)

func init() {
	// Allowed tags https://core.telegram.org/bots/api#html-style
	tgPolicy.AllowElements("b", "strong", "i", "em", "u", "ins", "s", "strike", "del", "code", "pre", "blockquote")
	tgPolicy.AllowAttrs("href").OnElements("a")
	tgPolicy.AllowAttrs("class").OnElements("code")
}

func MarkdownToTelegramHTML(md []byte) string {
	rewrittenMD := rewriteMarkdownTablesAsCodeFences(string(md))

	// Render HTML
	p := parser.NewWithExtensions(extensions)
	renderer := html.NewRenderer(html.RendererOptions{Flags: htmlFlags})
	unsafeHTML := markdown.Render(p.Parse([]byte(rewrittenMD)), renderer)

	// Sanitize tags
	sanitized := tgPolicy.SanitizeBytes(unsafeHTML)

	return normalizeTelegramHTML(string(sanitized))
}

func normalizeLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func trimOuterPipes(line string) string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	return trimmed
}

func rewriteMarkdownTablesAsCodeFences(md string) string {
	if md == "" {
		return md
	}

	md = normalizeLineEndings(md)

	lines := strings.Split(md, "\n")
	var out []string
	out = make([]string, 0, len(lines))

	inFence := false
	for i := 0; i < len(lines); {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if isFenceStartOrEnd(trimmed) {
			inFence = !inFence
			out = append(out, line)
			i++
			continue
		}

		if inFence {
			out = append(out, line)
			i++
			continue
		}

		if i+1 < len(lines) && isPotentialTableRow(line) && isTableDelimiterRow(lines[i+1]) {
			end := i + 2
			for end < len(lines) && isPotentialTableRow(lines[end]) {
				end++
			}

			headers := parseMarkdownTableRow(line)
			rows := make([][]string, 0, end-(i+2))
			for rowIndex := i + 2; rowIndex < end; rowIndex++ {
				rows = append(rows, parseMarkdownTableRow(lines[rowIndex]))
			}

			out = append(out, "```", formatTableAsCards(headers, rows), "```")
			i = end
			continue
		}

		out = append(out, line)
		i++
	}

	return strings.Join(out, "\n")
}

func parseMarkdownTableRow(line string) []string {
	parts := strings.Split(trimOuterPipes(line), "|")
	cells := make([]string, len(parts))
	for i, part := range parts {
		cells[i] = strings.TrimSpace(part)
	}

	return cells
}

func formatTableAsCards(headers []string, rows [][]string) string {
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	var b strings.Builder
	for rowIndex, row := range rows {
		if rowIndex > 0 {
			b.WriteString("\n\n")
		}

		b.WriteString(strconv.Itoa(rowIndex + 1))
		b.WriteString(")\n")
		for i, header := range headers {
			value := ""
			if i < len(row) {
				value = strings.TrimSpace(row[i])
			}

			if i == len(headers)-1 && len(row) > len(headers) {
				extras := make([]string, 0, len(row)-len(headers))
				for _, extra := range row[len(headers):] {
					extra = strings.TrimSpace(extra)
					if extra != "" {
						extras = append(extras, extra)
					}
				}
				if len(extras) > 0 {
					extraValue := strings.Join(extras, " | ")
					if value == "" {
						value = extraValue
					} else {
						value += " | " + extraValue
					}
				}
			}

			b.WriteString(header)
			b.WriteString(": ")
			b.WriteString(value)
			if i < len(headers)-1 {
				b.WriteByte('\n')
			}
		}
	}

	return b.String()
}

func isFenceStartOrEnd(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

func isPotentialTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed != "" && strings.Contains(trimmed, "|")
}

func isTableDelimiterRow(line string) bool {
	trimmed := trimOuterPipes(line)
	if trimmed == "" {
		return false
	}

	parts := strings.Split(trimmed, "|")
	if len(parts) == 0 {
		return false
	}

	for _, part := range parts {
		cell := strings.TrimSpace(part)
		if cell == "" || !tableDelimiterCellRe.MatchString(cell) {
			return false
		}
	}

	return true
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
