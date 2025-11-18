package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"mkBlog/config"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	"mkBlog/service/api"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
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
		a.GET("/articles", api.GetArticleSummary, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.GET("/allarticles", api.GetAllArticleSummaries, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.GET("/article/:title", api.GetArticleDetail, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.GET("/search", api.SearchArticle, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.GET("/categories", api.GetCategories, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.GET("/friends", api.GetFriendList, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.POST("/friends", api.ApplyFriend, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))

		a.GET("/comments", api.GetComments, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))
		a.POST("/comments", api.AddComment, middleware.RateLimit(config.Cfg.Server.Limiter.Requests, config.Cfg.Server.Limiter.Duration))

		if config.Cfg.Auth.Enabled {
			a.PUT("/article/:title", api.UploadArticle, middleware.AuthRequired())
			a.PUT("/image", api.UploadImage, middleware.AuthRequired())
			a.DELETE("/article/:title", api.DeleteArticle, middleware.AuthRequired())
			a.POST("/blockip", api.BlackIP, middleware.AuthRequired())
		} else {
			a.PUT("/article/:title", api.UploadArticle)
			a.PUT("/image", api.UploadImage)
			a.DELETE("/article/:title", api.DeleteArticle)
			a.POST("/blockip", api.BlackIP)
		}
	}
	return &service, nil
}

func (s *BlogService) Start() {
	if config.Cfg.Server.Devmode {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}
	if config.Cfg.Server.HTTP3Enabled {
		cert, err := tls.LoadX509KeyPair(config.Cfg.TLS.Cert, config.Cfg.TLS.Key)
		if err != nil {
			slog.Error("failed to load TLS certificate for HTTP3", "error", err)
			return
		}
		srv := http3.Server{
			Handler: router.GetRouter(),
			Addr:    ":" + fmt.Sprint(config.Cfg.Server.Port),
			TLSConfig: http3.ConfigureTLSConfig(&tls.Config{
				MinVersion:   tls.VersionTLS13,
				Certificates: []tls.Certificate{cert},
			}),
			QUICConfig: &quic.Config{},
		}
		slog.Info("starting HTTP3 server", "port", config.Cfg.Server.Port)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				slog.Error("failed to start HTTP3 server", "error", err)
			}
		}()
	}
	if config.Cfg.TLS.Enabled {
		srv := &http.Server{
			Addr:    ":" + fmt.Sprint(config.Cfg.Server.Port),
			Handler: router.GetRouter(),
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		// Start HTTPS server
		if err := srv.ListenAndServeTLS(config.Cfg.TLS.Cert, config.Cfg.TLS.Key); err != nil {
			slog.Error("failed to start HTTPS server", "error", err)
		}
	} else {
		if err := router.GetRouter().Run(":" + fmt.Sprint(config.Cfg.Server.Port)); err != nil {
			slog.Error("failed to start HTTP server", "error", err)
		}
	}
}
