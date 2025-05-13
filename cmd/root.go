package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

type CLIOptions struct {
	LogPath string
}

var rootCmd = &cobra.Command{
	Use:   "gogo-gb",
	Short: "the go-getting GB emulator",
	Long: `gogo-gb is a scrappy Game Boy emulator written in the Go programming language.

This is a side-project for fun with indeterminate goals and not a stable emulator.
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

const LOG_PREFIX = ""

func getLogger(cmd *cobra.Command) (*log.Logger, error) {
	logPath, err := cmd.Flags().GetString("log")
	if err != nil {
		return nil, fmt.Errorf("getting log flag: %w", err)
	}

	var logFile io.Writer

	switch logPath {
	case "", "stdout":
		logFile = os.Stdout
	case "stderr":
		logFile = os.Stderr
	default:
		logFile, err = os.Create(logPath)
		if err != nil {
			return nil, fmt.Errorf("opening log file '%s' for writing: %w", logPath, err)
		}
	}

	return log.New(logFile, LOG_PREFIX, log.LstdFlags), nil
}

func init() {
	rootCmd.PersistentFlags().String("log", "", "Path to log file. Default/empty implies stdout")
}
