package routers

import (
	"github.com/gin-gonic/gin"
	"mkBlog/internal/handler"
)

var Router *gin.Engine

func InitRouter() {
	Router = gin.Default()
	Router.GET("/", handler.HomeHandler)
	Router.GET("/home", handler.HomeHandler)
	Router.GET("/articles/:title", handler.DetailHandler)
	Router.GET("/images/:title/:path", handler.ImageHandler)
	Router.GET("/search", handler.SearchHandler)
	Router.GET("/friend", handler.GetFriend)
	Router.POST("/friend/apply", handler.ApplyFriend)
	Router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许前端域名
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // 处理预检请求
			return
		}
		c.Next()
	})
}
