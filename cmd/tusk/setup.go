package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/eventbus"
	"github.com/sandevgo/tuskbot/internal/providers/llm"
	"github.com/sandevgo/tuskbot/internal/providers/mcp"
	"github.com/sandevgo/tuskbot/internal/providers/mcp/tools"
	"github.com/sandevgo/tuskbot/internal/providers/rag"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/internal/service/command"
	"github.com/sandevgo/tuskbot/internal/service/memory"
	"github.com/sandevgo/tuskbot/internal/service/migration"
	"github.com/sandevgo/tuskbot/internal/service/state"
	"github.com/sandevgo/tuskbot/internal/service/swarm"
	"github.com/sandevgo/tuskbot/internal/storage/sqlite"
	"github.com/sandevgo/tuskbot/internal/transport/scheduler"
	"github.com/sandevgo/tuskbot/internal/transport/telegram"
	"github.com/sandevgo/tuskbot/pkg/log"
	pkgservice "github.com/sandevgo/tuskbot/pkg/service"
	"github.com/sandevgo/tuskbot/pkg/srv"
)

func NewServices(ctx context.Context) []srv.Service {
	logger := log.FromCtx(ctx)

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
	services := make([]srv.Service, 0)
	services = append(services, srv.NewCleanup(db.Close))

	// Run migrations after DB is ready
	if err := migration.Run(db, appCfg); err != nil {
		logger.Fatal().Err(err).Msg("failed to run migrations")
	}

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

	// 4.1 EventBus
	ebus := eventbus.New(100)

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
		rag.NewTiktokenCounter(),
		memory.NewSysPrompt(config.GetRuntimePath(), appCfg),
	)

	executor := agent.NewExecutor(mcpManager)

	// 7. Agent Service
	ag := agent.NewAgent(
		aiProvider,
		mcpManager,
		mem,
		executor,
		ebus,
	)

	// Scheduler
	scheduleService := scheduler.NewScheduler()
	services = append(services, scheduleService)

	// TaskRepo
	taskRepo := sqlite.NewTaskRepo(db)

	// SubAgent
	subAgent := agent.NewSubAgent(aiProvider, mcpManager, mem, executor)

	// Swarm
	swarmService := swarm.NewService(scheduleService, ag, subAgent, taskRepo)
	services = append(services, swarmService)

	// Register tools
	mcpManager.RegisterNativeTool(tools.NewFilesystem(appCfg.GetRuntimePath()))
	mcpManager.RegisterNativeTool(tools.NewShell(appCfg.GetRuntimePath()))
	mcpManager.RegisterNativeTool(tools.NewFetch())
	mcpManager.RegisterNativeTool(tools.NewSchedule(swarmService))

	// commands
	commands := command.NewCommands(appCfg, globState, mcpManager, swarmService, mem)
	cmdRouter := command.New(commands)

	// 8. Transports
	transports, err := initTransports(ctx, appCfg, ag, cmdRouter, ebus)
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
		mcp.NewPool(),
		mcp.NewRegistry(filStorage),
		mcp.NewToolCache(),
	)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func initTransports(
	ctx context.Context,
	cfg *config.AppConfig,
	ag core.Agent,
	router core.CmdRouter,
	subs core.EventSubscriber,
) ([]srv.Service, error) {
	var services []srv.Service

	// Telegram Bot
	if cfg.IsTelegramSelected() {
		bot, err := telegram.NewBot(cfg, ag, router, subs)
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

func NewSystemService(ctx context.Context) (core.SystemService, error) {
	runtimePath := config.GetRuntimePath()
	if err := initEnv(ctx, runtimePath); err != nil {
		return nil, err
	}

	svcCfg := config.NewSystemServiceConfig(ctx)
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	return pkgservice.NewKardianosManager(pkgservice.Config{
		Name:             svcCfg.Name,
		DisplayName:      svcCfg.DisplayName,
		Description:      svcCfg.Description,
		Executable:       execPath,
		Arguments:        []string{"run"},
		WorkingDirectory: runtimePath,
		UserService:      svcCfg.UserService,
		SandboxMode:      svcCfg.SandboxMode,
	})
}
