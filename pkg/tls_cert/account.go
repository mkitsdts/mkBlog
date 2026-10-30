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

// ACME 账号相关文件统一放在 <data>/acme 下
const (
	acmeDirName         = "acme"
	accountKeyFileName  = "account.key"
	accountMetaFileName = "account.json"
)

// 保存除私钥以外的账号信息，重启后可以直接拿到 kid
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

// 加载乐此加密的账号，返回 fs.ErrNotExist 时表示需要注册账号
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
		// 返回私钥，后续可能恢复账号
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
