package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"  // 注册 GIF 解码器
	_ "image/jpeg" // 注册 JPEG 解码器
	_ "image/png"  // 注册 PNG 解码器
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

func SaveImage(img *models.Image) error {
	// 目录/文件名清理
	title := filepath.Base(filepath.Clean(img.Title))
	base := config.Cfg.Server.ImageSavePath
	dir := filepath.Join(base, title)

	name := filepath.Clean(img.Name)
	if filepath.IsAbs(name) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid image name")
	}

	// 兼容 data URL 前缀
	dataStr := img.Data
	if i := strings.Index(dataStr, ","); i != -1 && strings.Contains(strings.ToLower(dataStr[:i]), "base64") {
		dataStr = dataStr[i+1:]
	}

	// base64 解码（容错 RawStdEncoding）
	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		if d2, err2 := base64.RawStdEncoding.DecodeString(dataStr); err2 == nil {
			data = d2
		} else {
			slog.Error("Failed to decode image base64", "error", err)
			return err
		}
	}

	ext := strings.ToLower(filepath.Ext(name))
	fp := filepath.Join(dir, name)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
		slog.Error("Failed to create dir", "error", err, "dir", filepath.Dir(fp))
		return err
	}

	if ext == ".webp" {
		// 已是 webp，原样落盘
		if err := os.WriteFile(fp, data, 0o644); err != nil {
			slog.Error("Failed to write webp file", "error", err, "path", fp)
			return err
		}
		slog.Info("Image saved (webp passthrough)", "path", fp)
		return nil
	}

	// 非 webp：先用标准库解码，再转 webp
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		slog.Error("Failed to decode source image", "error", err, "ext", ext)
		return err
	}

	// 改为 .webp
	fp = filepath.Join(dir, strings.TrimSuffix(name, ext)+".webp")
	f, err := os.Create(fp)
	if err != nil {
		slog.Error("Failed to create image file", "error", err, "path", fp)
		return err
	}
	defer f.Close()

	if err := webp.Encode(f, src, &webp.Options{Quality: 80}); err != nil {
		slog.Error("Failed to encode webp", "error", err, "path", fp)
		return err
	}

	slog.Info("Image saved (converted to webp)", "path", fp)
	return nil
}
