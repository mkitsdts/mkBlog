package log

import (
	"log/slog"
	"mkBlog/config"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Init() {
	_, err := os.OpenFile("./app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	logger := &lumberjack.Logger{
		Filename:   "app.log", // 日志文件路径
		MaxSize:    100,       // 单个文件最大 100 MB
		MaxBackups: 5,         // 保留最多 5 个旧文件
		MaxAge:     30,        // 保留最多 30 天
		Compress:   true,      // 是否压缩旧日志（.gz）
	}

	var level slog.Level
	if config.Cfg.Server.Devmode {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(logger, &slog.HandlerOptions{
		Level: level,
	})

	// 设置全局 logger
	slog.SetDefault(slog.New(handler))
}
