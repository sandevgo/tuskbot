package main

import (
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(core.TaskVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
