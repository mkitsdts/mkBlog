package sslcontroler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SaveFile(path string, data *[]byte) error {
	// 校验路径
	if path == "" {
		return fmt.Errorf("empty path")
	}

	if after, ok := strings.CutPrefix(path, string(os.PathSeparator)); ok {
		path = after
	}

	// 确保父目录存在
	dir := filepath.Dir(path)
	if dir != "." && dir != string(os.PathSeparator) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}

	// 在同一目录创建临时文件，写入后原子替换目标文件
	tmpFile, err := os.CreateTemp(dir, ".tmpfile-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	// 确保临时文件被清理（失败时）
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(*data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// 原子替换目标文件
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("rename temp to target: %w", err)
	}

	// 设置文件权限为 0644，按需调整
	if err := os.Chmod(path, 0o644); err != nil {
		return fmt.Errorf("chmod target file: %w", err)
	}

	return nil
}
