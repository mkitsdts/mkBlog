package service

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

// containsCJK 粗略判断查询字符串中是否包含中日韩统一表意文字（CJK Unified Ideographs）
func containsCJK(s string) bool {
	for _, r := range s {
		// 常用中日韩汉字范围：
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
			(r >= 0x20000 && r <= 0x2A6DF) || // Extension B
			(r >= 0x2A700 && r <= 0x2B73F) || // Extension C
			(r >= 0x2B740 && r <= 0x2B81F) || // Extension D
			(r >= 0x2B820 && r <= 0x2CEAF) || // Extension E
			(r >= 0xF900 && r <= 0xFAFF) { // CJK Compatibility Ideographs
			return true
		}
	}
	return false
}

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
