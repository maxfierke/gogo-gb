package cmd

import (
	"fmt"

	"github.com/maxfierke/gogo-gb/cpu/isa"
	"github.com/spf13/cobra"
)

// opcodesCmd represents the opcodes command
var opcodesCmd = &cobra.Command{
	Use:   "opcodes",
	Short: "Print the opcodes with value and mneumonic",
	Long:  `Print the opcodes with value and mneumonic`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := getLogger(cmd)
		if err != nil {
			return fmt.Errorf("getting logger: %w", err)
		}

		opcodes, err := isa.LoadOpcodes()
		if err != nil {
			return fmt.Errorf("unable to load opcodes: %w", err)
		}

		opcodes.DebugPrint(logger.Writer())

		return nil
	},
}

func init() {
	inspectCmd.AddCommand(opcodesCmd)
}
