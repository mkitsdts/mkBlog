package config

import (
	"path/filepath"
	"testing"
)

func TestFinalizeCertificatePaths(t *testing.T) {
	tests := []struct {
		name         string
		tlsEnabled   bool
		certControl  bool
		cert, key    string
		wantCertPath string
		wantKeyPath  string
	}{
		{
			name:         "tls enabled anchors paths to the data directory",
			tlsEnabled:   true,
			cert:         "server.crt",
			key:          "server.key",
			wantCertPath: filepath.Join("data", "server.crt"),
			wantKeyPath:  filepath.Join("data", "server.key"),
		},
		{
			name:         "cert control alone anchors paths to the data directory",
			certControl:  true,
			cert:         "server.crt",
			key:          "server.key",
			wantCertPath: filepath.Join("data", "server.crt"),
			wantKeyPath:  filepath.Join("data", "server.key"),
		},
		{
			name:         "disabled features keep the configured paths",
			cert:         "server.crt",
			key:          "server.key",
			wantCertPath: "server.crt",
			wantKeyPath:  "server.key",
		},
		{
			name:         "absolute paths are kept as is",
			tlsEnabled:   true,
			cert:         "/etc/mkblog/server.crt",
			key:          "/etc/mkblog/server.key",
			wantCertPath: "/etc/mkblog/server.crt",
			wantKeyPath:  "/etc/mkblog/server.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalTLS := Cfg.TLS
			originalCertControl := Cfg.CertControl
			originalServer := Cfg.Server
			t.Cleanup(func() {
				Cfg.TLS = originalTLS
				Cfg.CertControl = originalCertControl
				Cfg.Server = originalServer
			})

			Cfg.Server.DataPath = "data"
			Cfg.Server.ImageSavePath = "img"
			Cfg.TLS = TLSConfig{Enabled: tt.tlsEnabled, Cert: tt.cert, Key: tt.key}
			Cfg.CertControl = TLSCertAutoControlConfig{Enabled: tt.certControl}

			Finalize()

			if Cfg.TLS.Cert != tt.wantCertPath {
				t.Errorf("cert = %q, want %q", Cfg.TLS.Cert, tt.wantCertPath)
			}
			if Cfg.TLS.Key != tt.wantKeyPath {
				t.Errorf("key = %q, want %q", Cfg.TLS.Key, tt.wantKeyPath)
			}
		})
	}
}
