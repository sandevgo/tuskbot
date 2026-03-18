package main

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/spf13/cobra"
)

func runWithSystemService(action func(cmd *cobra.Command, ctx context.Context, svc core.SystemService) (string, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := NewSystemService(ctx)
		if err != nil {
			return err
		}

		msg, err := action(cmd, ctx, svc)
		if err != nil {
			return err
		}
		if msg != "" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
		}
		return nil
	}
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage TuskBot user/system service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install TuskBot service",
	Long:  `Installs TuskBot as a service for current user by default (TUSK_SERVICE_USER_MODE=true).`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Install(ctx); err != nil {
			return "", err
		}
		return "Service installed.", nil
	}),
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall TuskBot service",
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Uninstall(ctx); err != nil {
			return "", err
		}
		return "Service uninstalled.", nil
	}),
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TuskBot service",
	Long:  `Starts installed TuskBot service. Alias: 'tusk start'.`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Start(ctx); err != nil {
			return "", err
		}
		return "Service started.", nil
	}),
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop TuskBot service",
	Long:  `Stops installed TuskBot service. Alias: 'tusk stop'.`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Stop(ctx); err != nil {
			return "", err
		}
		return "Service stopped.", nil
	}),
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart TuskBot system service",
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		if err := svc.Stop(ctx); err != nil {
			return "", err
		}
		if err := svc.Start(ctx); err != nil {
			return "", err
		}
		return "Service restarted.", nil
	}),
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show TuskBot system service status",
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) (string, error) {
		status, err := svc.Status(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Service status: %s", status), nil
	}),
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStatusCmd)

	rootCmd.AddCommand(serviceCmd)
}
