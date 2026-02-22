package integration

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
//	appCfg := config.NewAppConfig(ctx)
//
//	filStorage := mcp.NewFileStorage(appCfg.GetMCPConfigPath())
//	mcpService, err := mcp.NewService(
//		appCfg.GetRuntimePath(),
//		mcp.NewPool(),
//		mcp.NewRegistry(filStorage),
//		mcp.NewToolCache(),
//	)
//	if err != nil {
//		t.Error(err)
//	}
//
//	go func() {
//		if err := mcpService.Start(ctx); err != nil {
//			t.Error(err)
//		}
//	}()
//
//	time.Sleep(5 * time.Second)
//
//	tools, err := mcpService.GetTools(ctx)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	var toolName string
//	for _, tool := range tools {
//		t.Log(tool)
//		if tool.Function.Name == "tavily" || tool.Function.Name == "tavily.search" {
//			toolName = tool.Function.Name
//			break
//		}
//	}
//	if toolName == "" && len(tools) > 0 {
//		toolName = tools[len(tools)-1].Function.Name
//	}
//	if toolName == "" {
//		t.Fatal("no tools found")
//	}
//	t.Logf("using tool: %s", toolName)
//
//	call := core.ToolCall{
//		ID:   "call_function_7wjc8y3tq85m_1",
//		Type: "function",
//		Function: core.FunctionCall{
//			Name:      toolName,
//			Arguments: "{\"query\": \"OpenAI o1 model features summary\"}",
//		},
//	}
//
//	calls := []core.ToolCall{call}
//
//	result, err := mcpService.CallTool(ctx, toolName, calls[0].Function.Arguments)
//	if err != nil {
//		t.Fatalf("tool execution failed: %v", err)
//	}
//	fmt.Println(result)
//}
//
//func initEnv(ctx context.Context, runtimePath string) error {
//	logger := log.FromCtx(ctx)
//	envFile := filepath.Join(runtimePath, ".env")
//
//	if _, err := os.Stat(envFile); err != nil {
//		if os.IsNotExist(err) {
//			return nil
//		}
//		return err
//	}
//
//	if err := godotenv.Load(envFile); err != nil {
//		logger.Warn().Err(err).Str("path", envFile).Msg("failed to load .env file")
//		return err
//	}
//
//	logger.Debug().Str("path", envFile).Msg("loaded .env file")
//	return nil
//}
