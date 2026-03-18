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
	Short: "Start the TuskBot services",
	Long:  `Initializes and starts all configured services (Telegram, CLI, etc.) and background workers.`,
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
}
