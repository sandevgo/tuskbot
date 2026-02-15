package main

import (
	"context"
	"os"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/service/ui"
	"github.com/sandevgo/tuskbot/pkg/log"
	"github.com/spf13/cobra"
)

var (
	debug bool
)

var rootCmd = &cobra.Command{
	Use:   "tusk",
	Short: "TuskBot — A Multi-Agent System",
	Long:  `TuskBot is a personal agentic bot.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all subcommands
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", config.IsDebug(), "enable debug logging")
}

func setupLogger(ctx context.Context) (context.Context, func()) {
	isDebug := debug || config.IsDebug()
	return log.NewContextWithLogger(ctx, isDebug)
}

func CustomizeHelp(rootCmd *cobra.Command) {

	cobra.AddTemplateFunc("StyleTitle", func(s string) string { return ui.TitleStyle.Render(s) })
	cobra.AddTemplateFunc("StyleUsage", func(s string) string { return ui.UsageStyle.Render(s) })
	cobra.AddTemplateFunc("StyleFlag", func(s string) string { return ui.FlagStyle.Render(s) })
	cobra.AddTemplateFunc("StyleDesc", func(s string) string { return ui.DescStyle.Render(s) })

	template := `
{{StyleTitle "USAGE"}}
  {{.UseLine}}
{{if gt (len .Commands) 0}}{{StyleTitle "AVAILABLE COMMANDS"}}
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding}} {{StyleDesc .Short}}{{end}}
{{end}}{{end}}
{{if .HasAvailableLocalFlags}}{{StyleTitle "FLAGS"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}
`
	// 3. Применяем шаблон
	rootCmd.SetHelpTemplate(template)
}
