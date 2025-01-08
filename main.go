package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	routers := gin.New()
	routers.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	routers.Run(":8080")
}
