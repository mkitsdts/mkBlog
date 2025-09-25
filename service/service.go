package service

import (
	"crypto/tls"
	"log/slog"
	"mkBlog/config"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	"mkBlog/service/api"
	"net/http"
	"os"
)

type BlogService struct {
}

func NewBlogService() (*BlogService, error) {
	var service BlogService

	if err := os.MkdirAll(config.Cfg.Server.ImageSavePath, 0755); err != nil {
		slog.Error("failed to create image save path", "error", err)
		return nil, err
	}

	a := router.GetRouter().Group("/api")
	{
		a.GET("/articles", api.GetArticleSummary)
		a.GET("/article/:title", api.GetArticleDetail) // 限流，防止爆破
		a.GET("/search", api.SearchArticle)
		a.GET("/categories", api.GetCategories)
		a.GET("/friends", api.GetFriendList)
		a.POST("/friends", api.ApplyFriend)

		a.GET("/comments", api.GetComments)
		a.POST("/comments", api.AddComment)

		if config.Cfg.Auth.Enabled {
			a.PUT("/article/:title", api.UploadArticle, middleware.AuthRequired())
			a.PUT("/image", api.UploadImage, middleware.AuthRequired())
			a.DELETE("/article/:title", api.DeleteArticle, middleware.AuthRequired())
		} else {
			a.PUT("/article/:title", api.UploadArticle)
			a.PUT("/image", api.UploadImage)
			a.DELETE("/article/:title", api.DeleteArticle)
		}
	}
	return &service, nil
}

func (s *BlogService) Start() {
	if config.Cfg.TLS.Enabled {
		srv := &http.Server{
			Addr:    ":8080",
			Handler: router.GetRouter(),
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.X25519, tls.CurveP256,
				},
			},
		}
		// Start HTTPS server
		if err := srv.ListenAndServeTLS(config.Cfg.TLS.Cert, config.Cfg.TLS.Key); err != nil {
			slog.Error("failed to start HTTPS server", "error", err)
		}
	} else {
		if err := router.GetRouter().Run(":8080"); err != nil {
			slog.Error("failed to start HTTP server", "error", err)
		}
	}
}
