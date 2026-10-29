package tlscert

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
)

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u *MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

var client *lego.Client
var leuser MyUser

const (
	renewalCheckInterval = 12 * time.Hour
	renewalRetryInitial  = 6 * time.Hour
	renewalRetryMax      = 24 * time.Hour

	// 默认 30s 对 DNS-01 校验来说太短，超时后重试等于多下一单。
	certificateTimeout = 5 * time.Minute
)

// Init 初始化 ACME 客户端。账号走落盘复用，只有确实没有账号时才注册，
// 避免每次启动都消耗 Let's Encrypt 的“新账号”配额（按 IP 计算）。
// 初始化失败时不会留下半初始化的 client，下一次 ensureClient 会重新初始化。
func Init() error {
	if err := prepareAccount(); err != nil {
		return err
	}

	newconfig := lego.NewConfig(&leuser)
	newconfig.Certificate.KeyType = certcrypto.RSA2048
	newconfig.Certificate.Timeout = certificateTimeout

	newClient, err := lego.NewClient(newconfig)
	if err != nil {
		return fmt.Errorf("create ACME client: %w", err)
	}

	provider, err := newChallengeProvider()
	if err != nil {
		return err
	}
	if err := newClient.Challenge.SetDNS01Provider(provider); err != nil {
		return fmt.Errorf("set dns01 provider: %w", err)
	}

	client = newClient
	if err := ensureRegistration(); err != nil {
		client = nil
		return err
	}
	return nil
}

// prepareAccount 复用落盘的 ACME 账号，没有账号时才生成新私钥。
func prepareAccount() error {
	meta, key, err := loadAccount()
	switch {
	case err == nil:
		leuser = MyUser{Email: meta.Email, key: key}
		if meta.RegistrationURI != "" {
			leuser.Registration = &registration.Resource{URI: meta.RegistrationURI}
		}
		if meta.Email != config.Cfg.CertControl.Email {
			slog.Warn("stored ACME account email differs from configuration, keeping the stored account",
				"account_email", meta.Email, "config_email", config.Cfg.CertControl.Email)
		}
		slog.Info("loaded existing ACME account", "email", leuser.Email, "uri", meta.RegistrationURI)
	case errors.Is(err, fs.ErrNotExist):
		newKey, err := newAccountKey()
		if err != nil {
			return fmt.Errorf("generate ACME account key: %w", err)
		}
		leuser = MyUser{Email: config.Cfg.CertControl.Email, key: newKey}
	default:
		return fmt.Errorf("load ACME account: %w", err)
	}
	return nil
}

func newChallengeProvider() (challenge.Provider, error) {
	switch strings.ToLower(config.Cfg.CertControl.DomainProvider) {
	case models.AliYun:
		cfg := alidns.NewDefaultConfig()
		cfg.APIKey = config.Cfg.CertControl.Key
		cfg.SecretKey = config.Cfg.CertControl.Secret
		provider, err := alidns.NewDNSProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("create alidns provider: %w", err)
		}
		return provider, nil
	case models.TencentCloud:
		cfg := tencentcloud.NewDefaultConfig()
		cfg.SecretID = config.Cfg.CertControl.Key
		cfg.SecretKey = config.Cfg.CertControl.Secret
		provider, err := tencentcloud.NewDNSProviderConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("create tencentcloud provider: %w", err)
		}
		return provider, nil
	default:
		return nil, errors.New("unsupported domain provider: " + config.Cfg.CertControl.DomainProvider)
	}
}

func StartContext(ctx context.Context) {
	retryDelay := renewalRetryInitial

	for {
		if !needRenew(config.Cfg.TLS.Cert) {
			retryDelay = renewalRetryInitial
			if !sleepContext(ctx, renewalCheckInterval) {
				return
			}
			continue
		}

		if err := ensureClient(); err != nil {
			slog.Error("Failed to initialize certificate renewal", "err", err)
			if !sleepContext(ctx, retryDelay) {
				return
			}
			retryDelay = nextRetryDelay(retryDelay)
			continue
		}

		if err := applyTLSCert(config.Cfg.TLS.Key, config.Cfg.TLS.Cert); err != nil {
			if !sleepContext(ctx, retryDelay) {
				return
			}
			retryDelay = nextRetryDelay(retryDelay)
			continue
		}

		retryDelay = renewalRetryInitial
		if err := updateCert(); err != nil {
			slog.Error("Failed to reload renewed certificate", "err", err)
		}
		if !sleepContext(ctx, renewalCheckInterval) {
			return
		}
	}
}

func ensureClient() error {
	if client != nil {
		return nil
	}
	return Init()
}

func nextRetryDelay(current time.Duration) time.Duration {
	next := current * 2
	if next > renewalRetryMax {
		return renewalRetryMax
	}
	return next
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func needRenew(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read cert file ", "err: ", err)
		return true
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		slog.Error("invalied PEM cert")
		return true
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		slog.Error("Failed to parse certficate ", "err: ", err)
		return true
	}

	expiry := cert.NotAfter
	remaining := time.Until(expiry)

	return remaining < 14*24*time.Hour
}

func applyTLSCert(keyPath, crtPath string) error {
	if client == nil {
		return errors.New("lego client is not initialized, call Init() first")
	}
	request := certificate.ObtainRequest{
		Domains: []string{config.Cfg.CertControl.Domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		slog.Error("Failed to apply cert, ", "err: ", err)
		return err
	}

	if err = os.WriteFile(keyPath, certificates.PrivateKey, 0600); err != nil {
		slog.Error("Failed to write private_key, ", "err: ", err)
		return err
	}
	if err = os.WriteFile(crtPath, certificates.Certificate, 0644); err != nil {
		slog.Error("Failed to write cert, ", "err: ", err)
		return err
	}
	return err
}
