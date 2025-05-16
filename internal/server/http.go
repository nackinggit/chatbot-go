package server

import "github.com/gin-gonic/gin"

func Route(e *gin.Engine) {
	apiV1 := e.Group("/api/v1/")
	apiV1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
