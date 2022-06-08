package cmd

import (
	// sdk "github.com/cosmos/cosmos-sdk/types"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stafihub/staking-election/log"
	// client "github.com/stafihub/cosmos-relay-sdk/client"
)

const flagConfig = "config"

var defaultConfigPath = os.ExpandEnv("./config.toml")

func startCmd() *cobra.Command {
	log.InitConsole()

	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"v"},
		Short:   "start staking-election",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfig)
			if err != nil {
				return err
			}

			logrus.Infof("config path: %s", configPath)
			return nil
		},
	}

	cmd.Flags().String(flagConfig, defaultConfigPath, "config file path")

	return cmd
}
