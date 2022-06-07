package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configPath        string
	debug             bool
	defaultConfigPath = os.ExpandEnv("./")
	appName           = "staking-election"
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

	// --debug flag
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(
		versionCmd(),
		validatorsCmd(),
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
