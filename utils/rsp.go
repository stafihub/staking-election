// Copyright 2020 tpkeeper
// SPDX-License-Identifier: LGPL-3.0-only

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	MaxPageSize     = 50
	DefaultPageSize = 10
)

type Rsp struct {
	Status  int64       `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Ok(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "80000",
		"message": msg,
		"data":    data,
	})
}

func Err(c *gin.Context, status, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"status":  status,
		"message": msg,
		"data":    struct{}{},
	})
}
