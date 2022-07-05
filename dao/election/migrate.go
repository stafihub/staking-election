// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package dao_election

import (
	"github.com/stafihub/staking-election/db"
)

func AutoMigrate(db *db.WrapDb) error {
	return db.Set("gorm:table_options", "ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8").
		AutoMigrate(SelectedValidator{}, AnnualRate{})
}
