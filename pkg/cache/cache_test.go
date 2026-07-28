package cache

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuildAssetCache(t *testing.T) {
	files := fstest.MapFS{
		"index.html": {
			Data: []byte("<!doctype html><title>mkBlog</title>"),
		},
		"assets/app.js": {
			Data: []byte("console.log('mkBlog')"),
		},
		"asset.bin": {
			Data: []byte{0x00, 0x01, 0x02},
		},
	}

	if err := BuildAssetCache(files); err != nil {
		t.Fatalf("build cache: %v", err)
	}

	index := globalAssetCache.Get("/")
	if index == nil {
		t.Fatal("expected / to resolve to index.html")
	}
	if got := string(index.raw); !strings.Contains(got, "mkBlog") {
		t.Fatalf("index content = %q", got)
	}
	if !globalAssetCache.Has("/assets/app.js") {
		t.Fatal("expected JavaScript asset in cache")
	}
	if !globalAssetCache.Has("/asset.bin") {
		t.Fatal("all files in the frontend dist should be cached")
	}
}

func TestBuildAssetCacheRequiresIndex(t *testing.T) {
	files := fstest.MapFS{
		"assets/app.js": {
			Data: []byte("console.log('mkBlog')"),
		},
	}

	err := BuildAssetCache(files)
	if err == nil || !strings.Contains(err.Error(), "static/dist/index.html is missing") {
		t.Fatalf("expected missing index error, got %v", err)
	}
}
