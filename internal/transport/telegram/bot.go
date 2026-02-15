package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/pkg/conv"
	"github.com/sandevgo/tuskbot/pkg/log"
	tele "gopkg.in/telebot.v3"
)

const baseContextKey = "base_context"

type Bot struct {
	bot     *tele.Bot
	cfg     *config.TelegramConfig
	agent   *agent.Agent
	ownerID int64
}

func NewBot(
	ctx context.Context,
	cfg *config.TelegramConfig,
	agent *agent.Agent,
) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	bot := &Bot{
		bot:     b,
		cfg:     cfg,
		agent:   agent,
		ownerID: cfg.OwnerID,
	}

	// Use context from Signal with logger
	b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			c.Set(baseContextKey, ctx)
			return next(c)
		}
	})

	// Middleware: Only allow the owner
	b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender().ID != bot.ownerID {
				return nil // Ignore unauthorized users
			}
			return next(c)
		}
	})

	b.Handle(tele.OnText, bot.handleMessage)

	return bot, nil
}

func (b *Bot) Start(ctx context.Context) error {
	log.FromCtx(ctx).Info().Msg("starting telegram bot")
	b.bot.Start()
	return nil
}

func (b *Bot) Shutdown(ctx context.Context) error {
	b.bot.Stop()
	return nil
}

func (b *Bot) handleMessage(c tele.Context) error {
	// Create a context for this request
	ctx := c.Get(baseContextKey).(context.Context)
	logger := log.FromCtx(ctx)
	sessionID := fmt.Sprintf("telegram-%d", c.Chat().ID)

	// Notify user we are working
	_ = c.Notify(tele.Typing)

	_, err := b.agent.Run(ctx, sessionID, c.Text(), func(msg core.Message) {
		// Send Reasoning (optional, good for debugging)
		//if strings.TrimSpace(msg.Reasoning) != "" {
		//	_ = c.Send(fmt.Sprintf("💭 <b>Thinking:</b>\n%s", msg.Reasoning), tele.ModeHTML)
		//  _ = c.Notify(tele.Typing)
		//}

		// Send Content
		if msg.Content != "" {
			htmlContent := strings.TrimSpace(conv.MarkdownToTelegramHTML([]byte(msg.Content)))
			if htmlContent != "" {
				if err := c.Send(htmlContent, tele.ModeHTML); err != nil {
					logger.Error().Err(err).Msg("failed to send telegram message")
				}
				_ = c.Notify(tele.Typing)
			}
		}

		// Notify about tool execution
		for _, tc := range msg.ToolCalls {
			_ = c.Send(fmt.Sprintf("🛠 Executing: %s", tc.Function.Name))
			_ = c.Notify(tele.Typing)
		}
	})

	if err != nil {
		logger.Error().Err(err).Msg("agent run failed")
		return c.Send(fmt.Sprintf("error: %v", err))
	}

	return nil
}
