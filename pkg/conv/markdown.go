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

func rewriteMarkdownTablesAsCodeFences(md string) string {
	if md == "" {
		return md
	}

	md = strings.ReplaceAll(md, "\r\n", "\n")
	md = strings.ReplaceAll(md, "\r", "\n")

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
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "|") {
		trimmed = strings.TrimPrefix(trimmed, "|")
	}
	if strings.HasSuffix(trimmed, "|") {
		trimmed = strings.TrimSuffix(trimmed, "|")
	}

	parts := strings.Split(trimmed, "|")
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
			b.WriteByte('\n')
		}

		values := mapTableRowToHeaders(headers, row)

		b.WriteString(strconv.Itoa(rowIndex + 1))
		b.WriteString(")\n")
		for i, header := range headers {
			b.WriteString(header)
			b.WriteString(": ")
			b.WriteString(values[i])
			if i < len(headers)-1 {
				b.WriteByte('\n')
			}
		}
	}

	return b.String()
}

func mapTableRowToHeaders(headers, row []string) []string {
	values := make([]string, len(headers))
	for i := range headers {
		if i < len(row) {
			values[i] = strings.TrimSpace(row[i])
		}
	}

	if len(headers) == 0 || len(row) <= len(headers) {
		return values
	}

	extras := make([]string, 0, len(row)-len(headers))
	for _, extra := range row[len(headers):] {
		extra = strings.TrimSpace(extra)
		if extra != "" {
			extras = append(extras, extra)
		}
	}
	if len(extras) == 0 {
		return values
	}

	extraValue := strings.Join(extras, " | ")
	last := len(values) - 1
	if values[last] == "" {
		values[last] = extraValue
	} else {
		values[last] = values[last] + " | " + extraValue
	}

	return values
}

func isFenceStartOrEnd(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

func isPotentialTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if !strings.Contains(trimmed, "|") {
		return false
	}
	return strings.Count(trimmed, "|") >= 1
}

func isTableDelimiterRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	if strings.HasPrefix(trimmed, "|") {
		trimmed = strings.TrimPrefix(trimmed, "|")
	}
	if strings.HasSuffix(trimmed, "|") {
		trimmed = strings.TrimSuffix(trimmed, "|")
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
