package utils

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"os"
	"path/filepath"
	"strings"
)

func SaveImage(img *models.Image) error {
	// 清理并约束 title 作为子目录，防止路径逃逸
	titleDir := filepath.Base(filepath.Clean(img.Title))
	base := config.Cfg.Server.ImageSavePath
	dir := filepath.Join(base, titleDir)

	// 规范化文件名，拒绝绝对路径和上跳
	name := filepath.Clean(img.Name)
	if strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
		return fmt.Errorf("invalid image name")
	}
	fp := filepath.Join(dir, name)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
		slog.Error("Failed to create dir", "error", err, "dir", filepath.Dir(fp))
		return err
	}

	// 兼容 data URL 前缀
	dataStr := img.Data
	if i := strings.Index(dataStr, ","); i != -1 && strings.Contains(strings.ToLower(dataStr[:i]), "base64") {
		dataStr = dataStr[i+1:]
	}

	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		slog.Error("Failed to decode image data", "error", err)
		return err
	}
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		slog.Error("Failed to write image file", "error", err, "path", fp)
		return err
	}
	slog.Info("Image saved", "path", fp)
	return nil
}
