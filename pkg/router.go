package pkg

import (
	"log/slog"
	"mkBlog/config"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func GetRouter() *gin.Engine {
	return r
}

func InitRouter() error {
	r = gin.New()
	gin.SetMode(gin.ReleaseMode)
	r.UseH2C = true
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

	// 1) 暴露图片目录为 /article（url: /article/{title}/{name} -> {ImageSavePath}/{title}/{name}）
	imgRoot := config.Cfg.Server.ImageSavePath
	if abs, err := filepath.Abs(imgRoot); err == nil {
		imgRoot = abs
	}
	r.StaticFile("/config.yaml", "./config.yaml")
	r.StaticFS("/article", gin.Dir(imgRoot, false))

	// 可选：为图片添加缓存头
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/article/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	})

	// 2) 其它静态资源
	if cache != nil {
		r.GET("/assets/*any", cache.Handler())
		r.GET("/", cache.Handler())
		r.GET("/index.html", cache.Handler())
		r.NoRoute(func(c *gin.Context) {
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
		// 去掉重复的 /images 映射，统一用 /article
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(404, gin.H{"msg": "not found"})
				return
			}
			c.File("./static/index.html")
		})
	}

	return nil
}
