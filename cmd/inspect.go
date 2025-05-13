package cmd

import (
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect cartridges, emulator internals, etc.",
	Long:  `Inspect provides subcommands for various debugging and emulator development purposes`,
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
