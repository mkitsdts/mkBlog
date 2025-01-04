package routers

import (
	"github.com/gin-gonic/gin"
)

func InitRouter(routers *gin.Engine) {
	routers.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	routers.Run(":8080")
}
