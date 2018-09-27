package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// database 数据库中间件
func database(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}
