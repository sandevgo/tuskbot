package main

import (
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/service/installer"
	"github.com/sandevgo/tuskbot/pkg/log"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:           "install",
	Short:         "Install TuskBot and its dependencies",
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// Setup logger
		var flushLog func()
		ctx, flushLog = setupLogger(ctx)
		defer flushLog()

		logger := log.FromCtx(ctx)
		logger.Info().Msg("starting installation process")

		// run wizard (includes save step)
		_, err := installer.RunWizard()
		if err != nil {
			return err
		}

		// Load the newly created .env file so NewAppConfig can see the values
		runtimePath := config.GetRuntimePath()
		envPath := filepath.Join(runtimePath, ".env")
		if err := godotenv.Load(envPath); err != nil {
			logger.Warn().Err(err).Str("path", envPath).Msg("failed to load .env file")
		}

		logger.Info().Msgf("initialized runtime directory at: %s", runtimePath)
		logger.Info().Msg("Installation complete! You can now run 'tusk start'.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
