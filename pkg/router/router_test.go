package router

import (
	"mkBlog/config"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResolveSiteAssetPath(t *testing.T) {
	dataRoot := t.TempDir()

	got, err := resolveSiteAssetPath(dataRoot, "images/avatar.jpg")
	if err != nil {
		t.Fatalf("resolve relative site asset: %v", err)
	}
	want := filepath.Join(dataRoot, "images", "avatar.jpg")
	if got != want {
		t.Fatalf("resolved path = %q, want %q", got, want)
	}
}

func TestResolveSiteAssetPathRejectsUnsafePaths(t *testing.T) {
	dataRoot := t.TempDir()
	tests := []string{"", "../avatar.jpg", filepath.Join(dataRoot, "avatar.jpg")}

	for _, configuredPath := range tests {
		t.Run(configuredPath, func(t *testing.T) {
			if _, err := resolveSiteAssetPath(dataRoot, configuredPath); err == nil {
				t.Fatalf("resolveSiteAssetPath(%q) succeeded, want error", configuredPath)
			}
		})
	}
}

func TestStaticAvatarUsesConfiguredDataFile(t *testing.T) {
	dataRoot := t.TempDir()
	firstAvatar := []byte("first avatar")
	secondAvatar := []byte("second avatar")
	if err := os.WriteFile(filepath.Join(dataRoot, "first.jpg"), firstAvatar, 0o644); err != nil {
		t.Fatalf("write first avatar: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataRoot, "second.jpg"), secondAvatar, 0o644); err != nil {
		t.Fatalf("write second avatar: %v", err)
	}

	oldDataPath := config.Cfg.Server.DataPath
	oldAvatarPath := config.Cfg.Site.AvatarPath
	defer func() {
		config.Cfg.Server.DataPath = oldDataPath
		config.Cfg.Site.AvatarPath = oldAvatarPath
	}()
	config.Cfg.Server.DataPath = dataRoot

	for _, test := range []struct {
		name       string
		avatarPath string
		want       []byte
	}{
		{name: "first", avatarPath: "first.jpg", want: firstAvatar},
		{name: "second", avatarPath: "second.jpg", want: secondAvatar},
	} {
		t.Run(test.name, func(t *testing.T) {
			config.Cfg.Site.AvatarPath = test.avatarPath
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest("GET", "/static/avatar.jpg", nil)

			if !serveStaticSiteAsset(context) {
				t.Fatal("avatar request was not handled as a site asset")
			}
			if recorder.Code != 200 {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}
			if got := recorder.Body.Bytes(); string(got) != string(test.want) {
				t.Fatalf("body = %q, want %q", got, test.want)
			}
			if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
				t.Fatalf("Cache-Control = %q, want no-cache", got)
			}
		})
	}
}
