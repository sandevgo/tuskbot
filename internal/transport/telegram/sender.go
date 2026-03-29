package telegram

import (
	"bytes"
	"context"
	"strings"

	"github.com/sandevgo/tuskbot/pkg/conv"
	"github.com/sandevgo/tuskbot/pkg/log"
	"golang.org/x/net/html"
	tele "gopkg.in/telebot.v3"
)

const maxTelegramMsgLen = 4000 // Safety margin below 4096

type sender struct {
	bot *tele.Bot
}

func newSender(bot *tele.Bot) *sender {
	return &sender{bot: bot}
}

// sendMarkdown converts Markdown to Telegram HTML and sends it in chunks if needed.
func (s *sender) sendMarkdown(ctx context.Context, to tele.Recipient, md string, silent bool) error {
	logger := log.FromCtx(ctx)
	htmlContent := strings.TrimSpace(conv.MarkdownToTelegramHTML([]byte(md)))

	chunks := splitHTML(htmlContent, maxTelegramMsgLen)
	for i, chunk := range chunks {
		opts := []interface{}{tele.ModeHTML}
		if silent && i == 0 {
			opts = append(opts, tele.Silent)
		}

		if _, err := s.bot.Send(to, chunk, opts...); err != nil {
			logger.Error().Err(err).Int("chunk", i).Int("len", len(chunk)).Msg("failed to send telegram chunk")
			return err
		}
	}
	return nil
}

type tagInfo struct {
	name     string
	fullOpen string
}

// splitHTML splits HTML into Telegram-safe chunks using a DOM parser.
// It ensures tags are closed correctly in each chunk.
func splitHTML(text string, maxLen int) []string {
	if maxLen <= 0 || len(text) <= maxLen {
		return []string{text}
	}

	doc, err := html.Parse(strings.NewReader(text))
	if err != nil {
		// Fallback to simple text splitting if HTML is malformed
		return splitText(text, maxLen)
	}

	var chunks []string
	var buf bytes.Buffer
	var stack []tagInfo

	// calculate length of closing tags for current stack
	closingLen := func() int {
		l := 0
		for i := len(stack) - 1; i >= 0; i-- {
			l += 3 + len(stack[i].name) // </tag>
		}
		return l
	}

	// write closing tags to buffer
	writeClosing := func() {
		for i := len(stack) - 1; i >= 0; i-- {
			buf.WriteString("</")
			buf.WriteString(stack[i].name)
			buf.WriteByte('>')
		}
	}

	// write opening tags to buffer (used when starting a new chunk)
	writeOpening := func() {
		for _, t := range stack {
			buf.WriteString(t.fullOpen)
		}
	}

	// traverse the DOM tree
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			text := n.Data
			for len(text) > 0 {
				available := maxLen - buf.Len() - closingLen()
				if available <= 0 {
					// Flush current chunk
					writeClosing()
					chunks = append(chunks, buf.String())
					buf.Reset()
					writeOpening()
					available = maxLen - buf.Len() - closingLen()
					if available <= 0 {
						// Safety break if maxLen is too small for the tag stack
						return
					}
				}

				// Determine how much text we can take.
				// We start by assuming we can take the rest of the text or fill available space.
				take := len(text)
				if take > available {
					take = available
				}

				// Escape the candidate text.
				// html.EscapeString expands entities (e.g. & -> &amp;), increasing length.
				candidate := html.EscapeString(text[:take])
				
				// If escaped text is too long, shrink take until it fits.
				for len(candidate) > available && take > 0 {
					take--
					candidate = html.EscapeString(text[:take])
				}

				if take == 0 {
					// Not enough space for even one character (after escaping), flush buffer.
					writeClosing()
					chunks = append(chunks, buf.String())
					buf.Reset()
					writeOpening()
					continue
				}

				// Try to find a nice break point (newline or space) if we are not taking everything.
				// We only look for breakpoints if we are truncating the text.
				foundNewline := false
				if take < len(text) {
					if idx := strings.LastIndexByte(text[:take], '\n'); idx > 0 {
						take = idx + 1
						candidate = html.EscapeString(text[:take])
						foundNewline = true
					} else if idx := strings.LastIndexByte(text[:take], ' '); idx > 0 {
						take = idx + 1
						candidate = html.EscapeString(text[:take])
					}
				}

				buf.WriteString(candidate)
				text = text[take:]

				// If we split at a newline, force a chunk break for readability
				if foundNewline && len(text) > 0 {
					writeClosing()
					chunks = append(chunks, buf.String())
					buf.Reset()
					writeOpening()
				}
			}

		case html.ElementNode:
			// Skip implicit tags added by the parser
			if n.Data == "html" || n.Data == "head" || n.Data == "body" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					traverse(c)
				}
				return
			}

			openTag := renderOpenTag(n)

			// If adding this tag exceeds limit, flush chunk first
			if buf.Len()+len(openTag)+closingLen() > maxLen {
				writeClosing()
				chunks = append(chunks, buf.String())
				buf.Reset()
				writeOpening()
			}

			buf.WriteString(openTag)
			stack = append(stack, tagInfo{name: n.Data, fullOpen: openTag})

			// Traverse children
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}

			// Render closing tag
			buf.WriteString("</")
			buf.WriteString(n.Data)
			buf.WriteByte('>')
			stack = stack[:len(stack)-1]

		case html.DocumentNode:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}
		}
	}

	traverse(doc)

	if buf.Len() > 0 {
		writeClosing()
		chunks = append(chunks, buf.String())
	}

	return chunks
}

func renderOpenTag(n *html.Node) string {
	var b bytes.Buffer
	b.WriteByte('<')
	b.WriteString(n.Data)
	for _, a := range n.Attr {
		b.WriteByte(' ')
		b.WriteString(a.Key)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(a.Val))
		b.WriteByte('"')
	}
	b.WriteByte('>')
	return b.String()
}

// splitText is a simple fallback for non-HTML text
func splitText(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}
		// simple split
		chunks = append(chunks, text[:maxLen])
		text = text[maxLen:]
	}
	return chunks
}
