package log

import (
	"log/slog"
	"mkBlog/models"
	"os"
	"path"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(flag bool) {
	logPath := path.Join(models.Default_Data_Path, models.Default_Log_File_Path)
	_, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	logger := &lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    100,     // 单个文件最大 100 MB
		MaxBackups: 5,       // 保留最多 5 个旧文件
		MaxAge:     30,      // 保留最多 30 天
		Compress:   true,    // 是否压缩旧日志（.gz）
	}

	var level slog.Level
	if flag {
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
