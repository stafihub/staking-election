// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package db

import (
	"fmt"
	"gorm.io/gorm"
)

//wrap for *gorm.DB
type WrapDb struct {
	*gorm.DB
	isTx bool
}

func NewWrapDb(db *gorm.DB) *WrapDb {
	return &WrapDb{
		DB:   db,
		isTx: false,
	}
}

//transaction begin，must call rollback or commitTransaction after call this
func (d *WrapDb) NewTransaction() *WrapDb {
	txDao := NewWrapDb(d.Begin())
	txDao.isTx = true
	return txDao
}

//Only for transactional dao calls
func (d *WrapDb) RollbackTransaction() error {
	if d.isTx {
		return d.Rollback().Error
	}
	return fmt.Errorf("is not transaction tx")
}

//Only for transactional dao calls
func (d *WrapDb) CommitTransaction() error {
	if d.isTx {
		return d.Commit().Error
	}
	return fmt.Errorf("is not transaction tx")
}
