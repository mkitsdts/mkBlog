package tlscert

import (
	"errors"
	"testing"

	"mkBlog/config"
)

func TestValidate(t *testing.T) {
	valid := config.TLSCertAutoControlConfig{
		Enabled:        true,
		Email:          "ops@example.com",
		Domain:         "example.com",
		Key:            "key",
		Secret:         "secret",
		DomainProvider: "Aliyun",
	}

	tests := []struct {
		name       string
		mutate     func(cfg *config.TLSCertAutoControlConfig)
		disabled   bool
		wantField  string
		wantNoErrs bool
	}{
		{name: "valid configuration", wantNoErrs: true},
		{
			name:       "cert_control disabled skips validation",
			disabled:   true,
			mutate:     func(cfg *config.TLSCertAutoControlConfig) { cfg.Domain = "" },
			wantNoErrs: true,
		},
		{name: "missing domain", mutate: func(cfg *config.TLSCertAutoControlConfig) { cfg.Domain = " " }, wantField: "domain"},
		{name: "missing email", mutate: func(cfg *config.TLSCertAutoControlConfig) { cfg.Email = "" }, wantField: "email"},
		{name: "missing credentials", mutate: func(cfg *config.TLSCertAutoControlConfig) { cfg.Secret = "" }, wantField: "key/secret"},
		{name: "unsupported provider", mutate: func(cfg *config.TLSCertAutoControlConfig) { cfg.DomainProvider = "cloudflare" }, wantField: "domain_provider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCertControl := config.Cfg.CertControl
			originalTLS := config.Cfg.TLS
			t.Cleanup(func() {
				config.Cfg.CertControl = originalCertControl
				config.Cfg.TLS = originalTLS
			})

			config.Cfg.CertControl = valid
			config.Cfg.TLS = config.TLSConfig{Cert: "data/server.crt", Key: "data/server.key"}
			if tt.disabled {
				config.Cfg.CertControl.Enabled = false
			}
			if tt.mutate != nil {
				tt.mutate(&config.Cfg.CertControl)
			}

			err := Validate()
			if tt.wantNoErrs {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			var configErr *ConfigError
			if !errors.As(err, &configErr) {
				t.Fatalf("expected a *ConfigError, got %v", err)
			}
			if configErr.Field != tt.wantField {
				t.Fatalf("field = %q, want %q", configErr.Field, tt.wantField)
			}
		})
	}
}
