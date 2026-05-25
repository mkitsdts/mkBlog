package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"mkBlog/config"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	tlscert "mkBlog/pkg/tls_cert"
	"mkBlog/service/api"
	"net"
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
	api.Init()

	if err := os.MkdirAll(config.Cfg.Server.ImageSavePath, 0755); err != nil {
		slog.Error("failed to create image save path", "error", err)
		return nil, err
	}

	a := router.GetRouter().Group("/api")
	if config.Cfg.Server.Devmode {
		{
			a.GET("/articles", api.GetArticleSummary)
			a.GET("/allarticles", api.GetAllArticleSummaries)
			a.GET("/article/:title", api.GetArticleDetail)
			a.GET("/search", api.SearchArticle)
			a.GET("/categories", api.GetCategories)
			a.GET("/friends", api.GetFriendList)
			a.POST("/friends", api.ApplyFriend)

			a.GET("/comments", api.GetComments)
			a.POST("/comments", api.AddComment)

			a.PUT("/article/:title", api.UploadArticle)
			a.PUT("/image", api.UploadImage)
			a.DELETE("/article/:title", api.DeleteArticle)
			a.POST("/blockip", api.BlackIP)
		}
	} else {
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

func (s *BlogService) Start() error {
	if config.Cfg.Server.Devmode {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}
	addr := ":" + fmt.Sprint(config.Cfg.Server.Port)
	if config.Cfg.Server.HTTP3Enabled {
		if err := tlscert.LoadCert(); err != nil {
			return fmt.Errorf("load TLS cert: %w", err)
		}
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			return fmt.Errorf("start HTTP3 server: %w", err)
		}
		srv := http3.Server{
			Handler: router.GetRouter(),
			Addr:    addr,
			TLSConfig: http3.ConfigureTLSConfig(&tls.Config{
				MinVersion:     tls.VersionTLS13,
				GetCertificate: tlscert.GetCurrentCert,
			}),
			QUICConfig: &quic.Config{},
		}
		slog.Info("starting HTTP3 server", "port", config.Cfg.Server.Port)
		go func() {
			if err := srv.Serve(conn); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start HTTP3 server", "error", err)
			}
		}()
	}
	if config.Cfg.CertControl.Enabled {
		if err := tlscert.Init(); err != nil {
			return fmt.Errorf("init cert control: %w", err)
		}
		go tlscert.Start()
	}
	// HTTP3 和 HTTP2 + TLS 是可以同时开启的， UDP 和 TCP 不冲突
	if config.Cfg.TLS.Enabled {
		if err := tlscert.LoadCert(); err != nil {
			return fmt.Errorf("load TLS cert: %w", err)
		}
		srv := &http.Server{
			Addr:    addr,
			Handler: router.GetRouter(),
			TLSConfig: &tls.Config{
				MinVersion:     tls.VersionTLS12,
				GetCertificate: tlscert.GetCurrentCert,
			},
		}
		// Start HTTPS server
		if err := srv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start HTTPS server", "error", err)
			return fmt.Errorf("start HTTPS server: %w", err)
		}
	} else {
		if err := router.GetRouter().Run(addr); err != nil {
			slog.Error("failed to start HTTP server", "error", err)
			return fmt.Errorf("start HTTP server: %w", err)
		}
	}
	return nil
}
