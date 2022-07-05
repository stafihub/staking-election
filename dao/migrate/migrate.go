// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package migrate

import (
	"fmt"
	"github.com/stafihub/staking-election/dao/election"
	"github.com/stafihub/staking-election/db"
)

func AutoMigrate(db *db.WrapDb) error {
	err := dao_election.AutoMigrate(db)
	if err != nil {
		return fmt.Errorf("dao_user.AutoMigrate %s", err)
	}
	return nil
}
