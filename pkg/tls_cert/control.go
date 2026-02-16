package tlscert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
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

func Init() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		slog.Error("GenerateKey failed", " : ", err)
		return
	}
	leuser = MyUser{
		Email: config.Cfg.CertControl.Email,
		key:   privateKey,
	}
	newconfig := lego.NewConfig(&leuser)
	newconfig.Certificate.KeyType = certcrypto.RSA2048
	client, err = lego.NewClient(newconfig)
	if err != nil {
		slog.Error("Failed to create lego client", "err :", err)
	}
	switch strings.ToLower(config.Cfg.CertControl.DomainProvider) {
	case models.AliYun:
		cfg := alidns.NewDefaultConfig()
		cfg.APIKey = config.Cfg.CertControl.Key
		cfg.SecretKey = config.Cfg.CertControl.Secret
		if p, err = alidns.NewDNSProviderConfig(cfg); err != nil {
			slog.Error("Failed to create dns provider config", "err : ", err)
			return
		}
	case models.TencentCloud:
		cfg := tencentcloud.NewDefaultConfig()
		cfg.SecretID = config.Cfg.CertControl.Key
		cfg.SecretKey = config.Cfg.CertControl.Secret
		if p, err = tencentcloud.NewDNSProviderConfig(cfg); err != nil {
			slog.Error("Failed to create dns provider config", "err : ", err)
			return
		}
	default:
		slog.Warn("Not implement dns provider.")
		return
	}
	if err := client.Challenge.SetDNS01Provider(p); err != nil {
		slog.Error("Failed to set dns to provider", "err : ", err)
		return
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		slog.Error("Failed to register ", "err : ", err)
	}
	leuser.Registration = reg
}

func Start() {
	for {
		if checkExpireDate(config.Cfg.TLS.Cert) {
			if err := applyTLSCert(config.Cfg.TLS.Key, config.Cfg.TLS.Cert); err != nil {
				time.Sleep(30 * time.Minute)
				continue
			}
			updateCert()
		}
		time.Sleep(12 * time.Hour)
	}
}

func checkExpireDate(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read cert file ", "err: ", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		slog.Error("invalied PEM cert")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		slog.Error("Failed to parse certficate ", "err: ", err)
	}

	expiry := cert.NotAfter
	remaining := time.Until(expiry)

	return remaining.Abs() < 14*24*time.Hour
}

func applyTLSCert(keyPath, crtPath string) error {
	request := certificate.ObtainRequest{
		Domains: []string{config.Cfg.CertControl.Domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		slog.Error("Failed to apply cert, ", "err: ", err)
		return err
	}

	if err = os.WriteFile(keyPath, certificates.PrivateKey, os.ModePerm); err != nil {
		slog.Error("Failed to write private_key, ", "err: ", err)
		return err
	}
	if err = os.WriteFile(crtPath, certificates.Certificate, os.ModePerm); err != nil {
		slog.Error("Failed to write cert, ", "err: ", err)
		return err
	}
	return err
}
