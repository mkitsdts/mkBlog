package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type BlogService struct {
}

const shutdownTimeout = 30 * time.Second

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
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancelServices := context.WithCancel(signalCtx)
	defer cancelServices()

	var wg sync.WaitGroup
	errCh := make(chan error, 8)
	shutdownFns := make([]func(context.Context) error, 0, 4)

	startService := func(name string, serve func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := serve(); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
				errCh <- fmt.Errorf("%s: %w", name, err)
			}
		}()
	}

	addShutdown := func(name string, shutdown func(context.Context) error) {
		shutdownFns = append(shutdownFns, func(ctx context.Context) error {
			if err := shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
				return fmt.Errorf("%s: %w", name, err)
			}
			return nil
		})
	}

	if config.Cfg.Server.Devmode {
		pprofSrv := &http.Server{
			Addr:    ":6060",
			Handler: http.DefaultServeMux,
		}
		slog.Info("starting pprof server", "addr", pprofSrv.Addr)
		startService("pprof server", pprofSrv.ListenAndServe)
		addShutdown("shutdown pprof server", pprofSrv.Shutdown)
	}

	addr := ":" + fmt.Sprint(config.Cfg.Server.Port)
	if config.Cfg.Server.HTTP3Enabled || config.Cfg.TLS.Enabled {
		if err := tlscert.LoadCert(); err != nil {
			return fmt.Errorf("load TLS cert: %w", err)
		}
	}

	if config.Cfg.Server.HTTP3Enabled {
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			slog.Warn("HTTP3 disabled because UDP address is unavailable", "addr", addr, "error", err)
		} else {
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
			startService("HTTP3 server", func() error {
				defer func() {
					if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
						slog.Warn("failed to close HTTP3 packet connection", "error", err)
					}
				}()
				return srv.Serve(conn)
			})
			addShutdown("shutdown HTTP3 server", func(ctx context.Context) error {
				err := srv.Shutdown(ctx)
				if closeErr := conn.Close(); closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
					err = errors.Join(err, closeErr)
				}
				return err
			})
		}
	}

	if config.Cfg.CertControl.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tlscert.StartContext(ctx)
		}()
	}

	// HTTP3 和 HTTP2 + TLS 是可以同时开启的， UDP 和 TCP 不冲突
	var tcpSrv *http.Server
	if config.Cfg.TLS.Enabled {
		tcpSrv = &http.Server{
			Addr:    addr,
			Handler: router.GetRouter(),
			TLSConfig: &tls.Config{
				MinVersion:     tls.VersionTLS12,
				GetCertificate: tlscert.GetCurrentCert,
			},
		}
		slog.Info("starting HTTPS server", "port", config.Cfg.Server.Port)
		startService("HTTPS server", func() error {
			return tcpSrv.ListenAndServeTLS("", "")
		})
	} else {
		tcpSrv = &http.Server{
			Addr:    addr,
			Handler: router.GetRouter(),
		}
		slog.Info("starting HTTP server", "port", config.Cfg.Server.Port)
		startService("HTTP server", tcpSrv.ListenAndServe)
	}
	addShutdown("shutdown TCP server", tcpSrv.Shutdown)

	var runErr error
	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		runErr = err
		slog.Error("service failed, shutting down", "error", err)
		cancelServices()
		stop()
	}
	cancelServices()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	var shutdownErr error
	for i := len(shutdownFns) - 1; i >= 0; i-- {
		if err := shutdownFns[i](shutdownCtx); err != nil {
			slog.Error("failed to shutdown service", "error", err)
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("all services stopped")
	case <-shutdownCtx.Done():
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown timeout after %s: %w", shutdownTimeout, shutdownCtx.Err()))
	}

	if runErr != nil || shutdownErr != nil {
		return errors.Join(runErr, shutdownErr)
	}
	return nil
}
