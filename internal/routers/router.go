package routers

import (
	"github.com/gin-gonic/gin"
	"mkBlog/internal/handler"
)

var Router *gin.Engine

func InitRouter() {
	Router = gin.Default()
	Router.GET("/", handler.HomeHandler)
	Router.GET("/detail/:id", handler.DetailHandler)
}
