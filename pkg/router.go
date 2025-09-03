package pkg

import (
	"github.com/gin-gonic/gin"
)

func NewRouter() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.100"})

	// 精确挂载静态资源，避免根通配符与 /api 冲突
	r.Static("/assets", "./static/assets")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	// 暴露根目录 config.yaml 用于前端运行时加载
	r.StaticFile("/config.yaml", "./config.yaml")
	r.StaticFile("/", "./static/index.html")

	// SPA 回退
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	return r, nil
}
