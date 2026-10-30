package tlscert

import (
	"fmt"
	"strings"

	"mkBlog/config"
	"mkBlog/models"
)

// cert_control 的配置本身有问题时的错误
type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("cert_control.%s: %s", e.Field, e.Reason)
}

// 校验自动证书控制所需的配置
func Validate() error {
	if !config.Cfg.CertControl.Enabled {
		return nil
	}

	if strings.TrimSpace(config.Cfg.CertControl.Domain) == "" {
		return &ConfigError{Field: "domain", Reason: "the domain to obtain certificates for is required"}
	}
	if strings.TrimSpace(config.Cfg.CertControl.Email) == "" {
		return &ConfigError{Field: "email", Reason: "an account email is required, otherwise Let's Encrypt cannot send expiry notices"}
	}
	if strings.TrimSpace(config.Cfg.CertControl.Key) == "" || strings.TrimSpace(config.Cfg.CertControl.Secret) == "" {
		return &ConfigError{Field: "key/secret", Reason: "the DNS provider credentials are required"}
	}
	if strings.TrimSpace(config.Cfg.TLS.Cert) == "" || strings.TrimSpace(config.Cfg.TLS.Key) == "" {
		return &ConfigError{Field: "tls.cert/tls.key", Reason: "the certificate and private key paths are required"}
	}

	switch strings.ToLower(config.Cfg.CertControl.DomainProvider) {
	case models.AliYun, models.TencentCloud:
		return nil
	default:
		return &ConfigError{
			Field:  "domain_provider",
			Reason: fmt.Sprintf("unsupported provider %q, want %q or %q", config.Cfg.CertControl.DomainProvider, models.AliYun, models.TencentCloud),
		}
	}
}
