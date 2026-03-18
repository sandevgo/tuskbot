package main

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/spf13/cobra"
)

func runWithSystemService(action func(cmd *cobra.Command, ctx context.Context, svc core.SystemService) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := NewSystemService(ctx)
		if err != nil {
			return err
		}
		return action(cmd, ctx, svc)
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
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) error {
		return svc.Install(ctx)
	}),
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall TuskBot service",
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) error {
		return svc.Uninstall(ctx)
	}),
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TuskBot service",
	Long:  `Starts installed TuskBot service. Alias: 'tusk start'.`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) error {
		return svc.Start(ctx)
	}),
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop TuskBot service",
	Long:  `Stops installed TuskBot service. Alias: 'tusk stop'.`,
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) error {
		return svc.Stop(ctx)
	}),
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart TuskBot system service",
	RunE: runWithSystemService(func(_ *cobra.Command, ctx context.Context, svc core.SystemService) error {
		if err := svc.Stop(ctx); err != nil {
			return err
		}
		return svc.Start(ctx)
	}),
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show TuskBot system service status",
	RunE: runWithSystemService(func(cmd *cobra.Command, ctx context.Context, svc core.SystemService) error {
		status, err := svc.Status(ctx)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), status)
		return err
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
