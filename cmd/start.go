package cmd

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	stafihubClient "github.com/stafihub/stafi-hub-relay-sdk/client"
	"github.com/stafihub/staking-election/config"
	"github.com/stafihub/staking-election/dao/migrate"
	"github.com/stafihub/staking-election/db"
	"github.com/stafihub/staking-election/log"
	"github.com/stafihub/staking-election/server"
	"github.com/stafihub/staking-election/task"
	"github.com/stafihub/staking-election/utils"
)

const (
	flagConfig   = "config"
	flagLogLevel = "log_level"
)

var defaultConfigPath = os.ExpandEnv("./config.toml")

func startElectionCmd() *cobra.Command {
	log.InitConsole()

	cmd := &cobra.Command{
		Use:     "start-election",
		Aliases: []string{"v"},
		Args:    cobra.ExactArgs(0),
		Short:   "Start staking-election procedure",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfig)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)
			logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				return err
			}
			logLevel, err := logrus.ParseLevel(logLevelStr)
			if err != nil {
				return err
			}
			logrus.SetLevel(logLevel)

			conf, err := config.Load(configPath)
			if err != nil {
				return err
			}
			fmt.Printf("\nconfig info: \nelectorAccount: %s\ngasPrice: %s\nkeystorePath: %s\nrTokenInfo: %+v\nstafihubEndpointList: %v\n\n",
				conf.ElectorAccount, conf.GasPrice, conf.KeystorePath, conf.RTokenInfo, conf.StafiHubEndpointList)

			//interrupt signal
			ctx := utils.ShutdownListener()

			fmt.Printf("Will open stafihub wallet from <%s>. \nPlease ", conf.KeystorePath)
			key, err := keyring.New(types.KeyringServiceName(), keyring.BackendFile, conf.KeystorePath, os.Stdin)
			if err != nil {
				return err
			}
			client, err := stafihubClient.NewClient(key, conf.ElectorAccount, conf.GasPrice, conf.StafiHubEndpointList)
			if err != nil {
				return fmt.Errorf("hubClient.NewClient err: %s", err)
			}

			t := task.NewTask(conf, client)
			err = t.Start()
			if err != nil {
				logrus.Errorf("task start err: %s", err)
				return err
			}
			defer func() {
				logrus.Infof("shutting down task ...")
				t.Stop()
			}()
			<-ctx.Done()
			return nil
		},
	}

	cmd.Flags().String(flagConfig, defaultConfigPath, "Config file path")
	cmd.Flags().String(flagLogLevel, logrus.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")

	return cmd
}

func startApiCmd() *cobra.Command {
	log.InitConsole()

	cmd := &cobra.Command{
		Use:     "start-api",
		Aliases: []string{"v"},
		Short:   "Start api server",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfig)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)
			logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				return err
			}
			logLevel, err := logrus.ParseLevel(logLevelStr)
			if err != nil {
				return err
			}
			logrus.SetLevel(logLevel)

			conf, err := config.Load(configPath)
			if err != nil {
				return err
			}
			fmt.Printf("\nconfig info: \nlistenAddr: %s\nrTokenInfo: %+v\nstafihubEndpointList: %v\n\n",
				conf.ListenAddr, conf.RTokenInfo, conf.StafiHubEndpointList)

			//interrupt signal
			ctx := utils.ShutdownListener()

			//init db
			db, err := db.NewDB(&db.Config{
				Host:   conf.Db.Host,
				Port:   conf.Db.Port,
				User:   conf.Db.User,
				Pass:   conf.Db.Pwd,
				DBName: conf.Db.Name,
			})
			if err != nil {
				logrus.Errorf("db err: %s", err)
				return err
			}
			logrus.Infof("db connect success")

			defer func() {
				sqlDb, err := db.DB.DB()
				if err != nil {
					logrus.Errorf("db.DB() err: %s", err)
					return
				}
				logrus.Infof("shutting down the db ...")
				sqlDb.Close()
			}()

			err = migrate.AutoMigrate(db)
			if err != nil {
				logrus.Errorf("dao autoMigrate err: %s", err)
				return err
			}
			client, err := stafihubClient.NewClient(nil, "", "", conf.StafiHubEndpointList)
			if err != nil {
				return fmt.Errorf("hubClient.NewClient err: %s", err)
			}
			//server
			server, err := server.NewServer(conf, client, db)
			if err != nil {
				logrus.Errorf("new server err: %s", err)
				return err
			}
			err = server.Start()
			if err != nil {
				logrus.Errorf("server start err: %s", err)
				return err
			}
			defer func() {
				logrus.Infof("shutting down server ...")
				server.Stop()
			}()

			<-ctx.Done()
			return nil
		},
	}

	cmd.Flags().String(flagConfig, defaultConfigPath, "Config file path")
	cmd.Flags().String(flagLogLevel, logrus.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")

	return cmd
}
