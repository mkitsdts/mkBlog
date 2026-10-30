package tlscert

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mkBlog/config"
	"mkBlog/models"
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
	certificateTimeout   = 5 * time.Minute
)

// 初始化 ACME 客户端
func Init() error {
	if err := Validate(); err != nil {
		return err
	}
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

// 准备乐此加密的账号，没有账号时生成新私钥。
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
	state := loadAttemptState()

	for {
		// 退避状态是落盘的：即使进程被重启过，也不会重新“立刻再申请一次”。
		if wait := state.wait(time.Now()); wait > 0 {
			slog.Warn("certificate issuance is backing off after a previous failure",
				"retry_in", wait.Round(time.Second),
				"consecutive_failures", state.ConsecutiveFails,
				"last_attempt_at", state.LastAttemptAt,
				"last_error", state.LastError)
			if !sleepContext(ctx, wait) {
				return
			}
		}

		now := time.Now()
		info, err := inspectTLSCert(config.Cfg.TLS.Cert, config.Cfg.TLS.Key, now)
		if err != nil {
			// 证书文件读不了 / 解析不了 / 与私钥不匹配：重试签发没有意义，退避等待人工处理
			state = state.failed(err, now).save()
			slog.Error("TLS certificate is unusable, refusing to request a new one automatically",
				"err", err,
				"hint", "fix the certificate/key files, or delete both of them to request a new pair",
				"state_file", statePath())
			continue
		}

		if info.State == CertValid {
			state = state.healthy(info).save()
			slog.Info("TLS certificate is valid, no renewal needed",
				"cert", config.Cfg.TLS.Cert, "not_after", info.NotAfter, "domains", info.Domains)
			if !sleepContext(ctx, renewalCheckInterval) {
				return
			}
			continue
		}

		slog.Info("TLS certificate needs to be issued",
			"state", info.State.String(), "cert", config.Cfg.TLS.Cert, "not_after", info.NotAfter)

		if err := ensureClient(); err != nil {
			// 配置错了重试再多次也没用，直接 panic 让用户立刻看到问题。
			var configErr *ConfigError
			if errors.As(err, &configErr) {
				slog.Error("cert_control is misconfigured, fix the configuration and restart", "err", err)
				panic(configErr)
			}
			state = state.failed(err, time.Now()).save()
			slog.Error("failed to initialize certificate renewal", "err", err)
			continue
		}

		// 先落盘尝试记录再发起申请：进程中途被 kill 也能保留现场并继续退避。
		state = state.attempted(info, time.Now()).save()
		if err := obtainCertificate(config.Cfg.TLS.Key, config.Cfg.TLS.Cert); err != nil {
			state = state.failed(err, time.Now()).save()
			slog.Error("failed to obtain TLS certificate",
				"err", err, "consecutive_failures", state.ConsecutiveFails, "state_file", statePath())
			continue
		}

		if err := updateCert(); err != nil {
			slog.Error("failed to reload renewed certificate", "err", err)
		}

		issued, inspectErr := inspectTLSCert(config.Cfg.TLS.Cert, config.Cfg.TLS.Key, time.Now())
		if inspectErr != nil {
			slog.Error("issued certificate is not usable, check the files on disk", "err", inspectErr)
			issued = certInfo{}
		}
		state = state.succeeded(issued, time.Now()).save()
		slog.Info("TLS certificate issued",
			"not_after", issued.NotAfter, "domains", issued.Domains, "state_file", statePath())

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

// retryDelay 是失败后的指数退避：0 次失败从 renewalRetryInitial 开始，最多 renewalRetryMax。
func retryDelay(failures int) time.Duration {
	delay := renewalRetryInitial
	for i := 0; i < failures && delay < renewalRetryMax; i++ {
		delay *= 2
	}
	if delay > renewalRetryMax {
		return renewalRetryMax
	}
	return delay
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

// obtainCertificate 向 ACME 服务器申请证书，并原子写入私钥与证书链。
func obtainCertificate(keyPath, certPath string) error {
	if client == nil {
		return errors.New("lego client is not initialized, call Init() first")
	}

	request := certificate.ObtainRequest{
		Domains: []string{config.Cfg.CertControl.Domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("obtain certificate for %s: %w", config.Cfg.CertControl.Domain, err)
	}

	// 写临时文件再 rename，避免进程崩溃时留下半截文件把下次启动搞挂。
	if err := writeFileAtomic(keyPath, certificates.PrivateKey, 0600); err != nil {
		return fmt.Errorf("write private key %s: %w", keyPath, err)
	}
	if err := writeFileAtomic(certPath, certificates.Certificate, 0644); err != nil {
		return fmt.Errorf("write certificate %s: %w", certPath, err)
	}
	return nil
}
