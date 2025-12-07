package cache

import (
	"testing"
	"time"
)

func TestCacheBuild(t *testing.T) {
	BuildAssetCache("../../static")
	t.Log("Cache build succeeded")
	asset := globalAssetCache.Get("/index.html")
	if asset == nil {
		t.Error("Expected /index.html to be in the asset cache")
		return
	}
	t.Log("Asset /index.html found in cache")
	if asset.raw != nil {
		t.Log(string(asset.raw))
	} else {
		t.Error("Expected raw data for /index.html to be non-nil")
	}
	t.Log(asset.etag)
}

func TestWatchCacheFiles(t *testing.T) {
	Init("../../static")
	go func() {
		asset := globalAssetCache.Get("/index.html")
		if asset == nil {
			panic("failed to get index.html")
		}
		if asset.raw != nil {
			println(string(asset.raw))
		}
		time.Sleep(10 * time.Second)
		asset = globalAssetCache.Get("/index.html")
		if asset == nil {
			t.Log("Successfully remove file")
		} else {
			t.Error("Failed to remove file")
		}
	}()
	time.Sleep(100 * time.Minute)
}
