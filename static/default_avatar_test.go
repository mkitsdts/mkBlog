package static

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDefaultAvatarCreatesEmbeddedAvatar(t *testing.T) {
	dataPath := t.TempDir()

	if err := EnsureDefaultAvatar(dataPath); err != nil {
		t.Fatalf("ensure default avatar: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dataPath, "avatar.jpg"))
	if err != nil {
		t.Fatalf("read default avatar: %v", err)
	}
	if !bytes.Equal(got, defaultAvatar) {
		t.Fatal("written avatar does not match the embedded default")
	}
}

func TestEnsureDefaultAvatarPreservesUserAvatar(t *testing.T) {
	dataPath := t.TempDir()
	avatarPath := filepath.Join(dataPath, "avatar.jpg")
	userAvatar := []byte("user-selected-avatar")
	if err := os.WriteFile(avatarPath, userAvatar, 0o644); err != nil {
		t.Fatalf("write user avatar: %v", err)
	}

	if err := EnsureDefaultAvatar(dataPath); err != nil {
		t.Fatalf("ensure default avatar: %v", err)
	}

	got, err := os.ReadFile(avatarPath)
	if err != nil {
		t.Fatalf("read user avatar: %v", err)
	}
	if !bytes.Equal(got, userAvatar) {
		t.Fatal("existing user avatar was overwritten")
	}
}
