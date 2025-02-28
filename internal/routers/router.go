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
	Router.GET("/articles/:id", handler.DetailHandler)
	Router.GET("/images/:title/:path", handler.ImageHandler)
	Router.GET("/search", handler.SearchHandler)
}
