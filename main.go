package main

import (
	"flag"
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"mkBlog/pkg/bloom"
	"mkBlog/pkg/cache"
	applog "mkBlog/pkg/log"
	"mkBlog/pkg/middleware"
	"mkBlog/pkg/router"
	"mkBlog/service"
	"os"
)

func Init(debugflag *bool) {
	if err := os.MkdirAll(models.Default_Data_Path, 0755); err != nil {
		return
	}
	if debugflag != nil {
		applog.Init(*debugflag)
	} else {
		applog.Init(false)
	}

	config.Init()

	bloom.Init()
	cache.Init(models.Default_Static_File_Path)
	middleware.Init()
}

func main() {
	debug := flag.Bool("debug", false, "启用调试模式")

	// 解析命令行参数
	flag.Parse()
	Init(debug)

	if err := router.InitRouter(); err != nil {
		panic("failed to create router: " + err.Error())
	}

	s, err := service.NewBlogService()
	if err != nil {
		slog.Error("failed to initialize blog service", "error", err)
		os.Exit(1)
	}
	if err := s.Start(); err != nil {
		slog.Error("blog service stopped", "error", err)
		os.Exit(1)
	}
}
