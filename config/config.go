// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Config struct {
	KeystorePath         string
	ElectorAccount       string
	StafiHubEndpointList []string
	GasPrice             string
	ListenAddr           string
	RTokenInfo           []RTokenInfo

	Db Db
}

type Db struct {
	Host string
	Port string
	User string
	Pwd  string
	Name string
}
type RTokenInfo struct {
	Denom           string
	MaxCommission   *Dec  `toml:",omitempty"`
	MaxMissedBlocks int64 `toml:",omitempty"`
	EndpointList    []string
}

func Load(configFilePath string) (*Config, error) {
	var cfg = Config{}
	if err := loadSysConfig(configFilePath, &cfg); err != nil {
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
	fmt.Println("load config success")
	return nil
}

type Dec struct {
	sdk.Dec
}

func (d *Dec) UnmarshalTOML(v interface{}) error {
	if v == nil {
		d = nil
		return nil
	}
	dec, err := sdk.NewDecFromStr(v.(string))
	if err != nil {
		return err
	}

	d.Dec = dec
	return nil
}
