package pkg

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func NewRouter() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.100"})

	// 构建静态资源内存缓存（假设构建产物都放在 ./static）
	cache, err := BuildAssetCache("./static")
	if err != nil {
		slog.Warn("build asset cache failed,", "error:", err)
	}

	// 手工注册处理静态文件
	// 1. config.yaml 仍直接文件读取（便于随时改），如果想缓存也可仿照处理
	r.StaticFile("/config.yaml", "./config.yaml")

	// 2. assets & index 走内存
	if cache != nil {
		r.GET("/assets/*any", cache.Handler())
		r.GET("/", cache.Handler())
		r.GET("/index.html", cache.Handler())
		// SPA 回退
		r.NoRoute(func(c *gin.Context) {
			c.Request.URL.Path = "/"
			cache.Handler()(c)
		})
	} else {
		// 回退传统文件方式（构建缓存失败才走）
		r.Static("/assets", "./static/assets")
		r.StaticFile("/", "./static/index.html")
		r.NoRoute(func(c *gin.Context) {
			c.File("./static/index.html")
		})
	}

	return r, nil
}
