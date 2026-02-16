package router

import (
	"mkBlog/config"
	"mkBlog/pkg/cache"
	"mkBlog/pkg/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func GetRouter() *gin.Engine {
	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果你只想允许自己域名，把 * 换成具体域名
		// 例如：https://mkitsdts.top
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// 处理预检请求，直接返回 200
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func InitRouter() error {
	gin.SetMode(gin.ReleaseMode)
	r = gin.New()
	r.Use(gin.Logger(), gin.Recovery(), CORSMiddleware())
	// 启用黑名单
	if !config.Cfg.Server.Devmode {
		r.Use(middleware.Blacklist())
	}

	// 1 暴露图片目录为 /article（支持无后缀访问，自动追加 .webp）
	imgRoot := config.Cfg.Server.ImageSavePath
	if abs, err := filepath.Abs(imgRoot); err == nil {
		imgRoot = abs
	}
	//
	r.GET("/api/site", func(c *gin.Context) {
		site := config.Cfg.Site
		public := struct {
			Signature      string `json:"signature"`
			About          string `json:"about"`
			AvatarPath     string `json:"avatarPath"`
			Server         string `json:"server"`
			DevMode        bool   `json:"devmode"`
			CommentEnabled bool   `json:"comment_enabled"`
			ICP            string `json:"icp"`
		}{
			Signature:      site.Signature,
			About:          site.About,
			AvatarPath:     site.AvatarPath,
			Server:         site.Server,
			DevMode:        site.DevMode,
			CommentEnabled: site.CommentEnabled,
			ICP:            site.ICP,
		}
		c.JSON(200, public)
	})
	// 自定义处理：优先尝试原路径；如最后一段无扩展名，则尝试追加 .webp
	r.GET("/article/*rel", func(c *gin.Context) {
		rel := strings.TrimPrefix(c.Param("rel"), "/")
		// 规范化并防止目录穿越
		clean := filepath.Clean(rel)
		candidate := filepath.Join(imgRoot, clean)
		// 确保在根目录之下
		if !strings.HasPrefix(candidate+string(os.PathSeparator), imgRoot+string(os.PathSeparator)) && candidate != imgRoot {
			c.JSON(400, gin.H{"msg": "invalid path"})
			return
		}

		// 如果是单段路径 /article/:title（没有后续文件名），这是前端 SPA 的文章详情路由，直接返回 index.html
		if !strings.Contains(rel, "/") || strings.HasSuffix(clean, "/") {
			if cache.GetGlobalAssetCache() != nil {
				// 复用缓存处理器返回 SPA 入口
				c.Request.URL.Path = "/"
				cache.GetGlobalAssetCache().Handler()(c)
			} else {
				c.File("./static/index.html")
			}
			return
		}

		// 如果带扩展名，直接尝试该文件
		base := filepath.Base(clean)
		if dot := strings.LastIndexByte(base, '.'); dot > 0 {
			if fileExists(candidate) {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
				c.File(candidate)
				return
			}
			c.JSON(404, gin.H{"msg": "image not found"})
			return
		}

		// 无扩展名：尝试追加 .webp
		webpPath := candidate + ".webp"
		if fileExists(webpPath) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.File(webpPath)
			return
		}
		// 也尝试原路径（例如目录索引被禁用，将返回 404）
		if fileExists(candidate) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.File(candidate)
			return
		}
		c.JSON(404, gin.H{"msg": "image not found"})
	})

	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/article/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	})

	// 2) 其它静态资源
	if cache.GetGlobalAssetCache() != nil {
		r.GET("/assets/*any", cache.GetGlobalAssetCache().Handler())
		r.GET("/", cache.GetGlobalAssetCache().Handler())
		r.GET("/index.html", cache.GetGlobalAssetCache().Handler())
		r.GET("/icon.svg", cache.GetGlobalAssetCache().Handler())
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(404, gin.H{"msg": "not found"})
				return
			}
			c.Request.URL.Path = "/"
			cache.GetGlobalAssetCache().Handler()(c)
		})
	} else {
		r.Static("/assets", "./static/assets")
		r.StaticFile("/", "./static/index.html")
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

// fileExists checks if a regular file exists at the given path.
func fileExists(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}
