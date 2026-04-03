package main

import (
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/updater"
	service "github.com/sandevgo/tuskbot/internal/service/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update TuskBot to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		var flushLog func()
		ctx, flushLog = setupLogger(ctx)
		defer flushLog()

		provider := updater.NewGitHubProvider(core.TuskRepositorySlug)
		systemSvc, err := NewSystemService(ctx)
		if err != nil {
			return err
		}
		svc := service.NewService(provider, systemSvc, core.TaskVersion)

		fmt.Println("Checking for updates...")
		release, err := svc.Check(ctx)
		if err != nil {
			return err
		}

		if release == nil {
			fmt.Println("Tusk is already up to date.")
			return nil
		}

		fmt.Printf("Updating to version %s...\n", release.Version)
		if err := svc.Update(ctx, release); err != nil {
			return err
		}

		fmt.Println("Update successful.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
