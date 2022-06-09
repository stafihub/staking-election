// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package api

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

// Create a map to hold the rate limiters for each visitor and a mutex.
var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

// Retrieve and return the rate limiter for the current visitor if it
// already exists. Otherwise create a new rate limiter and add it to
// the visitors map, using the IP address as the key.
func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Minute), 5)
		visitors[ip] = limiter
	}

	return limiter
}

func IpRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Call the getVisitor function to retreive the rate limiter for
		// the current user.
		limiter := getVisitor(ip)
		if !limiter.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
		c.Next()
	}
}

func MethodRateLimiter() gin.HandlerFunc {
	globalRate := rate.NewLimiter(rate.Every(time.Minute), 10)
	return func(c *gin.Context) {
		if !globalRate.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
		c.Next()
	}
}
