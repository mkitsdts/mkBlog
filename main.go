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
	staticfiles "mkBlog/static"
	"os"
)

func Init(debugflag *bool, serverFlags config.ServerConfig) {
	if err := os.MkdirAll(config.Cfg.Server.DataPath, 0755); err != nil {
		return
	}
	if err := staticfiles.EnsureDefaultAvatar(config.Cfg.Server.DataPath); err != nil {
		slog.Warn("failed to initialize default avatar", "error", err)
	}
	if debugflag != nil {
		applog.Init(*debugflag)
	} else {
		applog.Init(false)
	}

	config.Init()
	applyServerFlagOverrides(serverFlags)
	config.Finalize()

	bloom.Init()
	if err := cache.Init(staticfiles.Files()); err != nil {
		panic("failed to initialize embedded static files: " + err.Error())
	}
	middleware.Init()
}

func applyServerFlagOverrides(values config.ServerConfig) {
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "data_path":
			config.Cfg.Server.DataPath = values.DataPath
		case "log_file_path":
			config.Cfg.Server.LogFilePath = values.LogFilePath
		case "image_save_path":
			config.Cfg.Server.ImageSavePath = values.ImageSavePath
		case "config_file_path":
			config.Cfg.Server.ConfigFilePath = values.ConfigFilePath
		case "data_file_path":
			config.Cfg.Server.DataFilePath = values.DataFilePath
		case "server_port":
			config.Cfg.Server.Port = values.Port
		case "limiter_duration":
			config.Cfg.Server.Limiter.Duration = values.Limiter.Duration
		case "limiter_requests":
			config.Cfg.Server.Limiter.Requests = values.Limiter.Requests
		}
	})
}

func main() {
	debug := flag.Bool("debug", false, "启用调试模式")
	serverFlags := config.Cfg.Server
	flag.StringVar(&serverFlags.DataPath, "data_path", models.Default_Data_Path, "数据文件目录")
	flag.StringVar(&serverFlags.LogFilePath, "log_file_path", models.Default_Log_File_Path, "日志文件路径")
	flag.StringVar(&serverFlags.ImageSavePath, "image_save_path", models.Default_Image_Save_Path, "图片保存路径")
	flag.StringVar(&serverFlags.ConfigFilePath, "config_file_path", models.Default_Config_File_Path, "配置文件路径")
	flag.StringVar(&serverFlags.DataFilePath, "data_file_path", models.Default_Data_File_Path, "数据文件路径")
	flag.IntVar(&serverFlags.Port, "server_port", models.Default_Server_Port, "服务端口")
	flag.IntVar(&serverFlags.Limiter.Duration, "limiter_duration", models.Default_Limiter_Duartion, "限流时间窗口")
	flag.IntVar(&serverFlags.Limiter.Requests, "limiter_requests", models.Default_Limiter_Requests, "限流请求数")

	// 解析命令行参数
	flag.Parse()
	config.Cfg.Server = serverFlags
	Init(debug, serverFlags)

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
