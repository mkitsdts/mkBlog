//go:build lego_integration

// 这个文件会真的访问 ACME 服务器并消耗真实配额，所以默认不参与编译。
// 运行方式：
//
//	LEGO_INTEGRATION=1 go test -tags lego_integration ./pkg/tls_cert/ -v
//
// 默认指向 Let's Encrypt staging，避免把生产环境的“新账号 / 重复证书”配额烧光
// （那会让本机 IP 在一段时间内被拒绝服务）。
package tlscert

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mkBlog/config"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/registration"
)

// integrationClient 使用 staging 环境注册一个临时账号。
func integrationClient(t *testing.T) {
	t.Helper()

	if os.Getenv("LEGO_INTEGRATION") != "1" {
		t.Skip("set LEGO_INTEGRATION=1 to run the ACME integration test")
	}

	dataDir := filepath.Join("..", "..", "data")
	if _, err := os.Stat(filepath.Join(dataDir, "config.yaml")); err != nil {
		t.Skipf("no config file under %s, skipping ACME integration test", dataDir)
	}
	config.Cfg.Server.DataPath = dataDir
	config.Init()
	config.Finalize()

	if config.Cfg.CertControl.Domain == "" || config.Cfg.CertControl.Key == "" || config.Cfg.CertControl.Secret == "" {
		t.Skip("cert_control is not configured, skipping ACME integration test")
	}

	key, err := newAccountKey()
	if err != nil {
		t.Fatalf("generate account key: %v", err)
	}
	leuser = MyUser{Email: config.Cfg.CertControl.Email, key: key}

	newconfig := lego.NewConfig(&leuser)
	newconfig.Certificate.KeyType = certcrypto.RSA2048
	newconfig.Certificate.Timeout = certificateTimeout
	newconfig.CADirURL = lego.LEDirectoryStaging

	newClient, err := lego.NewClient(newconfig)
	if err != nil {
		t.Fatalf("create ACME client: %v", err)
	}

	cfg := alidns.NewDefaultConfig()
	cfg.APIKey = config.Cfg.CertControl.Key
	cfg.SecretKey = config.Cfg.CertControl.Secret
	provider, err := alidns.NewDNSProviderConfig(cfg)
	if err != nil {
		t.Fatalf("create alidns provider: %v", err)
	}
	if err := newClient.Challenge.SetDNS01Provider(provider); err != nil {
		t.Fatalf("set dns01 provider: %v", err)
	}

	reg, err := newClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		t.Fatalf("register ACME account: %v", err)
	}
	leuser.Registration = reg

	client = newClient
}

func TestIntegrationObtainCertificate(t *testing.T) {
	integrationClient(t)

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "server.key")
	certPath := filepath.Join(dir, "server.crt")

	if err := applyTLSCert(keyPath, certPath); err != nil {
		t.Fatalf("obtain certificate: %v", err)
	}

	info, err := inspectTLSCert(certPath, keyPath, time.Now())
	if err != nil {
		t.Fatalf("inspect certificate: %v", err)
	}
	if info.State != CertValid {
		t.Fatalf("expected a valid certificate, got %s", info.State)
	}
	t.Logf("certificate obtained for %v, not after %s", info.Domains, info.NotAfter)
}
