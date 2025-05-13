/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/maxfierke/gogo-gb/cart"
	"github.com/spf13/cobra"
)

var cartHeaderCmd = &cobra.Command{
	Use:   "cart-header",
	Short: "Print cartridge information",
	Long:  `Print cartridge information`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := getLogger(cmd)
		if err != nil {
			return fmt.Errorf("getting logger: %w", err)
		}

		cartPath := args[0]

		cartFile, err := os.Open(cartPath)
		if cartPath == "" || err != nil {
			return fmt.Errorf("unable to open cartridge file: %w", err)
		}
		defer cartFile.Close()

		cartReader, err := cart.NewReader(cartFile)
		if err != nil {
			if errors.Is(err, cart.ErrChecksum) {
				logger.Printf("WARN: Cartridge header does not match expected checksum. Continuing, but subsequent operations may fail")
			} else {
				return fmt.Errorf("unable to read cartridge: %w", err)
			}
		}

		cartReader.Header.DebugPrint(logger.Writer())

		return nil
	},
}

func init() {
	inspectCmd.AddCommand(cartHeaderCmd)
}
