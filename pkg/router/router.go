package router

import (
	"fmt"
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
	assetCache := cache.GetGlobalAssetCache()
	if assetCache == nil {
		return fmt.Errorf("embedded static asset cache is not initialized")
	}
	assetHandler := assetCache.Handler()

	if !config.Cfg.Server.Devmode {
		gin.SetMode(gin.ReleaseMode)
	}
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
			BgPicturePath  string `json:"bgPicturePath"`
			DevMode        bool   `json:"devmode"`
			CommentEnabled bool   `json:"comment_enabled"`
			ICP            string `json:"icp"`
		}{
			Signature:      site.Signature,
			About:          site.About,
			AvatarPath:     site.AvatarPath,
			BgPicturePath:  site.BgPicturePath,
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
			c.Request.URL.Path = "/"
			assetHandler(c)
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

	// 2) 其它静态资源全部从编译进二进制的文件系统提供。
	r.GET("/assets/*any", assetHandler)
	r.GET("/static/*any", func(c *gin.Context) {
		if serveStaticSiteAsset(c) {
			return
		}
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/static")
		assetHandler(c)
	})
	r.GET("/", assetHandler)
	r.GET("/index.html", assetHandler)
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"msg": "not found"})
			return
		}
		if assetCache.Has(c.Request.URL.Path) {
			assetHandler(c)
			return
		}
		c.Request.URL.Path = "/"
		assetHandler(c)
	})

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

func serveStaticSiteAsset(c *gin.Context) bool {
	switch c.Request.URL.Path {
	case "/static/avatar.jpg":
		serveSiteAsset(c, config.Cfg.Site.AvatarPath, "avatar.jpg")
		return true
	case "/static/background.jpg":
		serveSiteAsset(c, config.Cfg.Site.BgPicturePath, "")
		return true
	default:
		return false
	}
}

func serveSiteAsset(c *gin.Context, configuredPath, fallbackPath string) {
	assetPath, err := resolveSiteAssetPath(config.Cfg.Server.DataPath, configuredPath)
	if (err != nil || !fileExists(assetPath)) && fallbackPath != "" {
		assetPath, err = resolveSiteAssetPath(config.Cfg.Server.DataPath, fallbackPath)
	}
	if err != nil || !fileExists(assetPath) {
		c.JSON(http.StatusNotFound, gin.H{"msg": "site asset not found"})
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.File(assetPath)
}

func resolveSiteAssetPath(dataRoot, configuredPath string) (string, error) {
	configuredPath = strings.TrimSpace(configuredPath)
	if configuredPath == "" || filepath.IsAbs(configuredPath) {
		return "", fmt.Errorf("site asset path must be relative to the data directory")
	}

	root, err := filepath.Abs(dataRoot)
	if err != nil {
		return "", err
	}
	assetPath, err := filepath.Abs(filepath.Join(root, configuredPath))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, assetPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("site asset path escapes the data directory")
	}
	return assetPath, nil
}
