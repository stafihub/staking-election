// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	TaskTicker       int64 //seconds task interval
	KeystorePath     string
	ElectorAccount   string
	StafiHubEndpoint string
	RTokenInfo       []RTokenInfo
	ListenAddr       string
}

type RTokenInfo struct {
	Denom    string
	Endpoint string
}

func Load(defaultCfgFile string) (*Config, error) {
	configFilePath := flag.String("C", defaultCfgFile, "Config file path")
	flag.Parse()

	var cfg = Config{}
	if err := loadSysConfig(*configFilePath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadSysConfig(path string, config *Config) error {
	_, err := os.Open(path)
	if err != nil {
		return err
	}
	if _, err := toml.DecodeFile(path, config); err != nil {
		return err
	}
	fmt.Println("load sysConfig success")
	return nil
}
