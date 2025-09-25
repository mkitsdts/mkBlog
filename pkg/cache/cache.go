package cache

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

type cachedAsset struct {
	path        string
	raw         []byte
	gzipData    []byte
	brData      []byte
	etag        string
	modTime     time.Time
	contentType string
}

type AssetCache struct {
	items map[string]*cachedAsset // key: URL 路径 (/assets/xxx.css 或 /index.html)
}

func BuildAssetCache(root string) (*AssetCache, error) {
	ac := &AssetCache{items: make(map[string]*cachedAsset)}
	// 允许的扩展名集合
	allow := map[string]string{
		".html": "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".json": "application/json; charset=utf-8",
		".ico":  "image/x-icon",
		".svg":  "image/svg+xml",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".webp": "image/webp",
	}

	addFile := func(absPath, webPath string, info fs.FileInfo) error {
		ext := strings.ToLower(filepath.Ext(absPath))
		ct, ok := allow[ext]
		if !ok {
			return nil
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		etag := `"` + hex.EncodeToString(sum[:8]) + `"` // 截断 8 bytes 足够
		ca := &cachedAsset{
			path:        webPath,
			raw:         data,
			etag:        etag,
			modTime:     info.ModTime(),
			contentType: ct,
		}
		// gzip
		var gzBuf strings.Builder
		gzWriter, _ := gzip.NewWriterLevel(&gzBuf, gzip.BestCompression)
		_, _ = gzWriter.Write(data)
		_ = gzWriter.Close()
		ca.gzipData = []byte(gzBuf.String())

		var brBuf strings.Builder
		brWriter := brotli.NewWriterLevel(&brBuf, brotli.BestCompression)
		_, _ = brWriter.Write(data)
		_ = brWriter.Close()
		ca.brData = []byte(brBuf.String())

		ac.items[webPath] = ca
		return nil
	}

	// 扫描 root
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, _ := d.Info()
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		webPath := "/" + rel
		// assets 下的文件保持 /assets/... 前缀
		if strings.HasPrefix(rel, "assets/") {
			webPath = "/" + rel
		}
		// index.html 允许直接 /index.html
		return addFile(path, webPath, info)
	})
	if err != nil {
		return nil, err
	}

	return ac, nil
}

func (ac *AssetCache) Get(path string) *cachedAsset {
	if path == "/" {
		if v := ac.items["/index.html"]; v != nil {
			return v
		}
	}
	return ac.items[path]
}

func (ac *AssetCache) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		asset := ac.Get(p)
		if asset == nil {
			c.Status(404)
			return
		}

		// ETag / If-None-Match
		if inm := c.GetHeader("If-None-Match"); inm != "" && inm == asset.etag {
			c.Status(304)
			return
		}

		c.Header("ETag", asset.etag)
		c.Header("Content-Type", asset.contentType)

		// 缓存策略：带 hash 的文件（Vite 产物路径通常 /assets/xxxxx.[hash].css）
		if strings.HasPrefix(p, "/assets/") && strings.Contains(p, ".") && strings.Contains(p, ".") {
			c.Header("Cache-Control", "public,max-age=31536000,immutable")
		} else if p == "/index.html" || p == "/" {
			c.Header("Cache-Control", "no-cache")
		}

		ae := c.GetHeader("Accept-Encoding")
		if strings.Contains(ae, "br") && len(asset.brData) > 0 && isText(asset.contentType) {
			c.Header("Content-Encoding", "br")
			c.Writer.WriteHeader(200)
			_, _ = c.Writer.Write(asset.brData)
			return
		}
		if strings.Contains(ae, "gzip") && len(asset.gzipData) > 0 && isText(asset.contentType) {
			c.Header("Content-Encoding", "gzip")
			c.Writer.WriteHeader(200)
			_, _ = c.Writer.Write(asset.gzipData)
			return
		}
		// 原始
		c.Writer.WriteHeader(200)
		_, _ = c.Writer.Write(asset.raw)
	}
}

func isText(ct string) bool {
	return strings.HasPrefix(ct, "text/") || strings.Contains(ct, "javascript") || strings.Contains(ct, "json") || strings.Contains(ct, "svg")
}
