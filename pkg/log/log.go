package log

import (
	"log/slog"
	"mkBlog/config"
	"os"
)

func Init() {
	file, err := os.OpenFile("./app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}
	var level slog.Level
	if config.Cfg.Server.Devmode {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}
	logger := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(logger))
}
