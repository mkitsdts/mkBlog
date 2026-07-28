package cache

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"path"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
)

type cachedAsset struct {
	raw         []byte
	gzipData    []byte
	brData      []byte
	etag        string
	contentType string
}

type AssetCache struct {
	items map[string]*cachedAsset // key: URL 路径 (/assets/xxx.css 或 /index.html)
}

var globalAssetCache *AssetCache

func Init(root fs.FS) error {
	return BuildAssetCache(root)
}

func GetGlobalAssetCache() *AssetCache {
	return globalAssetCache
}

func BuildAssetCache(root fs.FS) error {
	assetCache := &AssetCache{items: make(map[string]*cachedAsset)}
	addFile := func(filePath, webPath string) error {
		ext := strings.ToLower(path.Ext(filePath))
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "application/octet-stream"
		}
		data, err := fs.ReadFile(root, filePath)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		etag := `"` + hex.EncodeToString(sum[:8]) + `"` // 截断 8 bytes 足够
		ca := &cachedAsset{
			raw:         data,
			etag:        etag,
			contentType: ct,
		}
		if isText(ct) {
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
		}

		assetCache.items[webPath] = ca
		return nil
	}

	// 扫描嵌入式文件系统。
	err := fs.WalkDir(root, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return addFile(filePath, "/"+filePath)
	})
	if err != nil {
		return fmt.Errorf("build embedded asset cache: %w", err)
	}
	if assetCache.items["/index.html"] == nil {
		return fmt.Errorf("build embedded asset cache: static/dist/index.html is missing")
	}
	globalAssetCache = assetCache
	slog.Info("embedded asset cache built", "count", len(globalAssetCache.items))
	return nil
}

func (ac *AssetCache) Get(path string) *cachedAsset {
	if path == "/" {
		if v := ac.items["/index.html"]; v != nil {
			return v
		}
	}
	return ac.items[path]
}

func (ac *AssetCache) Has(path string) bool {
	return ac.Get(path) != nil
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
		if strings.HasPrefix(p, "/assets/") && strings.Contains(p, ".") {
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
