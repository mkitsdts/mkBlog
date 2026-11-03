package tlscert

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"mkBlog/config"
)

func TestAccountPersistence(t *testing.T) {
	dataDir := t.TempDir()
	originalDataPath := config.Cfg.Server.DataPath
	config.Cfg.Server.DataPath = dataDir
	t.Cleanup(func() { config.Cfg.Server.DataPath = originalDataPath })

	if _, _, err := loadAccount(); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist before saving, got %v", err)
	}

	key, err := newAccountKey()
	if err != nil {
		t.Fatalf("generate account key: %v", err)
	}
	want := accountMeta{Email: "ops@example.com", RegistrationURI: "https://acme.example/acct/1"}
	if err := saveAccount(want, key); err != nil {
		t.Fatalf("save account: %v", err)
	}

	got, loadedKey, err := loadAccount()
	if err != nil {
		t.Fatalf("load account: %v", err)
	}
	if got != want {
		t.Fatalf("meta = %+v, want %+v", got, want)
	}
	if _, err := marshalAccountKey(loadedKey); err != nil {
		t.Fatalf("loaded key is unusable: %v", err)
	}

	// 账号信息丢了也要能用同一把私钥找回账号，不能直接当成“没有账号”去注册新的。
	if err := os.Remove(accountMetaPath()); err != nil {
		t.Fatalf("remove account metadata: %v", err)
	}
	meta, recoveredKey, err := loadAccount()
	if err != nil {
		t.Fatalf("load account without metadata: %v", err)
	}
	if meta.RegistrationURI != "" {
		t.Fatalf("expected empty registration URI, got %q", meta.RegistrationURI)
	}
	if _, err := marshalAccountKey(recoveredKey); err != nil {
		t.Fatalf("recovered key is unusable: %v", err)
	}
}
