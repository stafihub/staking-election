// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package db

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maxOpenConn = 10 //connect pool size
	maxIdleConn = 30
)

type Config struct {
	Host, Port, User, Pass, DBName, Mode string
}

//don't use soft delete
type BaseModel struct {
	ID        int64 `gorm:"not null;primaryKey;autoIncrement;column:id"`
	CreatedAt int   `gorm:"type:int(11);autoCreateTime;unsigned;not null;column:create_time"`
	UpdatedAt int   `gorm:"type:int(11);autoUpdateTime;unsigned;not null;column:update_time"`
}

func NewDB(cfg *Config) (wrapDb *WrapDb, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DBName)

	logLevel := logger.Error
	db, err := gorm.Open(
		mysql.New(
			mysql.Config{
				DSN:                       dsn,   // data source name
				DefaultStringSize:         256,   // default size for string fields
				DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
				DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
				DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
				SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
			}),
		&gorm.Config{
			Logger: logger.Default.LogMode(logLevel)})

	if err != nil {
		return nil, err
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxIdleConns(maxIdleConn)
	sqlDb.SetMaxOpenConns(maxOpenConn)
	wrapDb = NewWrapDb(db)
	logrus.Debug("new DB success")
	return
}
