package telegram

import (
	"context"
	"strings"

	"github.com/sandevgo/tuskbot/pkg/conv"
	"github.com/sandevgo/tuskbot/pkg/log"
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
	html := strings.TrimSpace(conv.MarkdownToTelegramHTML([]byte(md)))
	
	chunks := splitHTML(html, maxTelegramMsgLen)
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

// splitHTML splits text into chunks respecting Telegram's limit.
// It tries to split at newlines to preserve formatting.
func splitHTML(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}

		cut := maxLen
		// Try to find a good break point (newline) in the second half of the chunk
		if idx := strings.LastIndex(text[:maxLen], "\n"); idx > maxLen/3 {
			cut = idx
		}

		chunks = append(chunks, text[:cut])
		text = strings.TrimSpace(text[cut:])
	}
	return chunks
}
