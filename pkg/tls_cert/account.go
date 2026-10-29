package tlscert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mkBlog/config"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v4/registration"
)

// ACME 账号相关文件统一放在 <data>/acme 下。
//
// 账号必须落盘复用：Let's Encrypt 对“新账号注册”是按 IP 计数的
// （每个 IP 每 3 小时 10 个），每次启动都注册新账号会被直接限流封禁。
const (
	acmeDirName         = "acme"
	accountKeyFileName  = "account.key"
	accountMetaFileName = "account.json"
)

// accountMeta 保存除私钥以外的账号信息，用于重启后直接拿到 kid。
type accountMeta struct {
	Email           string `json:"email"`
	RegistrationURI string `json:"registration_uri"`
}

func acmeDir() string {
	return filepath.Join(config.Cfg.Server.DataPath, acmeDirName)
}

func accountKeyPath() string {
	return filepath.Join(acmeDir(), accountKeyFileName)
}

func accountMetaPath() string {
	return filepath.Join(acmeDir(), accountMetaFileName)
}

// loadAccount 读取已落盘的 ACME 账号。
// 返回的错误满足 errors.Is(err, fs.ErrNotExist) 时表示还没有账号，需要注册。
func loadAccount() (accountMeta, crypto.PrivateKey, error) {
	keyPEM, err := os.ReadFile(accountKeyPath())
	if err != nil {
		return accountMeta{}, nil, err
	}

	key, err := parseAccountKey(keyPEM)
	if err != nil {
		return accountMeta{}, nil, fmt.Errorf("parse %s: %w", accountKeyPath(), err)
	}

	raw, err := os.ReadFile(accountMetaPath())
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// 只有私钥也能恢复账号：用同一个 key 重新 newAccount 会返回已存在的账号，
		// 这里先返回空 meta，由 ensureRegistration 补全并落盘。
		return accountMeta{}, key, nil
	case err != nil:
		return accountMeta{}, nil, fmt.Errorf("read %s: %w", accountMetaPath(), err)
	}

	var meta accountMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return accountMeta{}, nil, fmt.Errorf("parse %s: %w", accountMetaPath(), err)
	}
	return meta, key, nil
}

func parseAccountKey(keyPEM []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}

func newAccountKey() (crypto.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// saveAccount 落盘账号私钥与账号信息，保证下次启动复用同一个账号。
func saveAccount(meta accountMeta, key crypto.PrivateKey) error {
	keyDER, err := marshalAccountKey(key)
	if err != nil {
		return err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := writeFileAtomic(accountKeyPath(), keyPEM, 0600); err != nil {
		return fmt.Errorf("save ACME account key: %w", err)
	}

	raw, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(accountMetaPath(), append(raw, '\n'), 0600); err != nil {
		return fmt.Errorf("save ACME account metadata: %w", err)
	}
	return nil
}

func marshalAccountKey(key crypto.PrivateKey) ([]byte, error) {
	if ecKey, ok := key.(*ecdsa.PrivateKey); ok {
		return x509.MarshalECPrivateKey(ecKey)
	}
	return x509.MarshalPKCS8PrivateKey(key)
}

// ensureRegistration 保证 ACME 账号已注册并把账号信息落盘。
func ensureRegistration() error {
	if leuser.Registration != nil && leuser.Registration.URI != "" {
		return nil
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return fmt.Errorf("register ACME account: %w", err)
	}
	leuser.Registration = reg
	slog.Info("registered new ACME account", "email", leuser.Email, "uri", reg.URI)

	if err := saveAccount(accountMeta{Email: leuser.Email, RegistrationURI: reg.URI}, leuser.key); err != nil {
		return err
	}
	return nil
}
