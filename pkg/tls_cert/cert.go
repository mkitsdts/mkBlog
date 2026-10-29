package tlscert

import (
	"crypto/tls"
	"errors"
	"log/slog"
	"mkBlog/config"
	"sync"
	"time"
)

// certNotLoadedWarnInterval 限制"证书还没准备好"日志的频率，
// 否则每一个 TLS 握手都会刷一条。
const certNotLoadedWarnInterval = time.Minute

var (
	currentCert *tls.Certificate
	certMux     sync.RWMutex

	lastWarnMux sync.Mutex
	lastWarnAt  time.Time
)

// LoadCert 从磁盘加载证书到内存。
func LoadCert() error {
	newCert, err := tls.LoadX509KeyPair(config.Cfg.TLS.Cert, config.Cfg.TLS.Key)
	if err != nil {
		return err
	}
	certMux.Lock()
	currentCert = &newCert
	certMux.Unlock()
	return nil
}

// GetCurrentCert 是 tls.Config 的回调。证书还没准备好时返回错误，
// 让握手失败但进程继续运行（首次部署时证书要等 cert_control 签发）。
func GetCurrentCert(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certMux.RLock()
	cert := currentCert
	certMux.RUnlock()
	if cert == nil {
		warnCertNotLoaded()
		return nil, errors.New("TLS certificate is not loaded yet")
	}
	return cert, nil
}

func warnCertNotLoaded() {
	lastWarnMux.Lock()
	defer lastWarnMux.Unlock()
	if time.Since(lastWarnAt) < certNotLoadedWarnInterval {
		return
	}
	lastWarnAt = time.Now()
	slog.Warn("TLS certificate is not loaded yet, handshakes will fail until it is issued",
		"cert", config.Cfg.TLS.Cert, "key", config.Cfg.TLS.Key)
}

// updateCert 在证书续期/签发成功后重新加载。
func updateCert() error {
	return LoadCert()
}
