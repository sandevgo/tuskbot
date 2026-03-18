package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
	"github.com/sandevgo/tuskbot/pkg/srv"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TuskBot user/system service",
	Long:  `Starts installed TuskBot service (alias for 'tusk service start'). Use 'tusk run' for foreground mode.`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Start(ctx); err != nil {
			return "", err
		}
		return "Service started.", nil
	}),
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop TuskBot user/system service",
	Long:  `Stops installed TuskBot service (alias for 'tusk service stop').`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Stop(ctx); err != nil {
			return "", err
		}
		return "Service stopped.", nil
	}),
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run TuskBot in foreground",
	Long:  `Runs TuskBot in the foreground process (PID 1 friendly for Docker/Kubernetes).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		return runApp(ctx)
	},
}

func runApp(ctx context.Context) error {
	// logger setup
	var flushLog func()
	ctx, flushLog = setupLogger(ctx)
	defer flushLog()

	logger := log.FromCtx(ctx)
	logger.Info().Msgf("starting tuskbot %s", core.TaskVersion)

	// Define services using the setup.go logic
	services := NewServices(ctx)

	// Start services
	srv.StartServices(ctx, services)

	// Wait for shutdown signal
	srv.ShutdownServices(ctx, services)
	logger.Info().Msg("tuskbot has been shut down gracefully")

	return nil
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(runCmd)
}
