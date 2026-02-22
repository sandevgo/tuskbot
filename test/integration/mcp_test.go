package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/providers/mcp"
	"github.com/sandevgo/tuskbot/pkg/log"
)

//func TestTavily(t *testing.T) {
//	ctx := context.Background()
//
//	if err := initEnv(ctx, config.GetRuntimePath()); err != nil {
//		t.Fatal(err)
//	}
//
//	var flushLog func()
//	ctx, flushLog = log.NewContextWithLogger(ctx, true)
//	defer flushLog()
//
//	mcpService := initMcp(ctx, t)
//	_, err := mcpService.GetTools(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	toolName := "call_function_7wjc8y3tq85m_1"
//	args := "{\"query\": \"OpenAI o1 model features summary\"}"
//
//	result, err := mcpService.CallTool(ctx, toolName, args)
//	if err != nil {
//		t.Fatalf("tool execution failed: %v", err)
//	}
//	fmt.Println(result)
//}

//func TestContext7(t *testing.T) {
//	ctx := context.Background()
//
//	if err := initEnv(ctx, config.GetRuntimePath()); err != nil {
//		t.Fatal(err)
//	}
//
//	var flushLog func()
//	ctx, flushLog = log.NewContextWithLogger(ctx, true)
//	defer flushLog()
//
//	mcpService := initMcp(ctx, t)
//
//	tools, err := mcpService.GetTools(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for _, tool := range tools {
//		t.Log(tool.Function.Name)
//	}
//
//	toolName := "context7.resolve-library-id"
//	args := "{\"query\": \"react\", \"libraryName\": \"npm\"}"
//
//	t.Logf("using tool: %s", toolName)
//
//	result, err := mcpService.CallTool(ctx, toolName, args)
//	if err != nil {
//		t.Fatalf("tool execution failed: %v", err)
//	}
//	fmt.Println(result)
//}

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

func initMcp(ctx context.Context, t *testing.T) *mcp.Service {
	appCfg := config.NewAppConfig(ctx)

	filStorage := mcp.NewFileStorage(appCfg.GetMCPConfigPath())
	mcpService, err := mcp.NewService(
		appCfg.GetRuntimePath(),
		mcp.NewPool(),
		mcp.NewRegistry(filStorage),
		mcp.NewToolCache(),
	)
	if err != nil {
		t.Error(err)
	}

	go func() {
		if err := mcpService.Start(ctx); err != nil {
			t.Errorf("failed to start mcp server %s", err.Error())
		}
	}()

	time.Sleep(1 * time.Second) // wait for connections
	return mcpService
}
