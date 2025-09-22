package pkg

import (
	"log/slog"
	"mkBlog/config"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 启用限流器
	r.Use(RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))

	// 允许本地访问
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// 构建静态资源内存缓存（假设构建产物都放在 ./static）
	cache, err := BuildAssetCache("./static")
	if err != nil {
		slog.Warn("build asset cache failed,", "error:", err)
	}

	// 手工注册处理静态文件
	// 1. config.yaml 仍直接文件读取（便于随时改），如果想缓存也可仿照处理
	r.StaticFile("/config.yaml", "./config.yaml")
	r.StaticFS("/images", gin.Dir(config.Cfg.Server.ImageSavePath, false))

	// 可选：为图片添加缓存头
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/images/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	})

	// 2. assets & index 走内存
	if cache != nil {
		r.GET("/assets/*any", cache.Handler())
		r.GET("/", cache.Handler())
		r.GET("/index.html", cache.Handler())
		// SPA 回退
		r.NoRoute(func(c *gin.Context) {
			// 对 /api/* 返回 404 JSON，避免错误地返回 index.html
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(404, gin.H{"msg": "not found"})
				return
			}
			c.Request.URL.Path = "/"
			cache.Handler()(c)
		})
	} else {
		r.Static("/assets", "./static/assets")
		r.StaticFile("/", "./static/index.html")
		r.StaticFS("/images", gin.Dir(config.Cfg.Server.ImageSavePath, false))
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(404, gin.H{"msg": "not found"})
				return
			}
			c.File("./static/index.html")
		})
	}

	return r, nil
}
