package service

import (
	"crypto/tls"
	"log/slog"
	"mkBlog/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BlogService struct {
	DB     *gorm.DB
	Router *gin.Engine
	Cfg    *config.Config
}

func NewBlogService(db *gorm.DB, r *gin.Engine, cfg *config.Config) {
	service := &BlogService{
		DB:     db,
		Router: r,
		Cfg:    cfg,
	}

	api := r.Group("/api")
	{
		api.GET("/articles", service.GetArticleSummary)
		api.GET("/article/:title", service.GetArticleDetail)
		api.GET("/search", service.SearchArticle)
		api.GET("/categories", service.GetCategories)
		api.GET("/friends", service.GetFriendList)
		api.POST("/friends", service.ApplyFriend)
		if cfg.Auth.Enabled {
			api.PUT("/article/:title", service.AddArticle, service.Auth())
			api.PUT("/image", service.AddImage, service.Auth())
		} else {
			api.PUT("/article/:title", service.AddArticle)
			api.PUT("/image", service.AddImage)
		}
	}

	if cfg.TLS.Enabled {
		srv := &http.Server{
			Addr:    ":8080",
			Handler: r,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.X25519, tls.CurveP256,
				},
			},
		}
		// Start HTTPS server
		if err := srv.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key); err != nil {
			slog.Error("failed to start HTTPS server", "error", err)
		}
	} else {
		if err := r.Run(":8080"); err != nil {
			slog.Error("failed to start HTTP server", "error", err)
		}
	}

}

func (s *BlogService) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "Bearer "+s.Cfg.Auth.Secret {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
