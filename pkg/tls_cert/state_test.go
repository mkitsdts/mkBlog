package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mkBlog/config"
)

// writeTestCert 生成一对自签名证书与私钥，certNotAfter 控制证书有效期。
func writeTestCert(t *testing.T, dir string, certNotAfter time.Time) (certPath, keyPath string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.com"},
		DNSNames:     []string{"example.com"},
		NotBefore:    certNotAfter.Add(-90 * 24 * time.Hour),
		NotAfter:     certNotAfter,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	certPath = filepath.Join(dir, "server.crt")
	keyPath = filepath.Join(dir, "server.key")
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0644); err != nil {
		t.Fatalf("write certificate: %v", err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	return certPath, keyPath
}

func TestInspectTLSCert(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		prepare func(t *testing.T, dir string) (string, string)
		want    certState
		wantErr bool
	}{
		{
			name: "valid certificate",
			prepare: func(t *testing.T, dir string) (string, string) {
				return writeTestCert(t, dir, now.Add(60*24*time.Hour))
			},
			want: CertValid,
		},
		{
			name: "certificate within renewal window",
			prepare: func(t *testing.T, dir string) (string, string) {
				return writeTestCert(t, dir, now.Add(7*24*time.Hour))
			},
			want: CertExpiring,
		},
		{
			name: "certificate expires exactly at the window boundary",
			prepare: func(t *testing.T, dir string) (string, string) {
				return writeTestCert(t, dir, now.Add(renewBefore))
			},
			want: CertValid,
		},
		{
			name: "certificate expires right inside the window",
			prepare: func(t *testing.T, dir string) (string, string) {
				return writeTestCert(t, dir, now.Add(renewBefore-time.Minute))
			},
			want: CertExpiring,
		},
		{
			name: "expired certificate still needs renewal",
			prepare: func(t *testing.T, dir string) (string, string) {
				return writeTestCert(t, dir, now.Add(-24*time.Hour))
			},
			want: CertExpiring,
		},
		{
			name: "missing certificate file",
			prepare: func(t *testing.T, dir string) (string, string) {
				return filepath.Join(dir, "server.crt"), filepath.Join(dir, "server.key")
			},
			want: CertMissing,
		},
		{
			name: "missing private key",
			prepare: func(t *testing.T, dir string) (string, string) {
				certPath, _ := writeTestCert(t, dir, now.Add(60*24*time.Hour))
				return certPath, filepath.Join(dir, "nope.key")
			},
			wantErr: true,
		},
		{
			name: "broken certificate",
			prepare: func(t *testing.T, dir string) (string, string) {
				certPath, keyPath := writeTestCert(t, dir, now.Add(60*24*time.Hour))
				if err := os.WriteFile(certPath, []byte("not a certificate"), 0644); err != nil {
					t.Fatalf("write certificate: %v", err)
				}
				return certPath, keyPath
			},
			wantErr: true,
		},
		{
			name: "certificate does not match private key",
			prepare: func(t *testing.T, dir string) (string, string) {
				certPath, _ := writeTestCert(t, dir, now.Add(60*24*time.Hour))
				_, keyPath := writeTestCert(t, t.TempDir(), now.Add(60*24*time.Hour))
				return certPath, keyPath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			certPath, keyPath := tt.prepare(t, dir)

			info, err := inspectTLSCert(certPath, keyPath, now)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got state %s", info.State)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.State != tt.want {
				t.Fatalf("state = %s, want %s", info.State, tt.want)
			}
			if tt.want == CertMissing {
				return
			}
			if want := "example.com"; len(info.Domains) != 1 || info.Domains[0] != want {
				t.Fatalf("domains = %v, want [%s]", info.Domains, want)
			}
		})
	}
}

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		failures int
		want     time.Duration
	}{
		{failures: 0, want: renewalRetryInitial},
		{failures: 1, want: renewalRetryInitial * 2},
		{failures: 2, want: renewalRetryMax},
		{failures: 3, want: renewalRetryMax},
		{failures: 100, want: renewalRetryMax},
	}

	for _, tt := range tests {
		if got := retryDelay(tt.failures); got != tt.want {
			t.Errorf("retryDelay(%d) = %s, want %s", tt.failures, got, tt.want)
		}
	}
}

func TestAttemptStateWait(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		state attemptState
		want  time.Duration
	}{
		{
			name:  "never attempted",
			state: attemptState{},
			want:  0,
		},
		{
			name:  "previous attempt is old enough",
			state: attemptState{LastAttemptAt: now.Add(-renewalRetryInitial - time.Minute)},
			want:  0,
		},
		{
			name:  "previous attempt still in backoff",
			state: attemptState{LastAttemptAt: now.Add(-time.Hour)},
			want:  renewalRetryInitial - time.Hour,
		},
		{
			name:  "failures double the backoff",
			state: attemptState{LastAttemptAt: now.Add(-time.Hour), ConsecutiveFails: 1},
			want:  renewalRetryInitial*2 - time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.wait(now); got != tt.want {
				t.Fatalf("wait = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAttemptStateTransitions(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	info := certInfo{State: CertExpiring, NotAfter: now.Add(24 * time.Hour)}

	state := attemptState{}.attempted(info, now)
	if !state.LastAttemptAt.Equal(now) || state.LastReason != "expiring" {
		t.Fatalf("attempted state = %+v", state)
	}

	failed := state.failed(errTest, now.Add(time.Minute))
	if failed.ConsecutiveFails != 1 || failed.LastError == "" {
		t.Fatalf("failed state = %+v", failed)
	}

	issued := failed.succeeded(certInfo{NotAfter: now.Add(90 * 24 * time.Hour)}, now.Add(2*time.Minute))
	if issued.ConsecutiveFails != 0 || issued.LastError != "" || !issued.LastSuccessAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("succeeded state = %+v", issued)
	}

	healthy := failed.healthy(info)
	if healthy.ConsecutiveFails != 0 || healthy.LastError != "" {
		t.Fatalf("healthy state = %+v", healthy)
	}
}

func TestAttemptStatePersistence(t *testing.T) {
	dataDir := t.TempDir()
	originalDataPath := config.Cfg.Server.DataPath
	config.Cfg.Server.DataPath = dataDir
	t.Cleanup(func() { config.Cfg.Server.DataPath = originalDataPath })

	if loaded := loadAttemptState(); !loaded.LastAttemptAt.IsZero() {
		t.Fatalf("expected an empty state before saving, got %+v", loaded)
	}

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	saved := attemptState{}.failed(errTest, now).save()

	loaded := loadAttemptState()
	if loaded.ConsecutiveFails != saved.ConsecutiveFails || loaded.LastError != saved.LastError {
		t.Fatalf("loaded state = %+v, want %+v", loaded, saved)
	}
	if !loaded.LastAttemptAt.Equal(now) {
		t.Fatalf("last attempt = %s, want %s", loaded.LastAttemptAt, now)
	}
}

var errTest = errors.New("boom")
