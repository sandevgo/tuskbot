package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/llm"
	"github.com/sandevgo/tuskbot/internal/providers/mcp"
	"github.com/sandevgo/tuskbot/internal/providers/rag"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/internal/service/command"
	"github.com/sandevgo/tuskbot/internal/service/memory"
	"github.com/sandevgo/tuskbot/internal/service/state"
	"github.com/sandevgo/tuskbot/internal/storage/sqlite"
	"github.com/sandevgo/tuskbot/internal/transport/telegram"
	"github.com/sandevgo/tuskbot/pkg/log"
	"github.com/sandevgo/tuskbot/pkg/srv"
)

func NewServices(ctx context.Context) []srv.Service {
	logger := log.FromCtx(ctx)
	services := make([]srv.Service, 0)

	// init env
	err := initEnv(ctx, config.GetRuntimePath())
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init env")
	}

	// 1. Configuration
	appCfg := config.NewAppConfig(ctx, config.GetRuntimePath())

	// 2. Storage
	db, messagesRepo, err := initStorage(ctx, appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize storage")
	}
	services = append(services, srv.NewCleanup(db.Close))

	// Knowledge Repo
	knowledgeRepo := sqlite.NewKnowledgeRepo(db)

	// 3. AI Provider
	aiProvider, err := llm.NewDynamicProvider(ctx, appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize LLM provider")
	}

	globState := state.NewGlobalState(aiProvider)

	// 4. RAG Provider (Embedder)
	embedModel, err := rag.NewEmbeddingModel(appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize embedding model")
	}
	services = append(services, srv.NewCleanup(embedModel.Shutdown))

	embedder := rag.NewEmbedder(embedModel)

	// 5. Knowledge Extractor Service
	// Runs in background to convert conversation history into atomic facts
	extractor := memory.NewExtractor(knowledgeRepo, aiProvider, embedder)
	services = append(services, extractor)

	// Embedding extractor
	embedderWorker := memory.NewEmbedderWorker(messagesRepo, embedder)
	services = append(services, embedderWorker)

	// 6. MCP & Tools
	mcpManager, err := initMCP(ctx, appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize MCP manager")
	}
	services = append(services, mcpManager)

	mem := memory.NewMemory(
		appCfg,
		messagesRepo,
		knowledgeRepo,
		embedder,
		memory.NewSysPrompt(appCfg),
	)

	executor := agent.NewExecutor(mcpManager)

	// 7. Agent Service
	ag := agent.NewAgent(
		aiProvider,
		mcpManager,
		mem,
		executor,
	)

	// commands
	commands := command.NewCommands(appCfg, globState, mcpManager)
	cmdRouter := command.New(commands)

	// 8. Transports
	transports, err := initTransports(ctx, appCfg, ag, cmdRouter)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize transports")
	}
	services = append(services, transports...)

	return services
}

// TODO: move Knowledge Repo initialization here
func initStorage(ctx context.Context, cfg *config.AppConfig) (*sql.DB, core.MessagesRepository, error) {
	db, err := sqlite.NewDB(ctx, cfg.GetDatabasePath())
	if err != nil {
		return nil, nil, err
	}
	return db, sqlite.NewMessagesRepo(db), nil
}

func initMCP(ctx context.Context, cfg *config.AppConfig) (*mcp.Service, error) {
	filStorage := mcp.NewFileStorage(cfg.GetMCPConfigPath())
	mgr, err := mcp.NewService(
		config.GetRuntimePath(),
		mcp.NewPool(),
		mcp.NewRegistry(filStorage),
		mcp.NewToolCache(),
	)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func initTransports(ctx context.Context, cfg *config.AppConfig, ag *agent.Agent, router core.CmdRouter) ([]srv.Service, error) {
	var services []srv.Service

	// Telegram Bot
	if cfg.IsTelegramSelected() {
		bot, err := telegram.NewBot(cfg, ag, router)
		if err != nil {
			return nil, err
		}
		services = append(services, bot)
	}

	return services, nil
}

func initEnv(ctx context.Context, runtimePath string) error {
	logger := log.FromCtx(ctx)
	envFile := filepath.Join(runtimePath, ".env")

	if _, err := os.Stat(envFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := godotenv.Load(envFile); err != nil {
		logger.Warn().Err(err).Str("path", envFile).Msg("failed to load .env file")
		return err
	}

	logger.Debug().Str("path", envFile).Msg("loaded .env file")
	return nil
}
