package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/pkg/conv"
	"github.com/sandevgo/tuskbot/pkg/log"
	tele "gopkg.in/telebot.v3"
)

const baseContextKey = "base_context"

type Bot struct {
	bot     *tele.Bot
	cfg     core.TelegramConfig
	agent   *agent.Agent
	router  core.CmdRouter
	ownerID int64
}

func NewBot(
	cfg core.TelegramConfig,
	agent *agent.Agent,
	router core.CmdRouter,
) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.GetTelegramToken(),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &Bot{
		bot:     b,
		cfg:     cfg,
		agent:   agent,
		router:  router,
		ownerID: cfg.GetTelegramOwnerID(),
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	log.FromCtx(ctx).Info().Msg("starting telegram bot")

	// Use context from Signal with logger
	b.bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			c.Set(baseContextKey, ctx)
			return next(c)
		}
	})

	// Middleware: Only allow the owner
	b.bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender().ID != b.ownerID {
				return nil // Ignore unauthorized users
			}
			return next(c)
		}
	})

	b.bot.Handle(tele.OnText, b.handleMessage)

	scope := tele.CommandScope{
		Type:   tele.CommandScopeAllPrivateChats,
		UserID: b.ownerID,
	}

	var cmds []tele.Command
	for _, cmd := range b.router.ListCommands() {
		cmds = append(cmds, tele.Command{
			Text:        cmd.Name(),
			Description: cmd.Description(),
		})
	}

	err := b.bot.SetCommands(cmds, scope)
	if err != nil {
		return fmt.Errorf("failed to set telegram commands: %w", err)
	}

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

	// Check if it's a command
	if response, isCmd := b.router.Execute(ctx, sessionID, c.Text()); isCmd {
		return c.Send(response)
	}

	// Start background typing indicator
	typingCtx, stopTyping := context.WithCancel(ctx)
	defer stopTyping()

	go func() {
		ticker := time.NewTicker(4 * time.Second) // Refresh before 5s expiry
		defer ticker.Stop()

		_ = c.Notify(tele.Typing)

		for {
			select {
			case <-ticker.C:
				_ = c.Notify(tele.Typing)
			case <-typingCtx.Done():
				return
			}
		}
	}()

	_, err := b.agent.Run(ctx, sessionID, c.Text(), func(msg core.Message) {
		// Send Content
		if msg.Content != "" {
			htmlContent := strings.TrimSpace(conv.MarkdownToTelegramHTML([]byte(msg.Content)))
			if htmlContent != "" {
				if err := c.Send(htmlContent, tele.ModeHTML); err != nil {
					logger.Error().Err(err).Msg("failed to send telegram message")
				}
			}
		}

		// Notify about tool execution
		for _, tc := range msg.ToolCalls {
			_ = c.Send(fmt.Sprintf("🛠 Executing: %s", tc.Function.Name))
		}
	})

	if err != nil {
		logger.Error().Err(err).Msg("agent failed to response")
		return c.Send(fmt.Sprintf("Agent failed to response: %v", err))
	}

	return nil
}
