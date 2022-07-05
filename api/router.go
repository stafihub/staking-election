// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stafihub/staking-election/api/election_handlers"
	"github.com/stafihub/staking-election/db"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func InitRouters(db *db.WrapDb) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/static", "./static")
	router.Use(Cors())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	rateHandler := election_handlers.NewHandler(db)
	router.GET("/stakingElection/api/v1/annualRateList", rateHandler.HandleGetAverageAnnualRate)
	router.GET("/stakingElection/api/v1/selectedValidators", rateHandler.HandleGetSelectedValidators)

	return router
}
