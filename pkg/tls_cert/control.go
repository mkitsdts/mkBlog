package tlscert

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
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
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

var client *lego.Client
var leuser MyUser
var p challenge.Provider

const (
	renewalCheckInterval = 12 * time.Hour
	renewalRetryInitial  = 6 * time.Hour
	renewalRetryMax      = 24 * time.Hour
)

func Init() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	leuser = MyUser{
		Email: config.Cfg.CertControl.Email,
		key:   privateKey,
	}
	newconfig := lego.NewConfig(&leuser)
	newconfig.Certificate.KeyType = certcrypto.RSA2048
	client, err = lego.NewClient(newconfig)
	if err != nil {
		return err
	}
	switch strings.ToLower(config.Cfg.CertControl.DomainProvider) {
	case models.AliYun:
		cfg := alidns.NewDefaultConfig()
		cfg.APIKey = config.Cfg.CertControl.Key
		cfg.SecretKey = config.Cfg.CertControl.Secret
		if p, err = alidns.NewDNSProviderConfig(cfg); err != nil {
			return err
		}
	case models.TencentCloud:
		cfg := tencentcloud.NewDefaultConfig()
		cfg.SecretID = config.Cfg.CertControl.Key
		cfg.SecretKey = config.Cfg.CertControl.Secret
		if p, err = tencentcloud.NewDNSProviderConfig(cfg); err != nil {
			return err
		}
	default:
		return errors.New("unsupported domain provider: " + config.Cfg.CertControl.DomainProvider)
	}
	if err := client.Challenge.SetDNS01Provider(p); err != nil {
		return err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	leuser.Registration = reg
	return nil
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
