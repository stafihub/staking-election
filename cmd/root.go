package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	appName = "staking-election"
)

// NewRootCmd returns the root command.
func NewRootCmd() *cobra.Command {
	// RootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: "Staking-election",
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {

		return nil
	}

	rootCmd.AddCommand(
		startCmd(),
		validatorsCmd(),
		versionCmd(),
	)
	return rootCmd
}

func Execute() {
	cobra.EnableCommandSorting = false

	rootCmd := NewRootCmd()
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
